// Пакет для работы с файлом конфига.

package config

import (
	"testing"
)

// TestMustLoad позволяет проверить корректность указания пути
// к файлу конфига в переменных окружения.
func TestMustLoad(t *testing.T) {
	var got *Config = MustLoad()
	if got == nil {
		t.Fatalf("failed to load config")
	}
}
