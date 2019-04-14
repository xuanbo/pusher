package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/xuanbo/pusher"
	"golang.org/x/net/websocket"
)

func main() {
	http.Handle("/websocket", websocket.Handler(pusher.Handler))

	// Index
	http.Handle("/", http.FileServer(http.Dir("./")))

	// admin
	http.HandleFunc("/admin", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = fmt.Fprintln(writer, *pusher.CManager.Online)
	})

	fmt.Println("Listen and serve on port 8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Listen and serve error: ", err)
	}
}
