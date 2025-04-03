package main

import (
	"log"

	"github.com/guilhermedesousa/social/internal/db"
	"github.com/guilhermedesousa/social/internal/env"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	addr := env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable")
	conn, err := db.New(addr, 3, 3, "15m")
	if err != nil {
		log.Panic(err)
	}

	defer conn.Close()

	log.Println("database connection pool established")

	db.SeedDatabase(conn)
}
