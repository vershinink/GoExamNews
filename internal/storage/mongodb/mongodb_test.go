// Пакет для работы с базой данных MongoDB.

package mongodb

import (
	"GoNews/internal/storage"
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Тестирование пакета mongodb требует подключения к базе данных
// MongoDB без установленной авторизации.

var path string = "mongodb://192.168.0.102:27017/"
var posts = []storage.Post{
	{
		Title:   fmt.Sprintf("Test post one %d", rand.Int()),
		Content: "Test content 1",
		Link:    "https://google.com",
		PubTime: time.Now(),
	},
	{
		Title:   fmt.Sprintf("Test post one %d", rand.Int()),
		Content: "Test content 2",
		Link:    "https://google.com",
		PubTime: time.Now(),
	},
	{
		Title:   fmt.Sprintf("Test post two %d", rand.Int()),
		Content: "Test content 3",
		Link:    "https://google.com",
		PubTime: time.Now(),
	},
}

// addOne добавляет один пост в БД. Функция для использования в тестах.
func (s *Storage) addOne(p storage.Post) error {
	bsn := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "title", Value: p.Title},
		{Key: "content", Value: p.Content},
		{Key: "pubTime", Value: primitive.NewDateTimeFromTime(time.Now())},
		{Key: "link", Value: p.Link},
	}
	collection := s.db.Database(dbName).Collection(colName)
	_, err := collection.InsertOne(context.Background(), bsn)
	if err != nil {
		return err
	}
	return nil
}
func (s *Storage) trun() error {
	collection := s.db.Database(dbName).Collection(colName)
	_, err := collection.DeleteMany(context.Background(), bson.D{})
	return err
}

func Test_new(t *testing.T) {

	// Для тестирования авторизации.
	// opts := setOpts(path, "admin", os.Getenv("DB_PASSWD"))

	opts := setTestOpts(path)
	st, err := new(opts)
	if err != nil {
		t.Fatal(err.Error())
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
		t.Fatal(err.Error())
	}
	defer st.Close()

	tests := []struct {
		name    string
		posts   <-chan storage.Post
		want    int
		wantErr bool
	}{
		{
			name:    "OK",
			posts:   ch,
			want:    3,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := st.AddPosts(context.Background(), tt.posts)
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
		t.Fatal(err.Error())
	}
	defer st.Close()

	// Очищаем коллекцию и заполняем постами заново.
	err = st.trun()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range posts {
		err := st.addOne(p)
		if err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name    string
		num     int
		opts    *storage.TextSearch
		want    int
		wantErr bool
	}{
		{
			name:    "OK_Wthout_search",
			num:     3,
			opts:    nil,
			want:    3,
			wantErr: false,
		},
		{
			name:    "OK_With_search_one",
			num:     3,
			opts:    &storage.TextSearch{Query: "one"},
			want:    2,
			wantErr: false,
		},
		{
			name:    "OK_With_search_two",
			num:     3,
			opts:    &storage.TextSearch{Query: "two"},
			want:    1,
			wantErr: false,
		},
		{
			name:    "No_search_results",
			num:     3,
			opts:    &storage.TextSearch{Query: "asdf"},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := st.Posts(context.Background(), tt.num, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.Posts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("Storage.Posts() = %v, want %v", len(got), tt.want)
			}
		})
	}
}
