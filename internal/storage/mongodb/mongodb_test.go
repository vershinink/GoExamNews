// Пакет для работы с базой данных MongoDB.

package mongodb

import (
	"GoNews/internal/storage"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// Тестирование пакета mongodb требует подключения к базе данных
// MongoDB без установленной авторизации.

var path string = "mongodb://localhost:27017/"
var posts = []storage.Post{
	{
		Title:   fmt.Sprintf("Test post %d", rand.Int()),
		Content: "Test content",
		Link:    "https://google.com",
		PubTime: time.Now(),
	},
	{
		Title:   fmt.Sprintf("Test post %d", rand.Int()),
		Content: "Test content",
		Link:    "https://google.com",
		PubTime: time.Now(),
	},
}

func Test_new(t *testing.T) {

	// Для тестирования авторизации.
	// opts := setOpts(path, "admin", os.Getenv("DB_PASSWD"))

	opts := setTestOpts(path)
	st, err := new(opts)
	if err != nil {
		t.Fatalf(err.Error())
	}
	st.Close()
}

func TestStorage_AddPosts(t *testing.T) {

	dbName = "testDB"
	colName = "testCollection"

	ch := make(chan storage.Post, len(posts))
	for _, p := range posts {
		ch <- p
	}
	close(ch)

	st, err := new(setTestOpts(path))
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer st.Close()

	type args struct {
		ctx   context.Context
		posts <-chan storage.Post
	}
	tests := []struct {
		name    string
		s       *Storage
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "OK",
			s:       st,
			args:    args{ctx: context.Background(), posts: ch},
			want:    2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.AddPosts(tt.args.ctx, tt.args.posts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.AddPosts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Storage.AddPosts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_Posts(t *testing.T) {

	dbName = "testDB"
	colName = "testCollection"

	st, err := new(setTestOpts(path))
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer st.Close()

	type args struct {
		ctx context.Context
		n   int
	}
	tests := []struct {
		name    string
		s       *Storage
		args    args
		wantErr bool
	}{
		{
			name:    "OK",
			s:       st,
			args:    args{ctx: context.Background(), n: 2},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Posts(tt.args.ctx, tt.args.n)
			if (err != nil) != tt.wantErr {
				if errors.Is(err, storage.ErrEmptyDB) {
					t.Errorf("Storage.Posts() error = %v, need to add posts", err)
					return
				}
				t.Errorf("Storage.Posts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.args.n {
				t.Errorf("Storage.Posts() = %v, want %v", len(got), tt.args.n)
			}
		})
	}
}
