// Пакет парсера RSS лент.
package parser

import (
	"GoNews/internal/config"
	"GoNews/internal/logger"
	"GoNews/internal/rss"
	"GoNews/internal/storage"
	"context"
	"errors"
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

var (
	ErrNoRssLinks = errors.New("config file RSS section is empty")
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
		links:   cfg.RSSFeeds,
		period:  cfg.RequestPeriod,
		client:  &http.Client{},
		storage: st,
		done:    make(chan bool),
	}
	return parser
}

// Start проверяет каждый url из p.links на валидность, затем запускает парсинг
// в отдельной горутине с шагом, указанным в файле конфига.
func (p *Parser) Start() {
	if len(p.links) == 0 {
		slog.Error("there are no correct rss links in the config file")
		slog.Error("parser cannot start")
		return
	}

	// Счетчик запущенных парсеров.
	var i int
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

	slog.Debug("parser started on N urls", slog.Int("N", i))
}

// parseRSS запускает парсинг RSS ленты с переданного url и записывает результаты в БД.
func (p *Parser) parseRSS(url string) {

	// Создаем контекст с отменой и запускаем ожидание этой отмены в отдельной горутине.
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-p.done
		cancel()
	}()

	// Создаем регулярное выражение для вырезания пустых строк из поля description.
	regex, err := regexp.Compile(`[\n]{2,}[\s]+`)
	if err != nil {
		slog.Error("cannot compile regexp", logger.Err(err))
	}

	// Создаем структуру запроса для переданного url.
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		slog.Error("cannot create new request", slog.String("url", url), logger.Err(err))
		return
	}

	// Запускаем бесконечный цикл с парсингом RSS ленты и записи постов в БД.
	for {
		slog.Debug("requesting data", slog.String("url", url))

		// Делаем запрос к RSS ленте. Если вернулась ошибка, то приостанавливаем цикл на время из конфига.
		resp, err := p.client.Do(req)
		if err != nil {
			slog.Error("cannot receive a response", slog.String("url", url), logger.Err(err))
			time.Sleep(p.period)
			continue
		}

		// Парсим данные из тела ответа. Если ошибка, то вычитываем все тело ответа и закрываем его.
		// Затем приостанавливаем цикл на время из конфига.
		feed, err := rss.Parse(resp.Body)
		if err != nil {
			slog.Error("cannot parse RSS feed", slog.String("url", url), logger.Err(err))

			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			time.Sleep(p.period)
			continue
		}

		// Вычитываем все тело ответа и закрываем его, освобождая память.
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		slog.Debug("data parsed successfully", slog.Int("posts", len(feed.Channel.Items)), slog.String("url", url))

		// Создаем канал posts с емкостью, равной количеству постов из спарсенной ленты.
		// Асинхронно подготавливаем каждый пост и отправляем в канал. Из этого канала будем
		// считывать данные и записывать в БД.
		posts := make(chan storage.Post, len(feed.Channel.Items))
		var wg sync.WaitGroup
		wg.Add(len(feed.Channel.Items))
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

		// После обработки всех постов закрываем канал.
		go func() {
			wg.Wait()
			close(posts)
		}()

		slog.Debug("sending data to DB", slog.String("url", url))

		// Вызываем метод для записи данных из канала в БД.
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

		// Приостанавливаем цикл на время из конфига. Затем начинаем заново.
		time.Sleep(p.period)
	}
}

// Shutdown посылает сигналы для остановки парсинга каждой url.
func (p *Parser) Shutdown() {
	for i := 0; i < len(p.links); i++ {
		p.done <- true
	}
	close(p.done)
}

// timeConv конвертирует переданную дату в time.Time в зависимости от формата.
// Если формат переданной даты не соответствует проверяемым, то возвращает текущее время и дату.
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
