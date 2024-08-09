// Пакет парсера RSS лент.
package parser

import (
	"GoNews/internal/rss"
	"GoNews/internal/storage/memdb"
	"net/http"
	"reflect"
	"testing"
	"time"
)

var parser = &Parser{
	links:  []string{"https://habr.com/ru/rss/hub/go/all/?fl=ru"},
	period: time.Minute * 5,
	client: &http.Client{
		Timeout: reqTime,
	},
	storage: memdb.New(),
}

func TestParser_Start(t *testing.T) {
	parserNoURL := *parser
	parserNoURL.links = []string{}
	parserIncorrect := *parser
	parserIncorrect.links = []string{"abcdef"}

	tests := []struct {
		name    string
		p       *Parser
		wantErr bool
	}{
		{
			name:    "URL_OK",
			p:       parser,
			wantErr: false,
		},
		{
			name:    "No_URLs",
			p:       &parserNoURL,
			wantErr: true,
		},
		{
			name:    "Incorrect_URL",
			p:       &parserIncorrect,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.Start(); (err != nil) != tt.wantErr {
				t.Errorf("Parser.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParser_parseRSS(t *testing.T) {
	st := memdb.New()
	var parser = &Parser{
		links:  []string{"https://habr.com/ru/rss/hub/go/all/?fl=ru"},
		period: time.Minute * 5,
		client: &http.Client{
			Timeout: reqTime,
		},
		storage: st,
	}

	type args struct {
		url string
	}
	tests := []struct {
		name      string
		p         *Parser
		args      args
		wantCount int
	}{
		{
			name:      "URL_OK",
			p:         parser,
			args:      args{url: "https://habr.com/ru/rss/hub/go/all/?fl=ru"},
			wantCount: 40,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go tt.p.parseRSS(tt.args.url)
			time.Sleep(time.Second * 5)
			if st.Len() != tt.wantCount {
				t.Errorf("Parser.parseRSS() = %v, want %v", st.Len(), tt.wantCount)
			}
		})
	}
}

func Test_postConv(t *testing.T) {
	resp, err := http.Get(parser.links[0])
	if err != nil {
		t.Errorf("cannot receive RSS feed from url: %s", parser.links[0])
	}
	defer resp.Body.Close()
	feed, err := rss.Parse(resp.Body)
	if err != nil {
		t.Errorf("cannot decode RSS feed")
	}

	type args struct {
		feed rss.Feed
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "OK",
			args: args{
				feed: feed,
			},
			want: 40,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got int
			posts := postConv(tt.args.feed)
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

	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "RFC1123",
			args: args{str: "Sat, 27 Jul 2024 00:00:00 UTC"},
			want: unix,
		},
		{
			name: "RFC1123Z",
			args: args{str: "Sat, 27 Jul 2024 00:00:00 +0000"},
			want: unix,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := timeConv(tt.args.str).Unix(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("timeConv() = %v, want %v", got, tt.want)
			}
		})
	}
}
