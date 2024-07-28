// Пакет для эмуляции работы с базой данных.
package memdb

import (
	"GoNews/internal/storage"
	"context"
	"fmt"
	"strconv"
)

// Storage - эмуляция пула подключений к БД.
type Storage struct {
	I int // счетчик спарсенных постов
}

// New - конструктор эмулятора пулов подключений.
func New() *Storage {
	return &Storage{}
}

// Close - эмуляция закрытия пула подключений.
func (s *Storage) Close() error {
	return nil
}

// AddPost - эмуляция метода добавления постов в БД.
func (s *Storage) AddPosts(ctx context.Context, posts <-chan storage.Post) (int, error) {
	for p := range posts {
		_ = p
		s.I++
	}
	return s.I, nil
}

// Posts - эмуляция метода получения постов из БД.
func (s *Storage) Posts(ctx context.Context, n int) ([]storage.Post, error) {
	var posts []storage.Post
	for i := 1; i <= n; i++ {
		var p storage.Post
		p.ID = strconv.Itoa(i)
		p.Title = fmt.Sprintf("Post %d", i)
		p.Content = fmt.Sprintf("Content %d", i)
		posts = append(posts, p)
	}
	return posts, nil
}
