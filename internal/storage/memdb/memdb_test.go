// Пакет для эмуляции работы с базой данных.

package memdb

import (
	"GoNews/internal/storage"
	"context"
	"reflect"
	"testing"
)

var posts = []storage.Post{
	{Title: "First post"},
	{Title: "Second post"},
}

func TestStorage_AddPosts(t *testing.T) {

	ch := make(chan storage.Post, len(posts))
	for _, p := range posts {
		ch <- p
	}
	close(ch)

	st := New()

	type args struct {
		ctx   context.Context
		posts <-chan storage.Post
	}
	tests := []struct {
		name string
		s    *Storage
		args args
		want int
	}{
		{
			name: "OK",
			s:    st,
			args: args{ctx: context.Background(), posts: ch},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.s.AddPosts(tt.args.ctx, tt.args.posts); got != tt.want {
				t.Errorf("Storage.AddPosts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_Posts(t *testing.T) {
	var arr = []storage.Post{
		{
			ID:      "1",
			Title:   "Post 1",
			Content: "Content 1",
		},
		{
			ID:      "2",
			Title:   "Post 2",
			Content: "Content 2",
		},
	}
	st := New()

	type args struct {
		ctx context.Context
		n   int
	}
	tests := []struct {
		name string
		s    *Storage
		args args
		want []storage.Post
	}{
		{
			name: "OK",
			s:    st,
			args: args{ctx: context.Background(), n: 2},
			want: arr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.s.Posts(tt.args.ctx, tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Storage.Posts() = %v, want %v", got, tt.want)
			}
		})
	}
}
