package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"
)

func (s *SQLiteStorage) routes() {
	http.HandleFunc("/", s.indexHandler)

	http.HandleFunc("/register", s.registerHandler)
	http.HandleFunc("/login", s.loginHandler)
	http.HandleFunc("/logout", s.logoutHandler)
	http.HandleFunc("/chat", s.chatHandler)
	http.HandleFunc("/message", s.messageHandler)

	http.HandleFunc("/json/user", s.userHandler)

	http.HandleFunc("/html/chat-history", s.chatHistoryHandler)
	http.HandleFunc("/html/chat", s.chatsHandler)
	http.HandleFunc("/html/models", s.modelsHandler)
}

func (s *SQLiteStorage) indexHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := s.getUserBySessionToken(r); err != nil {
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
	user, err := s.getUserBySessionToken(r)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	})
	if err := s.updateUserSessionTokenByName("", user.Name); err != nil {
		http.Error(w, "Failed to logout", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *SQLiteStorage) messageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var clientMessage *ClientMessage
	if err := json.NewDecoder(r.Body).Decode(&clientMessage); err != nil {
		http.Error(w, "Failed to decode message", http.StatusInternalServerError)
		return
	}
	user, err := s.getUserBySessionToken(r)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if clientMessage.ChatId == 0 {
		chat := newChat(user.Id, time.Now().Format(time.Layout))
		if err := s.insertChat(chat); err != nil {
			http.Error(w, "Failed to insert chat", http.StatusInternalServerError)
			return
		}
		clientMessage.ChatId = chat.Id
	}
	if err := s.insertMessage(newMessage(0, clientMessage.ChatId, clientMessage.Message, "user", time.Now().Format(time.Layout))); err != nil {
		http.Error(w, "Failed to insert message", http.StatusInternalServerError)
		return
	}
	var ollamaChat = newReqOllamaChat(clientMessage.Model, [][2]string{{"user", clientMessage.Message}}, false)
	b, err := json.Marshal(ollamaChat)
	if err != nil {
		http.Error(w, "Failed to marshal chat", http.StatusInternalServerError)
		return
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/chat", ollamaUrl), bytes.NewBuffer(b))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to send request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}
	w.Write(body)
}

func (s *SQLiteStorage) chatHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.getUserBySessionToken(r)
	if err != nil || user == nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.ServeFile(w, r, "html-content/chat.html")
}

/* --------------------------------- JSON --------------------------------- */

func (s *SQLiteStorage) userHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.getUserBySessionToken(r)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode user", http.StatusInternalServerError)
	}
}

/* --------------------------------- HTML --------------------------------- */

func (s *SQLiteStorage) chatHistoryHandler(w http.ResponseWriter, r *http.Request) {
	chatId, err := strconv.Atoi(r.URL.Query().Get("chat_id"))
	if err != nil {
		http.Error(w, "Chat ID not found", http.StatusBadRequest)
		return
	}
	chat, err := s.selectChatByID(chatId)
	if err != nil {
		http.Error(w, "Chat not found", http.StatusNotFound)
		return
	}
	user, err := s.getUserBySessionToken(r)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if chat.UserId != user.Id {
		http.Error(w, "User not allowed", http.StatusForbidden)
		return
	}
	messages, err := s.selectMessagesByChatID(chatId)
	if err != nil {
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}
	tmpl, err := template.ParseFiles("html-content/chat-history.html")
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, messages)
}

func (s *SQLiteStorage) chatsHandler(w http.ResponseWriter, r *http.Request) {
	chatTemplate := `<option class="default" value="" selected>Select a chat</option>`
	user, err := s.getUserBySessionToken(r)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
	}
	chats, err := s.selectChatsByUserID(user.Id)
	if err != nil {
		w.Write([]byte(chatTemplate))
		http.Error(w, "Failed to get chats", http.StatusInternalServerError)
	}
	chatTemplate += `{{ range . }}<option  value="{{ .Id }}">{{ .Name }}</option>{{ end }}`
	tmpl, err := template.New("chat").Parse(chatTemplate)
	if err != nil {
		w.Write([]byte(chatTemplate))
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, chats)
}

func (s *SQLiteStorage) modelsHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(ollamaUrl + "/api/tags")
	if err != nil {
		http.Error(w, "Failed to get models", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	var models *ResOllamaModels
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		http.Error(w, "Failed to decode models", http.StatusInternalServerError)
	}
	tmpl, err := template.ParseFiles("html-content/models.html")
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, models.Models)
}
