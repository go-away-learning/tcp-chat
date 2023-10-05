package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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

var connections []connection

func main() {
	listener, err := net.Listen(listenerProtocol, listenerAddress)
	if err != nil {
		log.Fatalf("error listening on %s: %v\n", listenerAddress, err)
	}
	defer listener.Close()
	log.Printf("Listening on %s", listenerAddress)

	chat := make(chan message)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("error accepting connection: %v\n", err)
		}

		go handleConnection(conn, chat)
		go handleChat(chat)
	}
}

func handleConnection(conn net.Conn, chat chan message) {
	remoteAddr := conn.RemoteAddr().String()
	log.Printf("Serving %s\n", remoteAddr)
	defer conn.Close()

	conn.Write([]byte("Hi! What's your name: "))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Printf("error readind message: %v\n", err)
	}
	username := strings.TrimSpace(string(buf[:n]))
	connections = append(connections, connection{
		conn:     conn,
		username: username,
	})
	chat <- message{username: username, content: fmt.Sprintf("INFO: %s joined the chat.\n", username)}
	conn.Write([]byte("Welcome, " + username + "!\n"))

	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("error readind message: %v\n", err)
		}
		content := strings.TrimSpace(string(buf[:n]))
		chat <- message{username: username, content: fmt.Sprintf("%s: %s\n", username, content)}
	}
}

func handleChat(chat chan message) {
	for message := range chat {
		for _, connection := range connections {
			if connection.username == message.username {
				continue
			}
			connection.conn.Write([]byte(message.content))
		}
	}
}
