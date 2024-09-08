// Пакет для работы с сервером и обработчиками API.

package server

import (
	"GoNews/internal/logger"
	"GoNews/internal/mocks"
	"GoNews/internal/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

// posts - тестовые данные.
var posts = []storage.Post{
	{ID: "1", Title: "Title one", Content: "Content 1", PubTime: time.Now(), Link: "https://google.com"},
	{ID: "2", Title: "Title two", Content: "Content 3", PubTime: time.Now(), Link: "https://ya.ru"},
	{ID: "3", Title: "Title three", Content: "Content 3", PubTime: time.Now(), Link: "https://bing.com"},
}

func TestIndex(t *testing.T) {
	t.Parallel()
	logger.Discard()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", Index())
	srv := httptest.NewServer(mux)
	defer srv.Close()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Index() error = unexpected status code, got = %v, want = %v", rr.Code, http.StatusOK)
	}

	if rr.Body.Len() == 0 {
		t.Errorf("Index() error = empty body")
	}
}

func TestPosts(t *testing.T) {
	logger.Discard()
	t.Parallel()

	tests := []struct {
		name        string
		argumentURL string
		searchParam string
		wantURL     []string
		respError   string
		mockError   error
	}{
		{
			name:        "Posts_OK",
			argumentURL: "3",
			searchParam: "",
			wantURL:     []string{"https://google.com", "https://ya.ru", "https://bing.com"},
			respError:   "",
			mockError:   nil,
		},
		{
			name:        "Incorrect_GET_request",
			argumentURL: "asd",
			searchParam: "",
			wantURL:     nil,
			respError:   "incorrect posts number",
			mockError:   nil,
		},
		{
			name:        "DB_error",
			argumentURL: "3",
			searchParam: "",
			wantURL:     nil,
			respError:   "failed to receive posts from DB",
			mockError:   errors.New("DB error"),
		},
		{
			name:        "Search_OK",
			argumentURL: "3",
			searchParam: "one",
			wantURL:     []string{"https://google.com"},
			respError:   "",
			mockError:   nil,
		},
		{
			name:        "Search_Not_found",
			argumentURL: "3",
			searchParam: "asdf",
			wantURL:     nil,
			respError:   "failed to receive posts from DB",
			mockError:   errors.New("DB error"),
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stMock := mocks.NewDB(t)

			// Исходя из тест-кейса устанавливаем поведение для мока только
			// если планируем дойти до него в тестируемой функции.
			if tt.respError == "" || tt.mockError != nil {
				stMock.
					On("Posts", mock.Anything, mock.AnythingOfType("int"), mock.Anything).
					Return(func(ctx context.Context, n int, q ...*storage.TextSearch) ([]storage.Post, error) {
						if q[0] == nil {
							return posts, tt.mockError
						}
						text := q[0].Query
						switch text {
						case "one":
							return posts[:1], tt.mockError
						case "two":
							return posts[1:2], tt.mockError
						case "three":
							return posts[2:], tt.mockError
						default:
							return posts, tt.mockError
						}
					}).
					Once()
			}

			mux := http.NewServeMux()
			mux.HandleFunc("GET /news/{n}", Posts(stMock))
			srv := httptest.NewServer(mux)
			defer srv.Close()

			str := fmt.Sprintf("/news/%s", tt.argumentURL)
			if tt.searchParam != "" {
				str = fmt.Sprintf("%s?s=%s", str, tt.searchParam)
			}
			req := httptest.NewRequest(http.MethodGet, str, nil)
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			body := rr.Body.String()

			if rr.Code != http.StatusOK {
				// Проверяем тело ответа и проваливаем тест, если содержимое
				// не совпадает с нашей ожидаемой ошибкой.
				body = strings.ReplaceAll(body, "\n", "")
				if body == tt.respError {
					t.SkipNow()
				}
				t.Fatalf("Posts() error = %s, want %s", body, tt.respError)
			}

			resp := []Response{}
			err := json.Unmarshal([]byte(body), &resp)
			if err != nil {
				t.Fatalf("Posts() error = cannot unmarshal response")
			}

			// Проверим только совпадение ссылок.
			urls := []string{}
			for _, v := range resp {
				urls = append(urls, v.Link)
			}

			if !reflect.DeepEqual(urls, tt.wantURL) {
				t.Errorf("Posts() = %v, want %v", urls, tt.wantURL)
			}
		})
	}
}

func Test_respConv(t *testing.T) {
	t.Parallel()

	type args struct {
		posts []storage.Post
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
	}{
		{
			name:    "respConv_OK",
			args:    args{posts: posts},
			wantLen: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := respConv(tt.args.posts)
			if got := len(resp); got != tt.wantLen {
				t.Errorf("respConv() = %v, want %v", got, tt.wantLen)
			}
		})
	}
}
