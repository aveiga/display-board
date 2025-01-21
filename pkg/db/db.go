package db

import "time"

type Message struct {
	Username string
	ID       int
	Message  string
	Created  time.Time
	Modified time.Time
}

type Database struct{}

// messages is a package-level slice that stores all messages
var messages = []Message{}

// GetMessages retrieves all messages from the database
func (db *Database) GetMessages() ([]Message, error) {
	return messages, nil
}

// AddMessage adds a new message to the database
func (db *Database) AddMessage(message Message) ([]Message, error) {
	if message.Created.IsZero() {
		message.Created = time.Now()
	}

	messages = append(messages, message)
	return messages, nil
}

func (db *Database) DeleteMessage(id int) ([]Message, error) {
	for i, message := range messages {
		if message.ID == id {
			messages = append(messages[:i], messages[i+1:]...)
			return messages, nil
		}
	}
	return messages, nil
}

func (db *Database) UpdateMessage(id int, message Message) ([]Message, error) {
	message.Modified = time.Now()

	for i, msg := range messages {
		if msg.ID == id {
			messages[i] = message
			return messages, nil
		}
	}
	return messages, nil
}
