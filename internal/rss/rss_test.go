// Пакет для декодирования RSS потока.
package rss

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

var (
	testDataNoItem = `
		<rss xmlns:dc="http://purl.org/dc/elements/1.1/" version="2.0">
			<channel>
				<title>Habr</title>
				<link>https://habr.com/ru/hubs/go/articles/</link>
				<description>Go – компилируемый, многопоточный язык программирования</description>
				<language>ru</language>
				<managingEditor>editor@habr.com</managingEditor>
				<generator>habr.com</generator>
				<pubDate>Thu, 25 Jul 2024 05:41:25 GMT</pubDate>
				<image>
					<link>https://habr.com/ru/</link>
					<url>https://habrastorage.org/webt/ym/el/wk/ymelwk3zy1gawz4nkejl_-ammtc.png</url>
					<title>Хабр</title>
				</image>
			</channel>
		</rss>
	`
	testDataIncorrect = `
		<rss xmlns:dc="http://purl.org/dc/elements/1.1/" version="2.0">
			<channel>
				<title>Habr</title>
				<link>https://habr.com/ru/hubs/go/articles/&</link>
			</channel>
		</rss>
	`
)

// TestParse позволяет проверить корректность десериализации RSS фрагмента,
// а также все пользовательские ошибки.
func TestParse(t *testing.T) {
	data, err := os.ReadFile("testFeed.xml")
	if err != nil {
		t.Fatalf("Parse() error = cannot read test XML feed")
	}
	dataOK := bytes.NewReader(data)

	var readerNil io.Reader
	var errDataIncorrect *xml.SyntaxError

	tests := []struct {
		name    string
		rss     io.Reader
		count   int
		wantErr bool
		gotErr  error
	}{
		{
			name:    "Parse_OK",
			rss:     dataOK,
			count:   2,
			wantErr: false,
			gotErr:  nil,
		},
		{
			name:    "Parse_No_Item",
			rss:     strings.NewReader(testDataNoItem),
			count:   0,
			wantErr: true,
			gotErr:  ErrEmptyFeed,
		},
		{
			name:    "Parse_Incorrect",
			rss:     strings.NewReader(testDataIncorrect),
			count:   0,
			wantErr: true,
			gotErr:  errDataIncorrect,
		},
		{
			name:    "Parse_Empty_Body",
			rss:     strings.NewReader(""),
			count:   0,
			wantErr: true,
			gotErr:  io.EOF,
		},
		{
			name:    "Data_Body_Nil",
			rss:     readerNil,
			count:   0,
			wantErr: true,
			gotErr:  ErrBodyNil,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			feed, err := Parse(tt.rss)

			// Если вернулась ошибка, но мы ее не ожидали - проваливаем тест.
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Если вернулась ошибка, и мы ее ожидали - проверяем тип ошибки.
			// При совпадении завершаем тест-кейс успешно.
			if tt.wantErr {
				if errors.Is(err, tt.gotErr) {
					t.Skipf("Parse() error matched, got = %v, want = %v", err, tt.gotErr)
				}
				if errors.As(err, &errDataIncorrect) {
					t.Skipf("Parse() error matched, got = %v, want = %v", err, tt.gotErr)
				}
				t.Fatalf("Parse() error not matched, got = %v, want = %v", err, tt.gotErr)
				return
			}

			// Проверяем наличие элементов в слайсе Items. Если слайс пустой, то
			// парсинг не прошел - проваливаем тест.
			got := len(feed.Channel.Items)
			if got != tt.count {
				t.Errorf("Parse() length = %v, want = %v", got, tt.count)
			}
		})
	}
}
