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

type MsgType string

const (
	CONNECT    MsgType = "CONNECT"
	COMMAND    MsgType = "COMMAND"
	DISCONNECT MsgType = "DISCONNECT"
)

type WsMsg struct {
	Type MsgType     `json:"type"`
	Data interface{} `json:"data"`
}

type ConnectInfo struct {
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

	var sshClient *ssh.Client

	defer func() {
		if sshClient != nil {
			fmt.Println("sshConn.Close()")
			sshClient.Close()
		}
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket Read Error:", err)
			break
		}

		var wsMsg WsMsg
		err = json.Unmarshal(message, &wsMsg)
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("Invalid message format"))
			log.Println("Unmarshal error:", err)
			continue
		}

		switch wsMsg.Type {
		case CONNECT:
			if sshClient != nil {
				conn.WriteMessage(websocket.TextMessage, []byte("Connection is already existing"))
				continue
			}

			connectInfoMap, ok := wsMsg.Data.(map[string]interface{})
			if !ok {
				conn.WriteMessage(websocket.TextMessage, []byte("Invalid connection information format"))
				continue
			}

			connectInfo := ConnectInfo{
				Host:     connectInfoMap["host"].(string),
				Username: connectInfoMap["username"].(string),
				Password: connectInfoMap["password"].(string),
			}

			sshClient, err = connectSSH(&connectInfo)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Connection error: %v", err)))
				continue
			}
			conn.WriteMessage(websocket.TextMessage, []byte("Connected to SSH server"))

		case COMMAND:
			if sshClient == nil {
				conn.WriteMessage(websocket.TextMessage, []byte("SSH client is not connected"))
				continue
			}

			command, ok := wsMsg.Data.(string)
			if !ok {
				conn.WriteMessage(websocket.TextMessage, []byte("Invalid command format"))
				continue
			}

			result, err := runCommand(sshClient, command)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Command error: %v", err)))
			} else {
				conn.WriteMessage(websocket.TextMessage, []byte(result))
			}
		}
	}
}

func connectSSH(connectInfo *ConnectInfo) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: connectInfo.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(connectInfo.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	client, err := ssh.Dial("tcp", connectInfo.Host, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return client, nil
}

func runCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b

	if err := session.Run(command); err != nil {
		return "", err
	}

	return b.String(), nil
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", sshHandler)

	log.Println("Server is running!: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
