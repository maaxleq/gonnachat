package chatroom

import (
	"fmt"
	"testing"
	"time"
)

func TestChatRoom(t *testing.T) {
	room := NewChatRoom()

	t.Run("TestSubscribeAndSay", func(t *testing.T) {
		user := "testuser"
		message := "Hello, chatroom!"

		subscription := room.Subscribe(user)

		go room.Say(user, message)

		// First message should be server message notifying user joined.
		select {
		case msg := <-subscription:
			expectedContent := fmt.Sprintf("%s entered the chat", user)
			if msg.Type != ServerMessage || msg.Content != expectedContent {
				t.Errorf("Expected server message '%s', got '%s'", expectedContent, msg.Content)
			}
		case <-time.After(time.Second):
			t.Errorf("Did not receive a message within the timeout")
		}

		// Second message should be user's message.
		select {
		case msg := <-subscription:
			if msg.Author != user || msg.Content != message {
				t.Errorf("Expected message '%s' from '%s', got '%s' from '%s'", message, user, msg.Content, msg.Author)
			}
		case <-time.After(time.Second):
			t.Errorf("Did not receive a message within the timeout")
		}
	})

	t.Run("TestUnsubscribe", func(t *testing.T) {
		user := "testuser2"

		room.Subscribe(user)
		room.Unsubscribe(user)

		if room.userExists(user) {
			t.Errorf("User '%s' still exists after unsubscribe", user)
		}
	})

	t.Run("TestIsNameAvailable", func(t *testing.T) {
		user := "testuser3"

		if !room.IsNameAvailable(user) {
			t.Errorf("Name '%s' is not available before subscribe", user)
		}

		room.Subscribe(user)

		if room.IsNameAvailable(user) {
			t.Errorf("Name '%s' is available after subscribe", user)
		}
	})
}

// Additional tests should be added to test chat commands and server messages.
