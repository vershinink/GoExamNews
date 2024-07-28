// Пакет для работы с базой данных MongoDB.

package mongodb

import (
	"GoNews/internal/storage"
	"context"
	"reflect"
	"testing"
	"time"
)

var path string = "mongodb://localhost:27017/"
var posts = []storage.Post{
	{
		Title:   "Test post 1",
		Content: "Test 111",
		Link:    "https://google.com",
		PubTime: time.Now(),
	},
	{
		Title:   "Test post 2",
		Content: "Test 222",
		Link:    "https://google.com",
		PubTime: time.Now(),
	},
}

func Test_new(t *testing.T) {
	st, err := new(path, "", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	st.Close()
}

func TestStorage_AddPosts(t *testing.T) {

	ch := make(chan storage.Post, len(posts))
	for _, p := range posts {
		ch <- p
	}
	close(ch)

	st, err := new(path, "", "")
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

	st, err := new(path, "", "")
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
		want    string
		wantErr bool
	}{
		{
			name:    "OK",
			s:       st,
			args:    args{ctx: context.Background(), n: 2},
			want:    "Test post 1",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Posts(tt.args.ctx, tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.Posts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got[0].Title, tt.want) {
				t.Errorf("Storage.Posts() = %v, want %v", got, tt.want)
			}
		})
	}
}
