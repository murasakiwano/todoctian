package todoctian

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/murasakiwano/todoctian/server/internal/openapi"
)

func Handler(pgConnString string) http.Handler {
	server := NewServer(pgConnString)
	return openapi.Handler(server, openapi.ServerOption(func(so *openapi.ServerOptions) {
		so.BaseRouter.Use(middleware.Logger)
	}))
}
