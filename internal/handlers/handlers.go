package handlers

import (
	"context"
	"fmt"
	"github.com/goombaio/namegenerator"
	"gonnachat/internal/chatroom"
	"log"
	"net/http"
	"nhooyr.io/websocket"
	"time"
)

var room = chatroom.NewChatRoom()
var nameGenerator = namegenerator.NewNameGenerator(time.Now().UTC().UnixNano())

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
		name = nameGenerator.Generate()
	} else {
		if !room.IsNameAvailable(name) {
			oldName := name
			name = nameGenerator.Generate()
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
