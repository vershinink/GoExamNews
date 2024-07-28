// Пакет содержит основную структуру Post для работы с постом из RSS лент и интерфейс
// для работы с любой реализацией базы данных, удовлетворяющей этому интерфейсу.
package storage

import (
	"context"
	"errors"
	"time"
)

// Ошибки при работе с БД
var (
	ErrEmptyDB     = errors.New("database is empty")
	ErrZeroRequest = errors.New("requested 0 posts")
)

type Post struct {
	ID      string    `json:"id" bson:"_id"`
	Title   string    `json:"title" bson:"title"`
	Content string    `json:"content" bson:"content"`
	PubTime time.Time `json:"pubTime" bson:"pubTime"`
	Link    string    `json:"link" bson:"link"`
}

type Interface interface {
	AddPosts(ctx context.Context, posts <-chan Post) (int, error)
	Posts(ctx context.Context, n int) ([]Post, error)
	Close() error
}
