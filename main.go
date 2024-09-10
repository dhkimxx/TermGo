package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

type SshInfo struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func sshHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade Error:", err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket Read Error:", err)
			break
		}

		var sshInfo SshInfo
		err = json.Unmarshal(message, &sshInfo)

		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("Invalid message format"))
			continue
		}

		result, err := connectSSH(&sshInfo)
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error: %v", err)))
		} else {
			conn.WriteMessage(websocket.TextMessage, []byte(result))
		}
	}
}

func connectSSH(sshInfo *SshInfo) (string, error) {
	config := &ssh.ClientConfig{
		User: sshInfo.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshInfo.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	conn, err := ssh.Dial("tcp", sshInfo.Host, config)
	if err != nil {
		return "", fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("ls -a"); err != nil {
		return "", fmt.Errorf("failed to run command: %w", err)
	}

	fmt.Printf("Return: %s", b.String())

	return b.String(), nil
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", sshHandler)

	log.Println("Server is running!: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
