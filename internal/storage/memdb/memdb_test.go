// Пакет для эмуляции работы с базой данных.

package memdb

import (
	"GoNews/internal/storage"
	"context"
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
			name: "Success",
			s:    st,
			args: args{ctx: context.Background(), posts: ch},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.AddPosts(tt.args.ctx, tt.args.posts); got != tt.want {
				t.Errorf("Storage.AddPosts() = %v, want %v", got, tt.want)
			}
		})
	}
}
