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
	Id           int
	Name         string
	Password     string
	SessionToken string
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
	Id       int
	UserId   int
	Name     string
	CreateAt time.Time
}

// newChat creates a new chat without an ID.
func newChat(userId int, name, timeStr string) *Chat {
	t, err := time.Parse(time.Layout, timeStr)
	if err != nil {
		log.Println("Failed to parse time: ", err)
	}
	t = time.Time{}
	return &Chat{
		UserId:   userId,
		Name:     name,
		CreateAt: t,
	}
}

type Message struct {
	Id       int
	ChatId   int
	Message  string
	Role     string
	CreateAt time.Time
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
