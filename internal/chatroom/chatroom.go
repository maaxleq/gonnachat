package chatroom

import (
	"errors"
	"fmt"
	"github.com/google/shlex"
	"log"
	"strings"
	"sync"
)

const msgBufferSize = 1000

type MessageType int

const (
	ChatMessage MessageType = iota
	ServerMessage
)

type Message struct {
	Type    MessageType
	Content string
	Author  string
}

func (msg Message) String() string {
	if msg.Type == ServerMessage {
		return fmt.Sprintf("# %s", msg.Content)
	} else {
		return fmt.Sprintf("%s: %s", msg.Author, msg.Content)
	}
}

type Command struct {
	Name string
	Args []string
}

func parseCommand(cmdString string) (Command, error) {
	cmdSlice, err := shlex.Split(cmdString)

	if err != nil || len(cmdSlice) < 1 {
		return Command{}, errors.New("command has no word")
	}

	return Command{
		Name: cmdSlice[0],
		Args: cmdSlice[1:],
	}, nil
}

type ChatRoom struct {
	senders  map[string]chan Message
	listener chan Message
	rwLock   sync.RWMutex
}

func NewChatRoom() *ChatRoom {
	senders := make(map[string]chan Message)
	listener := make(chan Message, msgBufferSize)

	chatRoom := ChatRoom{
		senders:  senders,
		listener: listener,
	}

	go func() {
		for {
			msg := <-chatRoom.listener
			go chatRoom.handleMessage(msg)
		}
	}()

	return &chatRoom
}

func (room *ChatRoom) handleCommand(originalMsg Message, cmd Command) {
	switch cmd.Name {
	case "tell":
		if len(cmd.Args) != 2 {
			room.tell(originalMsg.Author, Message{
				Type:    ServerMessage,
				Content: "malformed command",
			})
			return
		}
		if !room.userExists(cmd.Args[0]) {
			room.tell(originalMsg.Author, Message{
				Type:    ServerMessage,
				Content: fmt.Sprintf("user %s not found", cmd.Args[0]),
			})
			return
		}

		room.tell(cmd.Args[0], Message{
			Type:    ChatMessage,
			Content: cmd.Args[1],
			Author:  fmt.Sprintf("%s (whispering)", originalMsg.Author),
		})
	case "count":
		if len(cmd.Args) != 0 {
			room.tell(originalMsg.Author, Message{
				Type:    ServerMessage,
				Content: "malformed command",
			})
			return
		}

		room.tell(originalMsg.Author, Message{
			Type:    ServerMessage,
			Content: fmt.Sprintf("%d users connected", len(room.senders)),
		})
		return
	default:
		room.tell(originalMsg.Author, Message{
			Type:    ServerMessage,
			Content: fmt.Sprintf("command %s does not exist", cmd.Name),
		})
	}
}

func (room *ChatRoom) RLocked() bool {
	defer room.rwLock.RUnlock()
	return !room.rwLock.TryRLock()
}

func (room *ChatRoom) Locked() bool {
	defer room.rwLock.Unlock()
	return !room.rwLock.TryLock()
}

func (room *ChatRoom) handleMessage(msg Message) {
	if strings.HasPrefix(msg.Content, "/") {
		cmdString := strings.TrimPrefix(msg.Content, "/")
		cmd, err := parseCommand(cmdString)

		if err != nil {
			room.tell(msg.Author, Message{
				Type:    ServerMessage,
				Content: err.Error(),
			})
			return
		}

		room.handleCommand(msg, cmd)
	} else {
		room.broadcast(msg)
	}
}

func (room *ChatRoom) Say(name string, content string) {
	room.listener <- Message{
		Content: content,
		Author:  name,
		Type:    ChatMessage,
	}
}

func (room *ChatRoom) Subscribe(name string) chan Message {
	var sender chan Message

	func() {
		room.rwLock.Lock()
		defer room.rwLock.Unlock()
		sender = make(chan Message, msgBufferSize)
		room.senders[name] = sender
	}()

	room.broadcast(Message{
		Type:    ServerMessage,
		Content: fmt.Sprintf("%s entered the chat", name),
	})

	return sender
}

func (room *ChatRoom) Unsubscribe(name string) {
	func() {
		room.rwLock.Lock()
		defer room.rwLock.Unlock()
		delete(room.senders, name)
	}()

	room.broadcast(Message{
		Type:    ServerMessage,
		Content: fmt.Sprintf("%s left the chat", name),
	})
}

func (room *ChatRoom) tell(dest string, msg Message) {
	room.rwLock.RLock()
	defer room.rwLock.RUnlock()

	destChan, exists := room.senders[dest]
	if exists {
		destChan <- msg
	}
}

func (room *ChatRoom) IsNameAvailable(name string) bool {
	return !room.userExists(name)
}

func (room *ChatRoom) userExists(name string) bool {
	room.rwLock.RLock()
	defer room.rwLock.RUnlock()

	_, exists := room.senders[name]
	return exists
}

func (room *ChatRoom) broadcast(msg Message) {
	log.Print(msg)

	room.rwLock.RLock()
	defer room.rwLock.RUnlock()

	for _, sender := range room.senders {
		sender <- msg
	}
}
