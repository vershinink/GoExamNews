// Пакет для работы с базой данных MongoDB.
package mongodb

import (
	"GoNews/internal/config"
	"GoNews/internal/storage"
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Название базы и коллекции в БД.
const (
	dbName  string = "goNews"
	colName string = "posts"
)

// Storage - пул подключений к БД.
type Storage struct {
	db *mongo.Client
}

// New - обертка для конструктора пула подключений new.
func New(cfg *config.Config) *Storage {
	storage, err := new(cfg.StoragePath, cfg.StorageUser, cfg.StoragePasswd)
	if err != nil {
		log.Fatalf("failed to init storage: %s", err.Error())
	}
	return storage
}

// new - конструктор пула подключений к БД.
func new(path, user, password string) (*Storage, error) {
	const operation = "storage.mongodb.new"

	credential := options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		AuthSource:    "admin",
		Username:      user,
		Password:      password,
	}

	// Задаем опции подключения.
	// opts := options.Client().ApplyURI(path).SetAuth(credential)
	opts := options.Client().ApplyURI(path)
	_ = credential
	// Создаем подключение к MongoDB и проверяем его.
	db, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	err = db.Ping(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	// Создаем уникальный индекс по полю title.
	collection := db.Database(dbName).Collection(colName)
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "title", Value: -1}},
		Options: options.Index().SetUnique(true),
	}
	_, err = collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return &Storage{db: db}, nil
}

// Close - обертка для закрытия пула подключений.
func (s *Storage) Close() error {
	return s.db.Disconnect(context.Background())
}

// AddPosts читает посты из переданного канала и записывает их в БД.
// Возвращает количество успешно записанных постов.
func (s *Storage) AddPosts(ctx context.Context, posts <-chan storage.Post) int {
	var input []interface{}
	for p := range posts {
		bsn := bson.D{
			{Key: "_id", Value: primitive.NewObjectID()},
			{Key: "title", Value: p.Title},
			{Key: "content", Value: p.Content},
			{Key: "pubTime", Value: p.PubTime},
			{Key: "link", Value: p.Link},
		}
		input = append(input, bsn)
	}

	collection := s.db.Database(dbName).Collection(colName)
	opts := options.InsertMany().SetOrdered(false)
	res, _ := collection.InsertMany(ctx, input, opts)

	return len(res.InsertedIDs)
}
