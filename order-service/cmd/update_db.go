package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	dsn := "postgres://postgres:postgres@localhost:5432/order_db?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	orderID := "fb7ea212-64df-4a14-a6a5-5f40cabf0e30"
	res, err := db.Exec("UPDATE orders SET status = 'Cancelled', updated_at = $1 WHERE id = $2", time.Now().UTC(), orderID)
	if err != nil {
		log.Fatal(err)
	}
	rows, _ := res.RowsAffected()
	fmt.Printf("Updated %d rows\n", rows)
}
