// Пакет для работы с базой данных MongoDB.
package mongodb

import (
	"GoNews/internal/config"
	"GoNews/internal/storage"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Название базы и коллекции в БД. Используются переменные,
// а не константы, так как в тестах им присваиваются другие
// значения.
var (
	dbName  string = "goExam"
	colName string = "posts"
)

const tmConn time.Duration = time.Second * 20

// Storage - пул подключений к БД.
type Storage struct {
	db *mongo.Client
}

// New - обертка для конструктора пула подключений new.
func New(cfg *config.Config) *Storage {
	opts := setOpts(cfg.StoragePath, cfg.StorageUser, cfg.StoragePasswd)
	storage, err := new(opts)
	if err != nil {
		log.Fatalf("failed to init storage: %s", err.Error())
	}
	return storage
}

// setOpts настраивает опции нового подключения к БД.
// Функция вынесена отдельно для подмены ее в пакете
// с тестами.
func setOpts(path, user, password string) *options.ClientOptions {
	credential := options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		AuthSource:    "admin",
		Username:      user,
		Password:      password,
	}
	opts := options.Client().ApplyURI(path).SetAuth(credential)
	return opts
}

// setTestOpts возвращает опции нового подключения без авторизации.
// func setTestOpts(path string) *options.ClientOptions {
// 	return options.Client().ApplyURI(path)
// }

// new - конструктор пула подключений к БД.
func new(opts *options.ClientOptions) (*Storage, error) {
	const operation = "storage.mongodb.new"

	tm, cancel := context.WithTimeout(context.Background(), tmConn)
	defer cancel()

	db, err := mongo.Connect(tm, opts)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	err = db.Ping(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	collection := db.Database(dbName).Collection(colName)

	// Создаем уникальный индекс по полю title, чтобы избежать
	// записи уже существующих постов.
	indexUniq := mongo.IndexModel{
		Keys:    bson.D{{Key: "title", Value: -1}},
		Options: options.Index().SetUnique(true),
	}
	// Создаем индекс текстового поиска по полю title.
	indexText := mongo.IndexModel{
		Keys: bson.D{{Key: "title", Value: "text"}},
	}
	_, err = collection.Indexes().CreateMany(tm, []mongo.IndexModel{indexUniq, indexText})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return &Storage{db: db}, nil
}

// Close - обертка для закрытия пула подключений.
func (s *Storage) Close() error {
	return s.db.Disconnect(context.Background())
}

// AddPosts читает посты из переданного канала и записывает
// их в БД. Возвращает количество успешно записанных постов
// и ошибку, отличную от duplicate key error.
func (s *Storage) AddPosts(ctx context.Context, posts <-chan storage.Post) (int, error) {
	const operation = "storage.mongodb.AddPosts"

	var input []interface{}
	for p := range posts {
		bsn := bson.D{
			{Key: "_id", Value: primitive.NewObjectID()},
			{Key: "title", Value: p.Title},
			{Key: "content", Value: p.Content},
			{Key: "pubTime", Value: primitive.NewDateTimeFromTime(p.PubTime)},
			{Key: "link", Value: p.Link},
		}
		input = append(input, bsn)
	}

	collection := s.db.Database(dbName).Collection(colName)
	opts := options.InsertMany().SetOrdered(false)
	res, err := collection.InsertMany(ctx, input, opts)
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		return len(res.InsertedIDs), fmt.Errorf("%s: %w", operation, err)
	}

	return len(res.InsertedIDs), nil
}

// Posts возвращает посты из БД в соответствии с переданными опциями.
// Опции включают в себя лимит числа постов, оффсет для пагинации и
// запрос на текстовый поиск в заголовках.
// Если параметр опции nil, то вернет все посты, отсортированные по
// дате публикации.
func (s *Storage) Posts(ctx context.Context, op ...*storage.Options) ([]storage.Post, error) {
	const operation = "storage.mongodb.Posts"

	filter := bson.D{}
	sort := bson.D{{Key: "pubTime", Value: -1}}
	opts := options.Find()

	var query string
	var lim, off int64
	if op[0] != nil {
		query = op[0].SearchQuery
		lim = int64(op[0].Count)
		off = int64(op[0].Offset)
	}

	if query != "" {
		filter = bson.D{{Key: "$text", Value: bson.D{{Key: "$search", Value: query}}}}
		sort = bson.D{{Key: "score", Value: bson.D{{Key: "$meta", Value: "textScore"}}}}
	}
	opts = opts.SetSort(sort)

	if lim > 0 {
		opts = opts.SetLimit(lim)
	}

	if off > 0 {
		opts = opts.SetSkip(off)
	}

	collection := s.db.Database(dbName).Collection(colName)
	res, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	var posts []storage.Post
	err = res.All(ctx, &posts)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	if len(posts) == 0 {
		return nil, fmt.Errorf("%s: %w", operation, storage.ErrEmptyDB)
	}

	return posts, nil
}

// Count возвращает число постов, соответствующих условиям поиска.
func (s *Storage) Count(ctx context.Context, op ...*storage.Options) (int64, error) {
	const operation = "storage.mongodb.Count"

	filter := bson.D{}
	opts := options.Count().SetHint("_id_")

	if op[0] != nil && op[0].SearchQuery != "" {
		filter = bson.D{{Key: "$text", Value: bson.D{{Key: "$search", Value: op[0].SearchQuery}}}}
		opts = nil
	}

	collection := s.db.Database(dbName).Collection(colName)
	res, err := collection.CountDocuments(ctx, filter, opts)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}
	return res, nil
}

// PostById возвращает пост по переданному ID.
func (s *Storage) PostById(ctx context.Context, id string) (storage.Post, error) {
	const operation = "storage.mongodb.PostById"
	var post storage.Post

	if id == "" {
		return post, fmt.Errorf("%s: %w", operation, storage.ErrEmptyId)
	}

	obj, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return post, fmt.Errorf("%s: %w", operation, err)
	}

	collection := s.db.Database(dbName).Collection(colName)
	filter := bson.D{{Key: "_id", Value: obj}}
	res := collection.FindOne(ctx, filter)
	if res.Err() == mongo.ErrNoDocuments {
		return post, fmt.Errorf("%s: %w", operation, storage.ErrNotFound)
	}
	if res.Err() != nil {
		return post, fmt.Errorf("%s: %w", operation, res.Err())
	}

	err = res.Decode(&post)
	if err != nil {
		return post, fmt.Errorf("%s: %w", operation, err)
	}

	return post, nil
}
