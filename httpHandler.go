package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func serveFile(file string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, file)
	}
}

func (s *SQLiteStorage) routes() {
	http.HandleFunc("/", s.makeHandler(s.indexHandler, false))
	http.HandleFunc("/chat", s.makeHandler(serveFile("./html/chat.html"), true))

	http.HandleFunc("/register", s.makeHandler(s.registerHandler, false))
	http.HandleFunc("/login", s.makeHandler(s.loginHandler, false))
	http.HandleFunc("/logout", s.makeHandler(s.logoutHandler, true))

	http.HandleFunc("/api/chat", s.makeHandler(s.apiChatHandler, true))
	http.HandleFunc("/api/message", s.makeHandler(s.apiMessageHandler, true))

	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js/"))))
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
		Path:     "/",
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		Path:     "/",
		HttpOnly: false,
	})
	if err := s.updateUserSessionTokenByName("", s.user.Name); err != nil {
		logHttpErr(w, "Failed to logout", http.StatusInternalServerError, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
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
	if err := json.NewDecoder(r.Body).Decode(&chat); err != nil {
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
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		logHttpErr(w, "Failed to decode message", http.StatusBadRequest, err)
		return
	}
	if err := s.insertMessage(message); err != nil {
		logHttpErr(w, "Failed to insert message", http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
