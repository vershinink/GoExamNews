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
	dbName  string = "goNews"
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
func setTestOpts(path string) *options.ClientOptions {
	return options.Client().ApplyURI(path)
}

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

	// Создаем уникальный индекс по полю title, чтобы избежать
	// записи уже существующих постов.
	collection := db.Database(dbName).Collection(colName)
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "title", Value: -1}},
		Options: options.Index().SetUnique(true),
	}
	_, err = collection.Indexes().CreateOne(tm, indexModel)
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
	const operation = "storage.mongodb.Posts"

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

// Posts возвращает указанное число последних постов по дате
// публикации из БД.
func (s *Storage) Posts(ctx context.Context, n int) ([]storage.Post, error) {
	const operation = "storage.mongodb.Posts"

	if n == 0 {
		return nil, fmt.Errorf("%s: %w", operation, storage.ErrZeroRequest)
	}

	opts := options.Find().SetSort(bson.D{{Key: "pubTime", Value: -1}}).SetLimit(int64(n))
	filter := bson.D{}

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
