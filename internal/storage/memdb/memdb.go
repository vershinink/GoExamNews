// Пакет для эмуляции работы с базой данных.
package memdb

import (
	"GoNews/internal/storage"
	"context"
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
func (s *Storage) AddPosts(ctx context.Context, posts <-chan storage.Post) int {
	for p := range posts {
		_ = p
		s.I++
	}
	return s.I
}
