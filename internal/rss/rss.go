// Пакет для декодирования RSS потока.
package rss

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
)

var (
	ErrBodyNil   = errors.New("the response body is nil")
	ErrEmptyFeed = errors.New("the feed is empty")
)

// Feed - структура вывода RSS потока.
type Feed struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Items []Item `xml:"item"`
	} `xml:"channel"`
}

// Item - структура одного поста в RSS потоке.
type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Link        string `xml:"link"`
}

// Parse десериализует RSS поток в структуру Feed.
func Parse(body io.Reader) (Feed, error) {
	const operation = "rss.Parse"
	var feed Feed

	if body == nil {
		return feed, fmt.Errorf("%s: %w", operation, ErrBodyNil)
	}

	d := xml.NewDecoder(body)
	err := d.Decode(&feed)
	if err != nil {
		return feed, fmt.Errorf("%s: %w", operation, err)
	}

	if len(feed.Channel.Items) == 0 {
		return feed, fmt.Errorf("%s: %w", operation, ErrEmptyFeed)
	}

	return feed, nil
}
