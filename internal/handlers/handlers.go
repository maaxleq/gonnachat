package handlers

import (
	"context"
	"fmt"
	"gonnachat/internal/chatroom"
	"gonnachat/internal/namegenerator"
	"log"
	"net/http"
	"nhooyr.io/websocket"
)

var room = chatroom.NewChatRoom()

func MutexState(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("Write locked : %t / Read locked : %t", room.Locked(), room.RLocked())))
}

func WSChat(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade websocket", http.StatusInternalServerError)
		return
	}

	defer func() {
		_ = conn.Close(websocket.StatusInternalError, "The connection was closed due to an internal error")
	}()

	// Create a context that will be cancelled when there's an error
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel() // Make sure all paths cancel the context to avoid context leak

	name := r.URL.Query().Get("name")
	if name == "" {
		name = namegenerator.Generate()
	} else {
		if !room.IsNameAvailable(name) {
			oldName := name
			name = namegenerator.Generate()
			err := conn.Write(ctx, websocket.MessageText, []byte(fmt.Sprintf("# name %s is already taken, taking %s instead", oldName, name)))
			if err != nil {
				log.Printf("failed to write message: %v", err)
				cancel()
				return
			}
		}
	}
	defer room.Unsubscribe(name)

	in := room.Subscribe(name)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-in:
				err := conn.Write(ctx, websocket.MessageText, []byte(msg.String()))
				if err != nil {
					log.Printf("failed to write message: %v", err)
					cancel()
					return
				}
			}
		}
	}()

	for {
		_, msg, err := conn.Read(ctx)
		if err != nil {
			log.Printf("failed to read message: %v", err)
			cancel()
			return
		}

		room.Say(name, string(msg))
	}
}
