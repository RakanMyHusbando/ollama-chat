package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var (
	host      string
	port      string
	ollamaUrl string
)

func main() {
	godotenv.Load(".env")
	host = os.Getenv("HOST")
	port = os.Getenv("PORT")
	ollamaUrl = os.Getenv("OLLAMA_URL")
	if port == "" || ollamaUrl == "" {
		log.Fatal("PORT and OLLAMA_URL must be set in .env file")
	}

	s, err := newSQLiteStorage()
	if err != nil {
		log.Fatal(err)
	}
	s.routes()

	serverAddr := fmt.Sprintf("http://%s:%s", host, port)
	if host == "" {
		serverAddr = fmt.Sprintf("http://%s:%s", "localhost", port)
	}
	log.Printf("Server running on %s\n", serverAddr)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), nil))
}

func createToken(lenght int) string {
	bytes := make([]byte, lenght)
	if _, err := rand.Read(bytes); err != nil {
		log.Println("Failed to generate session cookie: ", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

func hashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

func checkPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (s *SQLiteStorage) getUserBySessionToken(r *http.Request) (*User, error) {
	stCoockie, err := r.Cookie("session_token")
	if err != nil || stCoockie.Value == "" {
		return nil, fmt.Errorf("No session token found")
	}
	return s.selectUserBySessionToken(stCoockie.Value)
}

func loadLoginPage(w http.ResponseWriter, errMsg string) {
	tmpl, err := template.ParseFiles("html-content/login.html")
	if err != nil {
		http.Error(w, "Failed to load this page", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, errMsg); err != nil {
		http.Error(w, "Failed to load this page", http.StatusInternalServerError)
		return
	}
}
