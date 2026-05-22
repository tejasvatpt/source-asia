package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/tejasvatpt/source-asia/internal/catalog"
	"github.com/tejasvatpt/source-asia/internal/handler"
	"github.com/tejasvatpt/source-asia/internal/ratelimit"
)

func main() {
	limiter := ratelimit.NewLimiter()
	store := catalog.NewStore()

	p1 := handler.NewPart1Handler(limiter)
	p2 := handler.NewPart2Handler(store)

	mux := http.NewServeMux()
	mux.HandleFunc("/request", p1.HandleRequest)
	mux.HandleFunc("/stats", p1.HandleStats)
	mux.HandleFunc("/products", p2.HandleProducts)
	mux.HandleFunc("/products/", p2.HandleProductByID)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)

	log.Printf("Server listening on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
