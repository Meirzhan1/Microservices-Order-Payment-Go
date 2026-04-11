package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	dsn := "postgres://postgres:postgres@localhost:5432/order_db?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	orderID := "cd7a4534-985c-4207-aabe-fb69f68f6067"
	res, err := db.Exec("UPDATE orders SET status = 'Cancelled' WHERE id = $1", orderID)
	if err != nil {
		log.Fatal(err)
	}
	rows, _ := res.RowsAffected()
	fmt.Printf("Updated %d rows\n", rows)
}
