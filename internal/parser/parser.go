package parser

import (
	"GoNews/internal/config"
	"GoNews/internal/logger"
	"GoNews/internal/rss"
	"GoNews/internal/storage"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// Parser - структура парсера RSS лент.
type Parser struct {
	links   []string
	period  time.Duration
	client  *http.Client
	log     *slog.Logger
	storage storage.Interface
}

// New - конструктор парсера RSS.
func New(cfg *config.Config, log *slog.Logger, st storage.Interface) *Parser {
	parser := &Parser{
		links:   cfg.RSSFeeds,
		period:  cfg.RequestPeriod,
		client:  &http.Client{},
		log:     log,
		storage: st,
	}
	return parser
}

// Start запускает парсинг каждого url в отдельной горутине с шагом, указанным в файле конфига.
func (p *Parser) Start() {
	for _, url := range p.links {
		go p.parseRSS(url)
	}

	p.log.Info("Parser started")
}

func (p *Parser) parseRSS(url string) {

	// TODO: добавить контексты.

	// Создаем структуру запроса для переданного url.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		p.log.Error("cannot create new request", slog.String("url", url), logger.Err(err))
		return
	}

	// Запускаем бесконечный цикл с парсингом RSS ленты и записи постов в БД.
	for {
		p.log.Debug("requesting data", slog.String("url", url))

		// Делаем запрос к RSS ленте. Если вернулась ошибка, то приостанавливаем цикл на время из конфига.
		resp, err := p.client.Do(req)
		if err != nil {
			p.log.Error("cannot receive a response", slog.String("url", url), logger.Err(err))
			time.Sleep(p.period)
			continue
		}

		// Парсим данные из тела ответа. Если ошибка, то вычитываем все тело ответа и закрываем его.
		// Затем приостанавливаем цикл на время из конфига.
		feed, err := rss.Parse(resp.Body)
		if err != nil {
			p.log.Error("cannot parse RSS feed", slog.String("url", url), logger.Err(err))

			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			time.Sleep(p.period)
			continue
		}

		// Вычитываем все тело ответа и закрываем его, освобождая память.
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		str := fmt.Sprintf("Parsed %d posts from %s\n", len(feed.Channel.Items), url)
		p.log.Info(str)

		time.Sleep(p.period)
	}
}
