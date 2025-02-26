package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

func (s *SQLiteStorage) routes() {
	http.HandleFunc("/register", s.registerHandler)
	http.HandleFunc("/login", s.loginHandler)
	http.HandleFunc("/logout", s.logoutHandler)
	http.HandleFunc("/htmx/chat-list", s.chatHandler)
	http.Handle("/", http.FileServer(http.Dir("public")))
}

func (s *SQLiteStorage) registerHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("username")
	password := r.FormValue("password")
	if len(name) < 8 || len(password) < 8 {
		http.Error(w, "Username and password must be at least 8 characters long", http.StatusBadRequest)
		return
	}
	passHash, err := hashPassword(password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	user := newUser(name, passHash, "")
	if err := s.insertUser(user); err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/login?username=%s&password=%s", name, password), http.StatusFound)
}

func (s *SQLiteStorage) loginHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("username")
	password := r.FormValue("password")
	user, err := s.selectUserByName(name)
	if err != nil || !checkPasswordHash(password, user.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	user.SessionToken = createToken(32)
	if err := s.updateUserSessionTokenByName(user.Name, user.SessionToken); err != nil {
		http.Error(w, "Failed to login", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    user.SessionToken,
		Expires:  time.Now().Add(24 * time.Hour),
		Path:     "/",
		HttpOnly: true,
	})
	http.Redirect(w, r, "/chat", http.StatusFound)
}

func (s *SQLiteStorage) logoutHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.getUserBySessionToken(r)
	if err != nil {
		log.Println(err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	})
	if err := s.updateUserSessionTokenByName(user.Name, ""); err != nil {
		http.Error(w, "Failed to logout", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

/* ----------------------------------------------- Chat ----------------------------------------------- */

func (s *SQLiteStorage) chatHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.getUserBySessionToken(r)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
	}
	chats, err := s.selectChatsByUserID(user.Id)
	if err != nil {
		http.Error(w, "Failed to get chats", http.StatusInternalServerError)
		return
	}
	chatTemplate := `{{ range . }} <div class="chat" id="chat-{ .Id }">{ .Name }</div>{{ end }}`
	tmpl, err := template.New("chat").Parse(chatTemplate)
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, chats); err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}
}
