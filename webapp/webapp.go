// Пакет webapp работает с собранными файлами клиентского приложения.
//
// Файлы клиентского приложения предоставлены SkillFactory и отдельно
// не изменялись.
package webapp

import (
	"embed"
	"io/fs"
	"log/slog"
)

//go:embed dist
var dist embed.FS

// Serve возвращает каталог с клиентским приложением.
func Serve() fs.FS {
	const operation = "webapp.Serve"

	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		slog.Error("webapp error", slog.String("op", operation))
	}

	return sub
}
