package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bryangodson/detect-fraud-mvp/handlers"
	"github.com/bryangodson/detect-fraud-mvp/services"
	_ "github.com/lib/pq"
)

func main() {
	// Optional: initialize Postgres if DATABASE_URL env var is provided.
	// DATABASE_URL example: postgres://user:pass@localhost:5432/dbname?sslmode=disable
	dsn := os.Getenv("DATABASE_URL")
	if dsn != "" {
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			log.Fatalf("failed to open DB: %v", err)
		}
		// test connection
		if err := db.Ping(); err != nil {
			log.Fatalf("failed to ping DB: %v", err)
		}
		// assign to services.DB so logDecision can use it
		services.DB = db
		// create table if not exists (simple migration)
		_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS fraud_decisions (
  id SERIAL PRIMARY KEY,
  transaction_id TEXT,
  user_id TEXT,
  amount NUMERIC,
  score DOUBLE PRECISION,
  decision TEXT,
  reasons JSONB,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
)`)
		if err != nil {
			log.Fatalf("failed to create table: %v", err)
		}
		fmt.Println("✅ Connected to Postgres and ensured table exists")
	} else {
		fmt.Println("⚠️ DATABASE_URL not set — decisions will be written to decisions.log")
	}

	// Ensure MODEL_SERVER_URL is shown so you remember to set it
	if services.ModelServerURL == "" {
		fmt.Println("⚠️ MODEL_SERVER_URL not set — service will fail model calls until you set it")
	} else {
		fmt.Println("Model server:", services.ModelServerURL)
	}

	// Register the HTTP handler
	http.HandleFunc("/fraud-check", handlers.FraudCheckHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)
	fmt.Println("🚀 Fraud API listening on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
