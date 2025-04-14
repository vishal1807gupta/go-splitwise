package main

import (
	"fmt"
	"log"
	"net/http"

	"go-splitwise/router"

	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

func main() {
	fmt.Println("Hello World")

	// insertStmt := `insert into "users" ("name","email","password") values('Vishal','vishal@gmail.com','vish18')`
	// insertDynStmt := `insert into "users" ("name","email","password") values($1,$2,$3)`
	// //_, err = db.Exec(insertStmt)
	// _, err = db.Exec(insertDynStmt, "Nishant", "nishant@gmail.com", "nish18")

	// if err != nil {
	// 	panic(err)
	// }

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

}
