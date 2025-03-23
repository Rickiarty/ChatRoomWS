package main

import (
	"fmt"
	"net/http"
)

func main() {
	// 獲取並顯示內網 IP
	internalIP := getInternalIP()

	// 啟動靜態檔案分享
	staticFileHandler()

	// WebSocket連線處理
	http.HandleFunc("/ws", handleConnection)

	go messageDispatcher()

	port := "8080"
	fmt.Printf("Open browser and then navigate the page:\nhttp://%s:%s/chatroom.html\n", internalIP, port)
	fmt.Printf("\n( [Ctrl + C] to exit )\n")
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("The server-side program failed to run.", err)
	}
}
