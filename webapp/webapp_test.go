// Пакет webapp работает с собранными файлами клиентского приложения.
//
// Файлы клиентского приложения предоставлены SkillFactory и отдельно
// не изменялись.
package webapp

import (
	"testing"
)

func TestServe(t *testing.T) {
	sub := Serve()
	f, err := sub.Open("index.html")
	if err != nil {
		t.Errorf("Serve() error = %v", err)
	}
	defer f.Close()
}
