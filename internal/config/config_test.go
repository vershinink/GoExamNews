// Пакет для работы с файлом конфига.

package config

import (
	"testing"
)

func TestMustLoad(t *testing.T) {
	var got *Config = MustLoad()
	if got == nil {
		t.Fatalf("failed to load config")
	}
}
