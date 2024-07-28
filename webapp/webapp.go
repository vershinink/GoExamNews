package webapp

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed dist
var dist embed.FS

// Serve возвращает каталог с клиентским приложением.
func Serve() http.FileSystem {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		log.Print(err)
	}

	return http.FS(sub)
}
