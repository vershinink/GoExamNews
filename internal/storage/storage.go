// Пакет содержит основную структуру Post для работы с постом из RSS лент и интерфейс
// для работы с любой реализацией базы данных, удовлетворяющей этому интерфейсу.
package storage

import (
	"context"
	"errors"
	"time"
)

// Ошибки при работе с БД.
var (
	ErrEmptyDB     = errors.New("database is empty")
	ErrZeroRequest = errors.New("requested 0 posts")
)

// Post - структура поста из RSS ленты для работы с БД.
type Post struct {
	ID      string    `json:"id" bson:"_id"`
	Title   string    `json:"title" bson:"title"`
	Content string    `json:"content" bson:"content"`
	PubTime time.Time `json:"pubTime" bson:"pubTime"`
	Link    string    `json:"link" bson:"link"`
}

// Interface - интерфейс хранилища постов из RSS лент.
//
//go:generate go run github.com/vektra/mockery/v2@v2.44.1 --name=DB
type DB interface {
	AddPosts(ctx context.Context, posts <-chan Post) (int, error)
	Posts(ctx context.Context, n int) ([]Post, error)
	Close() error
}
