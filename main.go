package main

import (
	"github.com/go-chi/chi"
	"gonnachat/internal/handlers"
	"log"
	"net/http"
)

func main() {
	r := chi.NewRouter()

	r.Get("/ws", handlers.WSChat)

	err := http.ListenAndServe(":3000", r)
	if err != nil {
		log.Fatal(err)
	}
}
