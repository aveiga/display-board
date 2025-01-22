package db

import (
	"testing"
)

func TestDatabase_GetMessages(t *testing.T) {
	// Reset messages slice before each test
	messages = []Message{}

	db := &Database{}

	// Test empty database
	msgs, err := db.GetMessages()
	if err != nil {
		t.Errorf("GetMessages() error = %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("Expected empty message list, got %d messages", len(msgs))
	}
}

func TestDatabase_AddMessage(t *testing.T) {
	// Reset messages slice before each test
	messages = []Message{}

	db := &Database{}

	testMessage := Message{
		Username: "testuser",
		ID:       1,
		Message:  "Hello, World!",
	}

	msgs, err := db.AddMessage(testMessage)
	if err != nil {
		t.Errorf("AddMessage() error = %v", err)
	}

	if len(msgs) != 1 {
		t.Errorf("Expected 1 message, got %d messages", len(msgs))
	}

	if msgs[0].Username != testMessage.Username {
		t.Errorf("Expected username %s, got %s", testMessage.Username, msgs[0].Username)
	}

	if msgs[0].Created.IsZero() {
		t.Error("Created time should not be zero")
	}
}

func TestDatabase_DeleteMessage(t *testing.T) {
	// Reset messages slice before each test
	messages = []Message{}

	db := &Database{}

	// Add a test message first
	testMessage := Message{
		Username: "testuser",
		ID:       1,
		Message:  "Hello, World!",
	}

	_, err := db.AddMessage(testMessage)
	if err != nil {
		t.Errorf("Setup failed: %v", err)
	}

	// Test deletion
	msgs, err := db.DeleteMessage(1)
	if err != nil {
		t.Errorf("DeleteMessage() error = %v", err)
	}

	if len(msgs) != 0 {
		t.Errorf("Expected 0 messages after deletion, got %d", len(msgs))
	}
}

func TestDatabase_UpdateMessage(t *testing.T) {
	// Reset messages slice before each test
	messages = []Message{}

	db := &Database{}

	// Add initial message
	originalMessage := Message{
		Username: "testuser",
		ID:       1,
		Message:  "Original message",
	}

	_, err := db.AddMessage(originalMessage)
	if err != nil {
		t.Errorf("Setup failed: %v", err)
	}

	// Update the message
	updatedMessage := Message{
		Username: "testuser",
		ID:       1,
		Message:  "Updated message",
	}

	msgs, err := db.UpdateMessage(1, updatedMessage)
	if err != nil {
		t.Errorf("UpdateMessage() error = %v", err)
	}

	if len(msgs) != 1 {
		t.Errorf("Expected 1 message, got %d messages", len(msgs))
	}

	if msgs[0].Message != "Updated message" {
		t.Errorf("Expected message to be 'Updated message', got '%s'", msgs[0].Message)
	}

	if msgs[0].Modified.IsZero() {
		t.Error("Modified time should not be zero")
	}
}

func TestDatabase_MessageNotFound(t *testing.T) {
	// Reset messages slice before each test
	messages = []Message{}

	db := &Database{}

	// Try to delete non-existent message
	msgs, err := db.DeleteMessage(999)
	if err != nil {
		t.Errorf("DeleteMessage() error = %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(msgs))
	}

	// Try to update non-existent message
	updatedMessage := Message{
		Username: "testuser",
		ID:       999,
		Message:  "Updated message",
	}

	msgs, err = db.UpdateMessage(999, updatedMessage)
	if err != nil {
		t.Errorf("UpdateMessage() error = %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(msgs))
	}
}
