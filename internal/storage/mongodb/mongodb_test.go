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
func (s *Storage) addOne(p storage.Post) (string, error) {
	bsn := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "title", Value: p.Title},
		{Key: "content", Value: p.Content},
		{Key: "pubTime", Value: primitive.NewDateTimeFromTime(time.Now())},
		{Key: "link", Value: p.Link},
	}
	collection := s.db.Database(dbName).Collection(colName)
	res, err := collection.InsertOne(context.Background(), bsn)
	if err != nil {
		return "", err
	}
	hex := res.InsertedID.(primitive.ObjectID)
	return hex.Hex(), nil
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
		_, err := st.addOne(p)
		if err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name    string
		opts    *storage.Options
		want    int
		wantErr bool
	}{
		{
			name:    "OK_All",
			opts:    nil,
			want:    3,
			wantErr: false,
		},
		{
			name:    "OK_Count_2",
			opts:    &storage.Options{Count: 2},
			want:    2,
			wantErr: false,
		},
		{
			name:    "OK_Offset_2",
			opts:    &storage.Options{Offset: 2},
			want:    1,
			wantErr: false,
		},
		{
			name:    "OK_Search_one",
			opts:    &storage.Options{SearchQuery: "one"},
			want:    2,
			wantErr: false,
		},
		{
			name:    "OK_With_search_two",
			opts:    &storage.Options{SearchQuery: "two"},
			want:    1,
			wantErr: false,
		},
		{
			name:    "No_search_results",
			opts:    &storage.Options{SearchQuery: "asdf"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "OK_Complex",
			opts:    &storage.Options{SearchQuery: "one", Count: 2, Offset: 1},
			want:    1,
			wantErr: false,
		},
		{
			name:    "OK_Empty_options",
			opts:    &storage.Options{},
			want:    3,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := st.Posts(context.Background(), tt.opts)
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

func TestStorage_Count(t *testing.T) {

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
		_, err := st.addOne(p)
		if err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name    string
		opts    *storage.Options
		want    int64
		wantErr bool
	}{
		{
			name:    "Count_All",
			opts:    nil,
			want:    3,
			wantErr: false,
		},
		{
			name:    "Count_Search_one",
			opts:    &storage.Options{SearchQuery: "one"},
			want:    2,
			wantErr: false,
		},
		{
			name:    "Count_Search_one",
			opts:    &storage.Options{SearchQuery: "asdf"},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := st.Count(context.Background(), tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.Count() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Storage.Count() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_PostById(t *testing.T) {

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

	var ids []string
	for _, p := range posts {
		id, err := st.addOne(p)
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}

	tests := []struct {
		name    string
		id      string
		want    string
		wantErr bool
	}{
		{
			name:    "OK_One",
			id:      ids[0],
			want:    "https://google.com",
			wantErr: false,
		},
		{
			name:    "Error_Empty_id",
			id:      "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Error_Incorrect_id",
			id:      "asdf",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Error_Not_found",
			id:      primitive.NewObjectID().Hex(),
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := st.PostById(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.PostById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Link != tt.want {
				t.Errorf("Storage.PostById().Link = %v, want %v", got.Link, tt.want)
			}
		})
	}
}
