package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

const (
	listenerProtocol = "tcp4"
	listenerAddress  = "127.0.0.1:3333"
)

type connection struct {
	conn     net.Conn
	username string
}

type message struct {
	username string
	content  string
}

var (
	connections []connection
	chat        = make(chan message)
	connMutex   sync.Mutex
)

func main() {
	listener, err := net.Listen(listenerProtocol, listenerAddress)
	if err != nil {
		log.Fatalf("error listening on %s: %v\n", listenerAddress, err)
	}
	defer listener.Close()
	log.Printf("Listening on %s", listenerAddress)

	go handleChat()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("error accepting connection: %v\n", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	conn.Write([]byte("Hi! What's your name: "))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Printf("error reading message: %v\n", err)
		return
	}
	username := strings.TrimSpace(string(buf[:n]))

	connMutex.Lock()
	connections = append(connections, connection{
		conn:     conn,
		username: username,
	})
	connMutex.Unlock()

	chat <- message{username: username, content: fmt.Sprintf("INFO: %s joined the chat.\n", username)}
	conn.Write([]byte("Welcome, " + username + "!\n"))

	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("error reading message: %v\n", err)
			break
		}
		content := strings.TrimSpace(string(buf[:n]))
		chat <- message{username: username, content: fmt.Sprintf("%s: %s\n", username, content)}
	}
}

func handleChat() {
	for msg := range chat {
		connMutex.Lock()
		for _, connection := range connections {
			if connection.username == msg.username {
				continue
			}
			_, err := connection.conn.Write([]byte(msg.content))
			if err != nil {
				log.Printf("error writing message to %s: %v\n", connection.username, err)
			}
		}
		connMutex.Unlock()
	}
}
