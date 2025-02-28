package main

var createTableQuerys = []string{
	`CREATE TABLE IF NOT EXISTS User (
		name TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL UNIQUE,
		session_token TEXT UNIQUE
	)`,
	`CREATE TABLE IF NOT EXISTS Chat (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		create_at TEXT NOT NULL,
		FOREIGN KEY(user_id) REFERENCES User(id)
	)`,
	`CREATE TABLE IF NOT EXISTS Message (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		role TEXT NOT NULL,
		create_at TEXT NOT NULL,
		FOREIGN KEY(chat_id) REFERENCES Chat(id)
	)`,
}
