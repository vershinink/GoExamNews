// Пакет парсера RSS лент.
package parser

import (
	"GoNews/internal/mocks"
	"GoNews/internal/rss"
	"GoNews/internal/storage"
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

func TestParser_Start(t *testing.T) {
	feed, err := os.ReadFile("testFeed.xml")
	if err != nil {
		t.Fatalf("Parser.Start() error = cannot read test XML feed")
	}

	tests := []struct {
		name      string
		urls      []string
		wantStart int
	}{
		// Проверяем корректные и некорректные ссылки, а также их
		// отсутствие. Если wantStart > 0, то мы не хотим получить
		// ошибку из тестируемой функции.
		{
			name:      "URL_OK",
			urls:      []string{"https://good-url.com", "https://good-url.com", "https://good-url.com"},
			wantStart: 3,
		},
		{
			name:      "No_URL",
			urls:      []string{},
			wantStart: 0,
		},
		{
			name:      "Incorrect_URL",
			urls:      []string{"asdf"},
			wantStart: 0,
		},
		{
			name:      "Partially_Correct",
			urls:      []string{"asdf", "https://good-url.com", "zxcv"},
			wantStart: 1,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var parser = &Parser{
				links:  tt.urls,
				period: time.Minute * 5,
				client: &http.Client{
					Transport: nil,
					Timeout:   reqTime,
				},
				storage: nil,
				done:    make(chan bool, len(tt.urls)),
			}

			var reqCount int
			// Создаем и настраиваем моки RoundTripper и хранилища, если
			// ожидается успешный запуск хотя бы одного парсинга.
			if tt.wantStart > 0 {
				rtMock := mocks.NewRoundTripper(t)
				rtMock.
					On("RoundTrip", mock.AnythingOfType("*http.Request")).
					Return(func(req *http.Request) (*http.Response, error) {
						resp := &http.Response{
							Status:        "200 OK",
							StatusCode:    200,
							Proto:         "HTTP/1.1",
							ProtoMajor:    1,
							ProtoMinor:    1,
							Body:          io.NopCloser(bytes.NewBuffer(feed)),
							ContentLength: int64(len(feed)),
							Request:       req,
							Header:        make(http.Header),
						}
						reqCount++
						return resp, nil
					}).
					Times(tt.wantStart)
				parser.client.Transport = rtMock

				stMock := mocks.NewDB(t)
				stMock.
					On("AddPosts", mock.Anything, mock.AnythingOfType("<-chan storage.Post")).
					Return(2, nil).
					Times(tt.wantStart)
				parser.storage = stMock
			}

			err := parser.Start()
			if err != nil {
				// Если ошибка ожидаема, то успешно завершаем тест-кейс.
				if tt.wantStart == 0 {
					t.SkipNow()
				}
				t.Fatalf("Parser.Start() error = %v", err)
			}

			time.Sleep(time.Second * 5)
			parser.Shutdown()

			if reqCount != tt.wantStart {
				t.Errorf("Parser.Start() starts = %d, want = %d", reqCount, tt.wantStart)
			}
		})
	}
}

