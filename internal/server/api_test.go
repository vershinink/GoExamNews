// Пакет для работы с сервером и обработчиками API.

package server

import (
	"GoNews/internal/logger"
	"GoNews/internal/mocks"
	"GoNews/internal/storage"
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
	{ID: "1", Title: "Title 1", Content: "Content 1", PubTime: time.Now(), Link: "https://google.com"},
	{ID: "2", Title: "Title 2", Content: "Content 3", PubTime: time.Now(), Link: "https://ya.ru"},
	{ID: "3", Title: "Title 2", Content: "Content 3", PubTime: time.Now(), Link: "https://bing.com"},
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
		wantURL     []string
		respError   string
		mockError   error
	}{
		{
			name:        "Posts_OK",
			argumentURL: "3",
			wantURL:     []string{"https://google.com", "https://ya.ru", "https://bing.com"},
			respError:   "",
			mockError:   nil,
		},
		{
			name:        "Incorrect_GET_request",
			argumentURL: "asd",
			wantURL:     nil,
			respError:   "incorrect posts number",
			mockError:   nil,
		},
		{
			name:        "DB_error",
			argumentURL: "3",
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
					On("Posts", mock.Anything, mock.AnythingOfType("int")).
					Return(posts, tt.mockError).
					Once()
			}

			mux := http.NewServeMux()
			mux.HandleFunc("GET /news/{n}", Posts(stMock))
			srv := httptest.NewServer(mux)
			defer srv.Close()

			str := fmt.Sprintf("/news/%s", tt.argumentURL)
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
