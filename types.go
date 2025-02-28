package main

import (
	"database/sql"
	"log"
	"time"
)

type SQLiteStorage struct {
	*sql.DB
}

func newSQLiteStorage() (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		return nil, err
	}
	for _, query := range createTableQuerys {
		if _, err = db.Exec(query); err != nil {
			return nil, err
		}
	}
	return &SQLiteStorage{db}, nil
}

type User struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Password     string `json:"password"`
	SessionToken string `json:"session_token"`
}

// newUser creates a new user without an ID.
func newUser(name, password, sessionToken string) *User {
	return &User{
		Name:         name,
		Password:     password,
		SessionToken: sessionToken,
	}
}

type Chat struct {
	Id       int       `json:"id"`
	UserId   int       `json:"user_id"`
	CreateAt time.Time `json:"create_at"`
}

// newChat creates a new chat without an ID.
func newChat(userId int, timeStr string) *Chat {
	t, err := time.Parse(time.Layout, timeStr)
	if err != nil {
		log.Println("Failed to parse time: ", err)
	}
	t = time.Time{}
	return &Chat{
		UserId:   userId,
		CreateAt: t,
	}
}

type Message struct {
	Id       int       `json:"id"`
	ChatId   int       `json:"chat_id"`
	Message  string    `json:"message"`
	Role     string    `json:"role"`
	CreateAt time.Time `json:"create_at"`
}

func newMessage(Id, chatId int, message, role, timeStr string) *Message {
	t, err := time.Parse(time.Layout, timeStr)
	if err != nil {
		log.Println("Failed to parse time: ", err)
	}
	t = time.Time{}
	return &Message{
		Id:       Id,
		ChatId:   chatId,
		Message:  message,
		Role:     role,
		CreateAt: t,
	}
}

type ResOllamaModels struct {
	Models []*ResOllamaModel `json:"models"`
}

type ResOllamaModel struct {
	Name      string                 `json:"name"`
	Model     string                 `json:"model"`
	Size      int                    `json:"size"`
	Digest    string                 `json:"digest"`
	Details   *ResOllamaModelDetails `json:"details"`
	ExpiresAt string                 `json:"expires_at"`
	SizeVram  int                    `json:"size_vram"`
}

type ResOllamaModelDetails struct {
	ParentModel       string   `json:"parent_model"`
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

type ReqOllamaChat struct {
	Model    string                   `json:"model"`
	Messages []*ReqOllamaChatMessages `json:"messages"`
	Stream   bool                     `json:"stream"`
}

type ReqOllamaChatMessages struct {
	Role    string `json:"role"`
	Message string `json:"message"`
}

// messages[i][0] is the role and messages[i][1] is the message.
func newReqOllamaChat(model string, messages [][2]string, stream bool) *ReqOllamaChat {
	var ollamaMessages []*ReqOllamaChatMessages
	for _, m := range messages {
		ollamaMessages = append(ollamaMessages, &ReqOllamaChatMessages{
			Role:    m[0],
			Message: m[1],
		})
	}
	return &ReqOllamaChat{
		Model:    model,
		Messages: ollamaMessages,
		Stream:   stream,
	}
}

type ClientMessage struct {
	Model   string `json:"model"`
	ChatId  int    `json:"chat_id"`
	Message string `json:"message"`
}
