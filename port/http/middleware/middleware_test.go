package middleware_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	chi "github.com/go-chi/chi/v5"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

func TestMain(m *testing.M) {
	cwd, _ := os.Getwd()
	projectRoot, _ := util.FindProjectRoot(cwd, "backend-portal")
	targetPath := filepath.Join(projectRoot, "test", "consul", "backend-portal", "feature-flag.yaml")

	_ = ffclient.Init(ffclient.Config{
		Retriever:    &fileretriever.Retriever{Path: targetPath},
		DataExporter: ffclient.DataExporter{},
	})
	defer ffclient.Close()

	m.Run()
}

func MountHandlers(router *chi.Mux, middlewares ...func(http.Handler) http.Handler) {
	if len(middlewares) > 0 {
		router.Use(middlewares...)
	}
	router.Get("/test", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message": "OK"}`))
		w.WriteHeader(http.StatusOK)
	})

	router.Post("/test", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message": "OK"}`))
		w.WriteHeader(http.StatusOK)
	})
}
