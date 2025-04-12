package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/ini.v1"
)

func generateRandomToken(size int) (string, error) {
	token := make([]byte, size)

	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}

	encodedToken := base64.URLEncoding.EncodeToString(token)
	return encodedToken, nil
}

var DB_ENDPOINT string
var DB_USER string
var DB_PASSWORD string
var DB_NAME string
var DB_PORT string

func connectDB() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", DB_USER, DB_PASSWORD, DB_ENDPOINT, DB_PORT, DB_NAME)
	return sql.Open("mysql", dsn)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		fmt.Fprintln(w, "Ok")
		return
	}

	db, err := connectDB()
	if err != nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
	}
	defer db.Close()

	query := fmt.Sprintf("SELECT token FROM %s WHERE id=? LIMIT 1", DB_NAME)
	var token string
	err = db.QueryRow(query, id).Scan(&token)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("Db query failed: %v", err), http.StatusInternalServerError)
		return
	}
	if err != sql.ErrNoRows {
		fmt.Fprintln(w, token)
		return
	}
	token, err = generateRandomToken(128)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
	}

	query = fmt.Sprintf("INSERT INTO %s (id, token) VALUES (?, ?)", DB_NAME)
	_, err = db.Exec(query, id, token)
	if err != nil {
		http.Error(w, "Failed to insert data into table", http.StatusInternalServerError)
	}
	fmt.Fprintln(w, token)
}

func main() {
	cfg, err := ini.Load("./server.ini")
	if err != nil {
		log.Fatal("Failed to read server.ini", err)
	}

	DB_ENDPOINT = cfg.Section("").Key("DB_ENDPOINT").String()
	DB_USER = cfg.Section("").Key("DB_USER").String()
	DB_PASSWORD = cfg.Section("").Key("DB_PASSWORD").String()
	DB_NAME = cfg.Section("").Key("DB_NAME").String()
	DB_PORT = cfg.Section("").Key("DB_PORT").String()

	http.HandleFunc("/", rootHandler)
	fmt.Println("Starting server on port 8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
