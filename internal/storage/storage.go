package storage

import (
	"errors"
	"time"
)

// Ошибки при работе с БД
var (
	ErrPostExists = errors.New("post already exists")
	ErrEmptyDB    = errors.New("database is empty")
)

type Post struct {
	ID      string
	Title   string
	Content string
	PubTime time.Time
	Link    string
}

type Interface interface {
}
