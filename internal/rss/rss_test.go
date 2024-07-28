// Пакет для декодирования RSS потока.
package rss

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"
	"testing"
)

var (
	testDataOK = `
		<rss xmlns:dc="http://purl.org/dc/elements/1.1/" version="2.0">
			<channel>
				<title>
					<![CDATA[ Все статьи подряд / Go / Хабр ]]>
				</title>
				<link>https://habr.com/ru/hubs/go/articles/</link>
				<description>
					<![CDATA[ Go – компилируемый, многопоточный язык программирования ]]>
				</description>
				<language>ru</language>
				<managingEditor>editor@habr.com</managingEditor>
				<generator>habr.com</generator>
				<pubDate>Thu, 25 Jul 2024 05:41:25 GMT</pubDate>
				<image>
					<link>https://habr.com/ru/</link>
					<url>https://habrastorage.org/webt/ym/el/wk/ymelwk3zy1gawz4nkejl_-ammtc.png</url>
					<title>Хабр</title>
				</image>
				<item>
					<title>
						<![CDATA[ Как наш ученик попал на стажировку в VK. История Артёма Мазура ]]>
					</title>
					<guid isPermaLink="true">https://habr.com/ru/articles/831252/</guid>
					<link>https://habr.com/ru/articles/831252/?utm_campaign=831252&amp;utm_source=habrahabr&amp;utm_medium=rss</link>
					<description>
						<![CDATA[ <img src="https://habrastorage.org/getpro/habr/upload_files/0ea/8be/7bc/0ea8be7bcdd15e01a22aa60ade414a24.jpg" /><p>Мы следим за жизнью всех ребят, которые приходят в ЦПМ и участвуют в наших проектах. Каждый раз, когда мы узнаем об их достижениях, нам очень трепетно и радостно! Сегодня мы хотим поделиться историей Артёма Мазура, который прошел на стажировку, внимание, в VK!</p><p></p> <a href="https://habr.com/ru/articles/831252/?utm_campaign=831252&amp;utm_source=habrahabr&amp;utm_medium=rss#habracut">Читать далее</a> ]]>
					</description>
					<pubDate>Wed, 24 Jul 2024 20:21:32 GMT</pubDate>
					<dc:creator>
						<![CDATA[ kodIIm ]]>
					</dc:creator>
				</item>
				<item>
					<title>
						<![CDATA[ Ошибки в языке Go — это большая ошибка ]]>
					</title>
					<guid isPermaLink="true">https://habr.com/ru/companies/karuna/articles/830346/</guid>
					<link>https://habr.com/ru/companies/karuna/articles/830346/?utm_campaign=830346&amp;utm_source=habrahabr&amp;utm_medium=rss</link>
					<description>
						<![CDATA[ <pre><code class="go">// гофер пытается найти логику среди обработки ошибок +-------+-------+-------+-------+-------+-------+ | | err | | err | | err | | ,_,,, | | | | | | (◉ _ ◉) | | | | | | /) (\ | | | | | &quot;&quot; &quot;&quot; | | | | + +-------+ +-------+ +-------+ | | err | err | | err | | | | | | | | | | | | | +-------+ +-------+ +-------+ + | err | | err | | | | | | | | | + +-------+ + +-------+ + | | err | | err | logic | | | | | | | | | | | | | +-------+-------+-------+-------+-------+-------+</code></pre><br> <p>Я пишу на Go несколько лет, в Каруне многие вещи сделаны на нём; язык мне нравится своей простотой, незамысловатой прямолинейностью и приличной эффективностью. На других языках я писать не хочу.</p><br> <p>Но сорян, к бесконечным <code>if err != nil</code> я до конца привыкнуть так и не смог.</p><br> <p>Да-да, я знаю все аргументы: явное лучше неявного, язык Go многословен, зато понятен, и всё такое. Но, блин, на мой взгляд Го-вэй Го-вэю рознь.</p> <a href="https://habr.com/ru/articles/830346/?utm_campaign=830346&amp;utm_source=habrahabr&amp;utm_medium=rss#habracut">Читать дальше &rarr;</a> ]]>
					</description>
					<pubDate>Tue, 23 Jul 2024 11:36:03 GMT</pubDate>
					<dc:creator>
						<![CDATA[ varanio (Karuna) ]]>
					</dc:creator>
				</item>
			</channel>
		</rss>
	`
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

func TestParse(t *testing.T) {
	var readerNil io.Reader
	var errDataIncorrect *xml.SyntaxError

	type args struct {
		body io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
		gotErr  error
	}{
		{
			name:    "Data_OK",
			args:    args{body: strings.NewReader(testDataOK)},
			want:    true,
			wantErr: false,
			gotErr:  nil,
		},
		{
			name:    "Data_No_Item",
			args:    args{body: strings.NewReader(testDataNoItem)},
			want:    false,
			wantErr: true,
			gotErr:  ErrEmptyFeed,
		},
		{
			name:    "Data_Incorrect",
			args:    args{body: strings.NewReader(testDataIncorrect)},
			want:    false,
			wantErr: true,
			gotErr:  errDataIncorrect,
		},
		{
			name:    "Data_Empty_Body",
			args:    args{body: strings.NewReader("")},
			want:    false,
			wantErr: true,
			gotErr:  io.EOF,
		},
		{
			name:    "Data_Body_Nil",
			args:    args{body: readerNil},
			want:    false,
			wantErr: true,
			gotErr:  ErrBodyNil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feed, err := Parse(tt.args.body)

			// Если вернулась ошибка, но мы ее не ожидали - проваливаем тест.
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Если вернулась ошибка, и мы ее ожидали - проверяем тип ошибки.
			if tt.wantErr {
				if errors.Is(err, tt.gotErr) {
					t.Logf("Parse() error matched, got = %v, want = %v", err, tt.gotErr)
					return
				}
				if errors.As(err, &errDataIncorrect) {
					t.Logf("Parse() error matched, got = %v, want = %v", err, tt.gotErr)
					return
				}
				// Если ошибка не совпадает с описанными в кейсах, то проваливаем тест и выводим эту ошибку.
				t.Errorf("Parse() error nor matched, got = %v, want = %v", err, tt.gotErr)
				return
			}

			// Проверяем наличие элементов в слайсе Items. Если слайс пустой, то
			// парсинг не прошел - проваливаем тест.
			got := len(feed.Channel.Items) > 0
			if got != tt.want {
				t.Errorf("Parse() = %v, want = %v", got, tt.want)
			}
		})
	}
}
