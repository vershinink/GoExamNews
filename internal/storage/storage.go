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
	ErrEmptyId     = errors.New("empty id")
)

// Post - структура поста из RSS ленты для работы с БД.
type Post struct {
	ID      string    `json:"id" bson:"_id"`
	Title   string    `json:"title" bson:"title"`
	Content string    `json:"content" bson:"content"`
	PubTime time.Time `json:"pubTime" bson:"pubTime"`
	Link    string    `json:"link" bson:"link"`
}

// TextSearch - структура запроса для текстового поиска в БД
// по заголовкам постов.
type Options struct {
	// SearchQuery - запрос для текстового поиска.
	SearchQuery string

	// Count - максимальное число возвращаемых постов.
	Count int

	// Offset - число постов на сдвиг в пагинации.
	Offset int
}

// Interface - интерфейс хранилища постов из RSS лент.
//
//go:generate go run github.com/vektra/mockery/v2@v2.44.1 --name=DB
type DB interface {
	AddPosts(ctx context.Context, posts <-chan Post) (int, error)
	Posts(ctx context.Context, op ...*Options) ([]Post, error)
	Count(ctx context.Context, q ...*Options) (int64, error)
	PostById(ctx context.Context, id string) (Post, error)
	Close() error
}
