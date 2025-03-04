package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var (
	host      string
	port      string
	ollamaUrl string
	proxy     *httputil.ReverseProxy
	client    = &http.Client{}
)

func main() {
	godotenv.Load(".env")
	host = os.Getenv("HOST")
	port = os.Getenv("PORT")
	ollamaUrl = os.Getenv("OLLAMA_URL")
	if port == "" || ollamaUrl == "" {
		log.Fatal("PORT and OLLAMA_URL must be set in .env file")
	}

	targetUrl, err := url.Parse(ollamaUrl)
	if err != nil {
		log.Fatal("Failed to parse ollama-url.")
	}
	proxy = httputil.NewSingleHostReverseProxy(targetUrl)

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

func logHttpErr(w http.ResponseWriter, msg string, code int, err error) {
	if err != nil {
		log.Printf("%s: %v", msg, err)
	} else {
		log.Println(msg)
	}
	http.Error(w, msg, code)
}

type authHandler func(http.ResponseWriter, *http.Request)

func (s *SQLiteStorage) authorize(f authHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := s.getUserBySessionToken(r)
		if err != nil || user == nil {
			logHttpErr(w, "User not found", http.StatusNotFound, err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		f(w, r)
	}
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