func TestParser_parseRSS(t *testing.T) {
	feed, err := os.ReadFile("testFeed.xml")
	if err != nil {
		t.Fatalf("Parser.parseRSS() error = cannot read test XML feed")
	}

	tests := []struct {
		name      string
		url       string
		wantCount int
		wantError bool
		mockError error
	}{
		// Тестируем три основных пути: успешный путь, ошибка сети
		// и ошибка базы данных. Ошибки самого парсинга RSS ленты
		// тестируются в пакете rss.
		{
			name:      "URL_OK",
			url:       "https://good-url.com",
			wantCount: 2,
			wantError: false,
			mockError: nil,
		},
		{
			name:      "URL_Bad",
			url:       "https://bad-url.com",
			wantCount: 0,
			wantError: true,
			mockError: nil,
		},
		{
			name:      "DB_error",
			url:       "good-url.com",
			wantCount: 0,
			wantError: true,
			mockError: errors.New("DB error"),
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var parser = &Parser{
				period: time.Minute * 5,
				client: &http.Client{
					Transport: nil,
					Timeout:   reqTime,
				},
				storage: nil,
				done:    make(chan bool, 1),
			}

			// Имитируем поведение RSS ресурса через содание мока интерфейса
			// RoundTripper в клиенте.
			rtMock := mocks.NewRoundTripper(t)
			rtMock.
				On("RoundTrip", mock.AnythingOfType("*http.Request")).
				Return(func(req *http.Request) (*http.Response, error) {
					var err error
					resp := &http.Response{
						Proto:      "HTTP/1.1",
						ProtoMajor: 1,
						ProtoMinor: 1,
						Request:    req,
						Header:     make(http.Header),
					}
					switch req.URL.Hostname() {
					case "bad-url.com":
						resp.Status = "404 Not Found"
						resp.StatusCode = 404
						resp.Body = io.NopCloser(bytes.NewBufferString("Not Found"))
						resp.ContentLength = int64(len("Not Found"))
						err = errors.New("Page not found")
					default:
						resp.Status = "200 OK"
						resp.StatusCode = 200
						resp.Body = io.NopCloser(bytes.NewBuffer(feed))
						resp.ContentLength = int64(len(feed))
					}
					return resp, err
				}).
				Once()
			parser.client.Transport = rtMock

			var count int

			// Создаем и настраиваем мок хранилища только если планируем дойти
			// до него в тестируемой функции.
			if !tt.wantError || tt.mockError != nil {
				stMock := mocks.NewDB(t)
				stMock.
					On("AddPosts", mock.Anything, mock.AnythingOfType("<-chan storage.Post")).
					Return(func(ctx context.Context, posts <-chan storage.Post) (int, error) {
						if tt.wantError {
							return 0, tt.mockError
						}
						for v := range posts {
							_ = v
							count++
						}
						return count, nil
					}).
					Once()
				parser.storage = stMock
			}

			go parser.parseRSS(tt.url)

			time.Sleep(time.Second * 5)
			parser.Shutdown()
			if count != tt.wantCount {
				t.Errorf("Parser.parseRSS() = %v, want = %v", count, tt.wantCount)
			}
		})
	}
}

func Test_postConv(t *testing.T) {
	file, err := os.ReadFile("testFeed.xml")
	if err != nil {
		t.Fatalf("postConv() error = cannot read test XML file")
	}
	rd := bytes.NewReader(file)

	feedOK, err := rss.Parse(rd)
	if err != nil {
		t.Fatalf("postConv() error = cannot decode RSS feed")
	}
	var feedEmpty rss.Feed

	tests := []struct {
		name string
		feed rss.Feed
		want int
	}{
		// Значение want должно совпасть с количеством элементов Item
		// в структуре rss.Feed.
		{
			name: "PostConv_OK",
			feed: feedOK,
			want: 2,
		},
		{
			name: "PostConv_Empty_Feed",
			feed: feedEmpty,
			want: 0,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			posts := postConv(tt.feed)
			if tt.want == 0 {
				if posts == nil {
					t.SkipNow()
				}
				t.Fatalf("postConv() = unexpected nil")
			}

			var got int
			for p := range posts {
				_ = p
				got++
			}
			if got != tt.want {
				t.Errorf("postConv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_timeConv(t *testing.T) {
	tm, _ := time.Parse(time.RFC1123, "Sat, 27 Jul 2024 00:00:00 UTC")
	unix := tm.Unix()

	tests := []struct {
		name string
		time string
		want int64
	}{
		{
			name: "RFC1123",
			time: "Sat, 27 Jul 2024 00:00:00 UTC",
			want: unix,
		},
		{
			name: "RFC1123Z",
			time: "Sat, 27 Jul 2024 00:00:00 +0000",
			want: unix,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := timeConv(tt.time).Unix(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("timeConv() = %v, want %v", got, tt.want)
			}
		})
	}
}
