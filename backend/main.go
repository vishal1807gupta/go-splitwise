package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go-splitwise/controller"
	"go-splitwise/router"

	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

func main() {
	fmt.Println("Hello World")

	// Set up CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://go-splitwise.vercel.app"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "X-Auth-Token"},
		AllowCredentials: true,
	})

	r := router.Router()

	handler := c.Handler(r)

	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for range ticker.C {
			controller.CleanupExpiredSessions()
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	fmt.Println("Listening at port 4000...")
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
