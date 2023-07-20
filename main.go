package main

import (
	"fmt"
	"github.com/go-chi/chi"
	"gonnachat/internal/handlers"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
)

const defaultPort = "3000"

func main() {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = defaultPort
	}

	r := chi.NewRouter()

	r.Mount("/debug/pprof", http.DefaultServeMux)
	r.Get("/mutex", handlers.MutexState)
	r.Get("/ws", handlers.WSChat)

	log.Printf("Server starting on port %s", port)

	err := http.ListenAndServe(fmt.Sprintf(":%s", port), r)
	if err != nil {
		log.Fatal(err)
	}
}
