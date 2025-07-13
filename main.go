package main

import (
	"log"
	"net/http"
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

	log.Println("Server started on http://localhost:8080")
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
