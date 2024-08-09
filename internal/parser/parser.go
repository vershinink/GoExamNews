// Пакет парсера RSS лент.
package parser

import (
	"GoNews/internal/config"
	"GoNews/internal/logger"
	"GoNews/internal/rss"
	"GoNews/internal/storage"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
	strip "github.com/grokify/html-strip-tags-go"
)

// reqTime - таймаут для запроса RSS ленты.
const reqTime time.Duration = time.Second * 10

var (
	ErrNoLinks = errors.New("RSS section of the config file has no correct URLs")
)

// Parser - структура парсера RSS лент.
type Parser struct {
	links   []string
	period  time.Duration
	client  *http.Client
	storage storage.Interface
	done    chan bool
}

// New - конструктор парсера RSS.
func New(cfg *config.Config, st storage.Interface) *Parser {
	parser := &Parser{
		links:  cfg.RSSFeeds,
		period: cfg.RequestPeriod,
		client: &http.Client{
			Timeout: reqTime,
		},
		storage: st,
		done:    make(chan bool),
	}
	return parser
}

// Start проверяет каждый url из списка ссылок p.links на валидность,
// затем запускает парсинг в отдельной горутине с шагом, указанным
// в файле конфига.
func (p *Parser) Start() error {
	if len(p.links) == 0 {
		return ErrNoLinks
	}

	// Счетчик запущенных парсеров, нужен для вывода в Debug сообщении.
	var i int
	// Валидатор нужен для проверки url на корректность.
	valid := validator.New()
	for _, url := range p.links {
		err := valid.Var(url, "url")
		if err != nil {
			slog.Error("invalid url", slog.String("url", url))
			continue
		}
		go p.parseRSS(url)
		i++
	}

	if i == 0 {
		return ErrNoLinks
	}

	slog.Debug(fmt.Sprintf("parser started on %d urls", i))
	return nil
}

// Shutdown посылает сигналы для остановки парсинга url.
func (p *Parser) Shutdown() {
	for i := 0; i < len(p.links); i++ {
		p.done <- true
	}
	close(p.done)
}

// parseRSS запускает парсинг RSS ленты с переданного url в бесконечном
// цикле и с периодом, указанным в парсере. Каждую итерацию цикла
// запрашивается и десериализуется RSS лента. Затем все новые посты
// записываются в БД.
func (p *Parser) parseRSS(url string) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-p.done
		cancel()
	}()

	// Создаем новый запрос с контекстом для переданного url. Контекст
	// используется для отмены запроса в случае остановки парсера.
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		slog.Error("cannot create new request", slog.String("url", url), logger.Err(err))
		return
	}

	for {
		slog.Debug("requesting data", slog.String("url", url))

		resp, err := p.client.Do(req)
		if err != nil {
			slog.Error("cannot receive a response", slog.String("url", url), logger.Err(err))
			time.Sleep(p.period)
			continue
		}

		feed, err := rss.Parse(resp.Body)

		// Для корректного переиспользования соединения и освобождения
		// памяти следует вычитать все тело ответа до EOF и закрыть его,
		// как указано в описании к методу Do клиента.
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if err != nil {
			slog.Error("cannot parse RSS feed", slog.String("url", url), logger.Err(err))
			time.Sleep(p.period)
			continue
		}

		slog.Debug("data parsed successfully", slog.Int("posts", len(feed.Channel.Items)), slog.String("url", url))

		posts := postConv(feed)

		slog.Debug("sending data to DB", slog.String("url", url))

		num, err := p.storage.AddPosts(ctx, posts)
		if err != nil {
			slog.Error("error on adding posts", slog.String("url", url), logger.Err(err))
			time.Sleep(p.period)
			continue
		}

		switch num {
		case 0:
			slog.Info("No posts was added", slog.String("url", url))
		default:
			slog.Info("Posts from url added successfully", slog.Int("posts", num), slog.String("url", url))
		}

		time.Sleep(p.period)
	}
}

// postConv создает и возвращает канал с емкостью, равной количеству
// постов из переданной RSS ленты. Асинхронно подготавливает каждый
// пост и отправляет в канал.
func postConv(feed rss.Feed) <-chan storage.Post {
	ln := len(feed.Channel.Items)
	posts := make(chan storage.Post, ln)
	var wg sync.WaitGroup
	wg.Add(ln)

	go func() {
		wg.Wait()
		close(posts)
	}()

	// Создаем регулярное выражение для вырезания пустых строк из поля
	// description. Функция StripTags из пакета strip вырезает HTML тэги,
	// но оставляет много пустых строк, если такие были.
	regex, err := regexp.Compile(`[\n]{2,}[\s]+`)
	if err != nil {
		slog.Error("cannot compile regexp", logger.Err(err))
	}

	for _, v := range feed.Channel.Items {
		go func(i rss.Item) {
			defer wg.Done()

			var p storage.Post
			p.Title = i.Title
			desc := strip.StripTags(i.Description)
			p.Content = regex.ReplaceAllString(desc, "\n")
			p.Link = i.Link
			p.PubTime = timeConv(i.PubDate)
			posts <- p
		}(v)
	}

	return posts
}

// timeConv конвертирует переданную дату в time.Time в зависимости от формата.
// Если формат переданной даты не соответствует проверяемым, то возвращает
// текущее время и дату.
func timeConv(str string) time.Time {
	r, _ := utf8.DecodeLastRuneInString(str)
	if r == utf8.RuneError {
		return time.Now()
	}

	var t time.Time
	var err error
	switch {
	case unicode.IsDigit(r):
		t, err = time.Parse(time.RFC1123Z, str)
	default:
		t, err = time.Parse(time.RFC1123, str)
	}
	if err != nil {
		return time.Now()
	}
	return t
}
