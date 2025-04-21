package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"go-splitwise/controller"
	"go-splitwise/router"

	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

func main() {
	fmt.Println("Hello World")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})

	r := router.Router()

	handler := c.Handler(r)
	log.Fatal(http.ListenAndServe(":4000", handler))
	fmt.Println("Listening at port 4000...")

	ticker := time.NewTicker(24 * time.Hour) // Once per day
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			controller.CleanupExpiredSessions()
		}
	}()

}
