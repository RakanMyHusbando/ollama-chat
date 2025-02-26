package main

type Storage interface {
	insertUser(u *User) error
	selectUserByName(n string) (*User, error)
	selectBySessionToken(st string) (*User, error)
	updateUserSessionTokenByName(st string, n string) error

	insertChat(c *Chat) error
	selectChatByID(id int) (*Chat, error)
	selectChatsByUserID(userID int) ([]*Chat, error)
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
	prep, err := s.Prepare("INSERT INTO User (name, password, session_token) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer prep.Close()
	_, err = prep.Exec(u.Name, u.Password, u.SessionToken)
	return err
}

func (s *SQLiteStorage) selectUserByName(n string) (*User, error) {
	row := s.QueryRow("SELECT name, password, session_token FROM User WHERE name = ?", n)
	u := &User{}
	err := row.Scan(&u.Name, &u.Password, &u.SessionToken)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *SQLiteStorage) selectUserBySessionToken(st string) (*User, error) {
	row := s.QueryRow("SELECT name, password, session_token FROM User WHERE session_token = ?", st)
	u := &User{}
	err := row.Scan(&u.Name, &u.Password, &u.SessionToken)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *SQLiteStorage) updateUserSessionTokenByName(st string, n string) error {
	prep, err := s.Prepare("UPDATE User SET session_token = ? WHERE name = ?")
	if err != nil {
		return err
	}
	defer prep.Close()
	_, err = prep.Exec(st, n)
	return err
}

/* ----------------------------------------------- Chat ----------------------------------------------- */

func (s *SQLiteStorage) insertChat(c *Chat) error {
	return nil
}

func (s *SQLiteStorage) selectChatByID(id int) (*Chat, error) {
	return nil, nil
}

func (s *SQLiteStorage) selectChatsByUserID(userID int) ([]*Chat, error) {
	return nil, nil
}

func (s *SQLiteStorage) updateChatById(c *Chat) error {
	return nil
}

func (s *SQLiteStorage) deleteChatById(id int) error {
	return nil
}

/* ----------------------------------------------- Message ----------------------------------------------- */

func (s *SQLiteStorage) insertMessage(m *Message) error {
	return nil
}

func (s *SQLiteStorage) selectMessageByID(id int) (*Message, error) {
	return nil, nil
}

func (s *SQLiteStorage) selectMessagesByChatID(chatID int) ([]*Message, error) {
	return nil, nil
}

func (s *SQLiteStorage) updateMessageById(m *Message) error {
	return nil
}

func (s *SQLiteStorage) deleteMessageById(id int) error {
	return nil
}
