package db

import (
	"database/sql"
	"log"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

var db *sql.DB
var once sync.Once

func Connect() *sql.DB {
	once.Do(func() {
		databaseURL, ok := os.LookupEnv("DATABASE_URL")
		if !ok {
			log.Fatalf("DATABASE_URL is required")
		}
		var err error
		db, err = sql.Open("postgres", databaseURL)
		if err != nil {
			log.Fatal(err)
		}
		if err := db.Ping(); err != nil {
			log.Fatal(err)
		}
		log.Println("Connected to database")
	})

	return db
}
