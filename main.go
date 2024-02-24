package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	dbInstance, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Fail to open connection with database: %v", err)
	}

	dbInstance.SetConnMaxLifetime(0)
	dbInstance.SetMaxIdleConns(1)
	dbInstance.SetMaxOpenConns(1)

	db = dbInstance
}

func main() {
	defer db.Close()

	store := NewPostgresTransactionStore(db)
	server := NewServer(store)

	addr := fmt.Sprintf(":%s", os.Getenv("API_PORT"))
	log.Printf("Listening in %s...", addr)

	err := http.ListenAndServe(addr, server)
	if err != nil {
		log.Fatalf("Fail to start server on addr: %q", addr)
	}
}
