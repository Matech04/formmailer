package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {

	config := DefaultConfig()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   config.CORS,
		AllowedMethods:   []string{"POST"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: false,
	})

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		formHandler(w, r, config)
	})

	handler := c.Handler(http.DefaultServeMux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Domyślny port dla lokalnego rozwoju
	}

	// Użyj zmiennej port w logach i adresie serwera
	log.Printf("Server started on http://0.0.0.0:%s", port)
	server := &http.Server{
		Addr:         ":" + port, // Użyj zmiennej port
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}
