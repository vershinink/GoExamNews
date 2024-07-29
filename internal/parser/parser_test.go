// Пакет парсера RSS лент.
package parser

import (
	"GoNews/internal/storage/memdb"
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"
)

var st = memdb.New()

var parser = &Parser{
	links:   []string{"https://habr.com/ru/rss/hub/go/all/?fl=ru"},
	period:  time.Minute * 5,
	client:  &http.Client{},
	log:     slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
	storage: st,
}

func TestParser_parseRSS(t *testing.T) {
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
			st.I = 0
			go tt.p.parseRSS(tt.args.url)
			time.Sleep(time.Second * 5)
			if st.I != tt.wantCount {
				t.Errorf("Parser.parseRSS = %v, want %v", st.I, tt.wantCount)
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
