// Creates demo.db in current directory for db-agent-safety showcase.
// Run: cd showcase/db-agent-safety && go run ../../scripts/init-demo-db/
package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	_ = os.Remove("demo.db")
	db, err := sql.Open("sqlite", "demo.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE products (id INTEGER PRIMARY KEY, name TEXT, price REAL);
		INSERT INTO products (name, price) VALUES ('Widget A', 19.99), ('Widget B', 29.99), ('Gadget X', 49.99);
		CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT, password_hash TEXT);
		INSERT INTO users (email, password_hash) VALUES ('admin@example.com', 'hash123'), ('user@example.com', 'hash456');
		CREATE TABLE sales (id INTEGER PRIMARY KEY, product_id INTEGER, amount REAL);
		INSERT INTO sales (product_id, amount) VALUES (1, 2), (2, 1), (1, 5);
	`)
	if err != nil {
		panic(err)
	}
	fmt.Println("Created demo.db with products, users, sales")
}
