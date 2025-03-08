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
	http.HandleFunc("/chat", s.chatHandler)

	http.HandleFunc("/register", s.registerHandler)
	http.HandleFunc("/login", s.loginHandler)
	http.HandleFunc("/logout", s.authorize(s.logoutHandler))

	http.HandleFunc("/ollama/", ollamaHandler)

	http.HandleFunc("/api/chat", s.authorize(s.apiChatHandler))
	http.HandleFunc("/api/message", s.authorize(s.apiMessageHandler))

	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js/"))))
}

func ollamaHandler(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = strings.Replace(r.URL.Path, "/ollama/", "/api/", 1)
	proxy.ServeHTTP(w, r)
}

func ollamaStreamHandler(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = strings.Replace(r.URL.Path, "/ollama/stream/", "/api/", 1)
	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Del("Content-Length")
		resp.Header.Del("Content-Encoding")
		return nil
	}
	flusher := w.(http.Flusher)
	w.Header().Set("Transfer-Encoding", "chunked")
	proxy.ServeHTTP(w, r)
	flusher.Flush()

}

func (s *SQLiteStorage) indexHandler(w http.ResponseWriter, r *http.Request) {
	if user, err := s.getUserBySessionToken(r); err != nil || user == nil {
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
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    strconv.Itoa(user.Id),
		Expires:  time.Now().Add(24 * time.Hour),
		Path:     "/",
		HttpOnly: false,
	})
	http.Redirect(w, r, "/chat", http.StatusFound)
}

func (s *SQLiteStorage) logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	})
	if err := s.updateUserSessionTokenByName("", s.user.Name); err != nil {
		logHttpErr(w, "Failed to logout", http.StatusInternalServerError, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *SQLiteStorage) chatHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	http.ServeFile(w, r, "html/chat.html")
}

/* --------------------------------- API --------------------------------- */

func (s *SQLiteStorage) apiChatHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getChatHandler(w, r)
	case http.MethodPost:
		s.postChatHandler(w, r)
	default:
		logHttpErr(w, "Method not allowed", http.StatusMethodNotAllowed, nil)
	}
}

func (s *SQLiteStorage) getChatHandler(w http.ResponseWriter, r *http.Request) {
	qMsg := r.URL.Query().Get("msg")
	qChatId := r.URL.Query().Get("id")
	chats := []*Chat{}
	if qChatId != "" {
		if chat, err := s.selectChatById(qChatId); err == nil {
			chats = append(chats, chat)
		}
	}
	var err error
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

func (s *SQLiteStorage) apiMessageHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.postMessageHandler(w, r)
	default:
		logHttpErr(w, "Method not allowed", http.StatusMethodNotAllowed, nil)
	}
}

func (s *SQLiteStorage) postMessageHandler(w http.ResponseWriter, r *http.Request) {
	var message *Message
	if err := json.NewDecoder(r.Body).Decode(message); err != nil {
		logHttpErr(w, "Failed to decode message", http.StatusBadRequest, err)
		return
	}
	if err := s.insertMessage(message); err != nil {
		logHttpErr(w, "Failed to insert message", http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
