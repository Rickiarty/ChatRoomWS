package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// 升級器配置
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允許所有來源
	},
}

// 客戶端管理
type Client struct {
	conn     *websocket.Conn
	sendBuff chan []byte
}

var (
	clients       = make(map[*Client]bool)
	broadcastBuff = make(chan []byte)
	clientLock    sync.Mutex
)

// 靜態檔案分享Handler
func staticFileHandler() {
	fs := http.FileServer(http.Dir("./wwwroot")) // 指定靜態文件目錄
	http.Handle("/", fs)
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
	// 升級HTTP到WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("failed to upgrade to WebSocket:", err)
		return
	}
	defer conn.Close()

	client := &Client{conn: conn, sendBuff: make(chan []byte)}
	clientLock.Lock()
	clients[client] = true
	clientLock.Unlock()

	defer func() {
		clientLock.Lock()
		delete(clients, client)
		clientLock.Unlock()
	}()

	go handleMessages(client)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("failed to read message:", err)
			break
		}
		broadcastBuff <- message
	}
}

func handleMessages(client *Client) {
	for message := range client.sendBuff {
		err := client.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			fmt.Println("failed to send message:", err)
			break
		}
	}
}

func messageDispatcher() {
	for {
		message := <-broadcastBuff
		clientLock.Lock()
		for client := range clients {
			select {
			case client.sendBuff <- message:
			default:
				close(client.sendBuff)
				delete(clients, client)
			}
		}
		clientLock.Unlock()
	}
}

// 獲取內網 IP 地址
func getInternalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("It can not get the internal network IP to this server:\n", err)
		return "unknown IP"
	}

	for _, addr := range addrs {
		// 檢查 IP 地址是否為有效的內網地址
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	return "unknown IP"
}
