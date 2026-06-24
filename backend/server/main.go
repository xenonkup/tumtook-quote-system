package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"tumtook-quote/backend/internal/quote"
	"tumtook-quote/backend/routes"
)

func main() {
	ctx := context.Background()
	if err := quote.InitDB(ctx); err != nil {
		log.Fatalf("เชื่อม Postgres ไม่ได้: %v", err)
	}
	defer quote.CloseDB()
	log.Println("เชื่อม Postgres + สร้างตาราง quotes สำเร็จ")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Tumtook Quote API running on :%s", port)
	if err := http.ListenAndServe(":"+port, routes.Handler()); err != nil {
		log.Fatal(err)
	}
}
