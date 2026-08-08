package webhandler

import (
	"compress/gzip"
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/paper-indonesia/pdk/v2/chiExt"
	"github.com/paper-indonesia/pdk/v2/chiExt/middleware"
	webmiddleware "github.com/paper-indonesia/pivot-backoffice/handler/web/middleware"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/application"
	"github.com/paper-indonesia/pivot-backoffice/views"
)

type Handler interface {
	Register(*chi.Mux)
}

type WebServer struct {
	server *http.Server
}

func NewWebServer(addr string, app *application.Application, handler ...Handler) *WebServer {
	compressor := chiMiddleware.NewCompressor(gzip.DefaultCompression)

	muxer := chiExt.New(
		&chiExt.Config{
			LoggerWorkerCount:          10,
			LoggerWorkerExpiryDuration: 3 * time.Minute,
			Recorder:                   nil,
			ContextOverrideConfig: &middleware.ContextOverrideConfig{
				Methods: []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodGet},
			},
		},
		chiExt.WithLogger(app.GetLogger()),
		chiExt.WithNewRelic(app.GetNewRelic()),
	).(*chi.Mux) // chiExt.New return http.Handler, but we need chi.Mux

	// register assets handler
	muxer.Get("/assets/*", webmiddleware.AssetsCache(http.FileServer(views.Assets)))

	for _, h := range handler {
		h.Register(muxer)
	}

	return &WebServer{
		server: &http.Server{
			Addr:        addr,
			ReadTimeout: 5 * time.Second,
			Handler:     chiMiddleware.Recoverer(chiMiddleware.Logger(compressor.Handler(muxer))),
		},
	}
}

func (w *WebServer) ListenAndServe() error {
	return w.server.ListenAndServe()
}

func (s *WebServer) Stop(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}
