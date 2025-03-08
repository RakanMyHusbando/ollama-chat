package main

import (
	"database/sql"
)

type SQLiteStorage struct {
	user *User
	db   *sql.DB
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
	return &SQLiteStorage{db: db}, nil
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
	Id        string     `json:"id"`
	Name      string     `json:"name"`
	UserId    int        `json:"user_id"`
	CreatedAt string     `json:"created_at"`
	Messages  []*Message `json:"messages"`
}

func newChat(id string, name string, userId int, timeStr string, messages []*Message) *Chat {
	return &Chat{
		Id:        id,
		Name:      name,
		UserId:    userId,
		CreatedAt: timeStr,
		Messages:  messages,
	}
}

type Message struct {
	Id        int    `json:"id"`
	ChatId    string `json:"chat_id"`
	Content   string `json:"content"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

func newMessage(id int, chatId, content, role, createdAt string) *Message {
	return &Message{
		Id:        id,
		ChatId:    chatId,
		Content:   content,
		Role:      role,
		CreatedAt: createdAt,
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
