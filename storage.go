package main

import (
	"strings"
)

type Storage interface {
	insertUser(u *User) error
	selectUserByName(n string) (*User, error)
	selectBySessionToken(st string) (*User, error)
	updateUserSessionTokenByName(st string, n string) error

	insertChat(c *Chat) error
	selectChatByID(id string) (*Chat, error)
	selectChatsByUserID(userID string) ([]*Chat, error)
	updateChatById(c *Chat) error
	deleteChatById(id int) error

	insertMessage(m *Message) error
	selectMessageByID(id int) (*Message, error)
	selectMessagesByChatID(chatID int) ([]*Message, error)
	updateMessageById(m *Message) error
	deleteMessageById(id int) error
}

/* ----------------------------------------------- User ----------------------------------------------- */

func (s *SQLiteStorage) insertUser(u *User) error {
	prep, err := s.db.Prepare("INSERT INTO User (name, password, session_token) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer prep.Close()
	_, err = prep.Exec(u.Name, u.Password, u.SessionToken)
	return err
}

func (s *SQLiteStorage) selectUserByName(n string) (*User, error) {
	row := s.db.QueryRow("SELECT id, name, password, session_token FROM User WHERE name = ?", n)
	u := &User{}
	err := row.Scan(&u.Id, &u.Name, &u.Password, &u.SessionToken)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *SQLiteStorage) selectUserBySessionToken(st string) (*User, error) {
	row := s.db.QueryRow("SELECT id, name, password, session_token FROM User WHERE session_token = ?", st)
	u := &User{}
	err := row.Scan(&u.Id, &u.Name, &u.Password, &u.SessionToken)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *SQLiteStorage) updateUserSessionTokenByName(st string, n string) error {
	prep, err := s.db.Prepare("UPDATE User SET session_token = ? WHERE name = ?")
	if err != nil {
		return err
	}
	defer prep.Close()
	_, err = prep.Exec(st, n)
	return err
}

/* ----------------------------------------------- Chat ----------------------------------------------- */

func (s *SQLiteStorage) insertChat(c *Chat) error {
	prep, err := s.db.Prepare("INSERT INTO Chat (id, user_id, name, created_at) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer prep.Close()
	if _, err := prep.Exec(c.Id, c.UserId, c.Name, c.CreatedAt); err != nil {
		return err
	}
	return nil
}

func (s *SQLiteStorage) selectChatById(id string) (*Chat, error) {
	row := s.db.QueryRow("SELECT id, user_id, name, created_at FROM Chat WHERE id = ?", id)
	c := &Chat{}
	err := row.Scan(&c.Id, &c.UserId, &c.Name, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *SQLiteStorage) selectChatsByUserID(userID int) ([]*Chat, error) {
	rows, err := s.db.Query("SELECT id, user_id, name, created_at FROM Chat WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	chats := []*Chat{}
	for rows.Next() {
		c := &Chat{}
		if err := rows.Scan(&c.Id, &c.UserId, &c.Name, &c.CreatedAt); err != nil {
			return nil, err
		}
		chats = append(chats, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return chats, nil
}

func (s *SQLiteStorage) updateChatById(c *Chat) error {
	query := "UPDATE Chat SET"
	if c.Name != "" {
		query += " name = ?,"
	}
	if c.CreatedAt != "" {
		query += " created_at = ?,"
	}
	query = strings.TrimSuffix(query, ",")
	query += " WHERE id = ?"
	prep, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer prep.Close()
	if _, err := prep.Exec(c.Name, c.CreatedAt, c.Id); err != nil {
		return err
	}
	return nil
}

func (s *SQLiteStorage) deleteChatById(id string) error {
	prep, err := s.db.Prepare("DELETE FROM Chat WHERE id = ?")
	if err != nil {
		return err
	}
	defer prep.Close()
	if _, err := prep.Exec(id); err != nil {
		return err
	}
	return nil
}

/* ----------------------------------------------- Message ----------------------------------------------- */

func (s *SQLiteStorage) insertMessage(m *Message) error {
	prep, err := s.db.Prepare("INSERT INTO Message (chat_id, content, role, created_at) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer prep.Close()
	if _, err := prep.Exec(m.ChatId, m.Content, m.Role, m.CreatedAt); err != nil {
		return err
	}
	return nil
}

func (s *SQLiteStorage) selectMessageByID(id int) (*Message, error) {
	prep, err := s.db.Prepare("SELECT id, chat_id, content, role, created_at FROM Message WHERE id = ?")
	if err != nil {
		return nil, err
	}
	defer prep.Close()
	var m Message
	if err := prep.QueryRow(id).Scan(&m.Id, &m.ChatId, &m.Content, &m.Role, &m.CreatedAt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *SQLiteStorage) selectMessagesByChatID(chatID string) ([]*Message, error) {
	prep, err := s.db.Prepare("SELECT id, chat_id, content, role, created_at FROM Message WHERE chat_id = ?")
	if err != nil {
		return nil, err
	}
	defer prep.Close()
	rows, err := prep.Query(chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []*Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.Id, &m.ChatId, &m.Content, &m.Role, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}

func (s *SQLiteStorage) updateMessageById(m *Message) error {
	prep, err := s.db.Prepare("UPDATE Message SET content = ?, role = ?, created_at = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer prep.Close()
	if _, err := prep.Exec(m.Content, m.Role, m.CreatedAt, m.Id); err != nil {
		return err
	}
	return nil
}

func (s *SQLiteStorage) deleteMessageById(id int) error {
	prep, err := s.db.Prepare("DELETE FROM Message WHERE id = ?")
	if err != nil {
		return err
	}
	defer prep.Close()
	if _, err := prep.Exec(id); err != nil {
		return err
	}
	return nil
}
