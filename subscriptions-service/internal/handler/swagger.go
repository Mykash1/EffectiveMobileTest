package handler

import (
	"net/http"
	"os"
	"path/filepath"
)

func SwaggerUI() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		staticDir := "internal/handler/static"
		if _, err := os.Stat(staticDir); os.IsNotExist(err) {
			staticDir = "static"
		}

		filePath := filepath.Join(staticDir, filepath.Clean(r.URL.Path))
		if r.URL.Path == "/" || r.URL.Path == "" {
			filePath = filepath.Join(staticDir, "index.html")
		}

		http.ServeFile(w, r, filePath)
	})
}
