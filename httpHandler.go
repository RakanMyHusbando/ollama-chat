package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (s *SQLiteStorage) routes() {
	http.HandleFunc("/", s.indexHandler)
	http.HandleFunc("/chat", s.authorize(s.chatHandler))

	http.HandleFunc("/register", s.registerHandler)
	http.HandleFunc("/login", s.loginHandler)
	http.HandleFunc("/logout", s.authorize(s.logoutHandler))

	http.HandleFunc("/ollama/", ollamaHandler)

	http.HandleFunc("/api/chat", s.authorize(s.apiChatHandler))

	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js/"))))
}

func ollamaHandler(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = strings.Replace(r.URL.Path, "/ollama/", "/api/", 1)
	proxy.ServeHTTP(w, r)
}

func (s *SQLiteStorage) indexHandler(w http.ResponseWriter, r *http.Request) {
	if user, err := s.getUserBySessionToken(r); err != nil && user != nil {
		loadLoginPage(w, "")
		return
	}
	http.Redirect(w, r, "/chat", http.StatusFound)
}

func (s *SQLiteStorage) registerHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("username")
	password := r.FormValue("password")
	if len(name) < 8 || len(password) < 8 {
		loadLoginPage(w, "Username and password must be at least 8 characters long")
		return
	}
	passHash, err := hashPassword(password)
	if err != nil {
		loadLoginPage(w, "Failed to hash password")
		return
	}
	user := newUser(name, passHash, "")
	if err := s.insertUser(user); err != nil {
		loadLoginPage(w, "User already exists")
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/login?username=%s&password=%s", name, password), http.StatusFound)
}

func (s *SQLiteStorage) loginHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("username")
	password := r.FormValue("password")
	user, err := s.selectUserByName(name)
	if err != nil || !checkPasswordHash(password, user.Password) {
		loadLoginPage(w, "Invalid credentials")
		return
	}
	user.SessionToken = createToken(32)
	if err := s.updateUserSessionTokenByName(user.SessionToken, user.Name); err != nil {
		loadLoginPage(w, "Failed to update session token")
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
	user, _ := s.getUserBySessionToken(r)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	})
	if err := s.updateUserSessionTokenByName("", user.Name); err != nil {
		logHttpErr(w, "Failed to logout", http.StatusInternalServerError, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *SQLiteStorage) chatHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "html/chat.html")
}

/* --------------------------------- API --------------------------------- */

func (s *SQLiteStorage) apiChatHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getChatHandler(w, r)
	case http.MethodPost:
		s.postChatHandler(w, r)
	}
}

func (s *SQLiteStorage) getChatHandler(w http.ResponseWriter, r *http.Request) {
	qMsg := r.URL.Query().Get("msg")
	qChatId := r.URL.Query().Get("chatId")
	_, err := s.getUserBySessionToken(r)
	if err != nil {
		logHttpErr(w, "User not found", http.StatusNotFound, err)
		return
	}
	chats := []*Chat{}
	if chatId, err := strconv.Atoi(qChatId); qChatId != "" && err == nil {
		if chat, err := s.selectChatByID(chatId); err == nil {
			chats = append(chats, chat)
		}
	}
	if qMsg == "true" {
		for _, chat := range chats {
			chat.Messages, err = s.selectMessagesByChatID(chat.Id)
			if err != nil {
				logHttpErr(w, "Failed to get messages", http.StatusInternalServerError, err)
				return
			}
		}
	}
	b, err := json.Marshal(chats)
	if err != nil {
		logHttpErr(w, "Failed to marshal chats", http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (s *SQLiteStorage) postChatHandler(w http.ResponseWriter, r *http.Request) {
	var chat *Chat
	if err := json.NewDecoder(r.Body).Decode(chat); err != nil {
		logHttpErr(w, "Failed to decode chat", http.StatusBadRequest, err)
		return
	}
	if err := s.insertChat(chat); err != nil {
		logHttpErr(w, "Failed to insert chat", http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
