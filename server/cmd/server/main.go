package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	todoctian "github.com/murasakiwano/todoctian/server"
)

func main() {
	pgConnString := os.Getenv("PG_DB_URL")
	r := chi.NewRouter()
	r.Mount("/", todoctian.Handler(pgConnString))

	http.ListenAndServe(":5656", r)
}
