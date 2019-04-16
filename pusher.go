package pusher

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"golang.org/x/net/websocket"
)

// websocket connection manager instance
var CManager = &ConnManager{
	Online:      new(int32),
	connections: new(sync.Map),
}

// websocket connection manager
type ConnManager struct {
	// websocket connection number
	Online *int32
	// websocket connection
	connections *sync.Map
}

// add websocket connection
// online number + 1
func (m *ConnManager) Connected(k, v interface{}) {
	m.connections.Store(k, v)

	atomic.AddInt32(m.Online, 1)
}

// remove websocket connection by key
// online number - 1
func (m *ConnManager) DisConnected(k interface{}) {
	m.connections.Delete(k)

	atomic.AddInt32(m.Online, -1)
}

// get websocket connection by key
func (m *ConnManager) Get(k interface{}) (v interface{}, ok bool) {
	return m.connections.Load(k)
}

// iter websocket connections
func (m *ConnManager) Foreach(f func(k, v interface{})) {
	m.connections.Range(func(k, v interface{}) bool {
		f(k, v)
		return true
	})
}

// send message to one websocket connection
func (m *ConnManager) Send(k string, msg *Message) {
	v, ok := m.Get(k)
	if ok {
		if conn, ok := v.(*websocket.Conn); ok {
			if err := websocket.JSON.Send(conn, msg); err != nil {
				fmt.Println("Send msg error: ", err)
			}
		} else {
			fmt.Println("invalid type, expect *websocket.Conn")
		}
	} else {
		fmt.Println("connection not exist")
	}
}

// send message to multi websocket connections
func (m *ConnManager) SendMulti(keys []string, msg interface{}) {
	for _, k := range keys {
		v, ok := m.Get(k)
		if ok {
			if conn, ok := v.(*websocket.Conn); ok {
				if err := websocket.JSON.Send(conn, msg); err != nil {
					fmt.Println("Send msg error: ", err)
				}
			} else {
				fmt.Println("invalid type, expect *websocket.Conn")
			}
		} else {
			fmt.Println("connection not exist")
		}
	}
}

// broadcast message to all websocket connections otherwise own connection
func (m *ConnManager) Broadcast(conn *websocket.Conn, msg *Message) {
	m.Foreach(func(k, v interface{}) {
		if c, ok := v.(*websocket.Conn); ok && c != conn {
			if err := websocket.JSON.Send(c, msg); err != nil {
				fmt.Println("Send msg error: ", err)
			}
		}
	})
}

// websocket Handler
// usage: http.Handle("/websocket", websocket.Handler(pusher.Handler))
func Handler(conn *websocket.Conn) {
	// handle connected
	var userId string
	var err error
	if userId, err = doConnected(conn); err != nil {
		fmt.Println("Client connect error: ", err)
		return
	}

	fmt.Println("Client connected, userId: ", userId)

	for {
		msg := new(Message)

		if err := websocket.JSON.Receive(conn, msg); err != nil {
			fmt.Println("Can't receive, error: ", err)
			break
		}

		msg.UpdateAt = Timestamp()

		fmt.Println("Received from client: ", msg)

		// handle received message
		if err := doReceived(conn, msg); err != nil {
			fmt.Println("Received message error: ", err)
			break
		}
	}

	// handle disConnected
	if err := doDisConnected(userId, conn); err != nil {
		fmt.Println("Client disconnected error: ", err)
		return
	}

	fmt.Println("Client disconnected, userId: ", userId)
}

// handle on websocket connected
func doConnected(conn *websocket.Conn) (string, error) {
	var userId string
	var err error
	if userId, err = validConn(conn); err != nil {
		return "", err
	}

	// add websocket connection
	CManager.Connected(userId, conn)

	timestamp := Timestamp()
	var msg = &Message{
		MessageType: OnlineNotify,
		MediaType:   Text,
		From:        userId,
		CreateAt:    timestamp,
		UpdateAt:    timestamp,
	}
	CManager.Broadcast(conn, msg)

	// add to last message store
	LastMessage.Add(msg)

	// send recent message
	err = LastMessage.Foreach(func(msg *Message) {
		if err := websocket.JSON.Send(conn, msg); err != nil {
			fmt.Println("Send msg error: ", err)
		}
	})
	if err != nil {
		fmt.Println("send recent message error: ", err)
	}

	return userId, nil
}

// handle on websocket disConnected
func doDisConnected(userId string, conn *websocket.Conn) error {
	// del websocket connection
	CManager.DisConnected(userId)

	timestamp := Timestamp()
	var msg = &Message{
		MessageType: OfflineNotify,
		MediaType:   Text,
		From:        userId,
		CreateAt:    timestamp,
		UpdateAt:    timestamp,
	}
	CManager.Broadcast(conn, msg)

	// add to last message store
	LastMessage.Add(msg)

	return nil
}

// handle received message
func doReceived(conn *websocket.Conn, msg *Message) error {
	// todo parse message

	msg.UpdateAt = Timestamp()

	// broadcast message to all websocket connections
	CManager.Broadcast(conn, msg)

	// add to last message store
	LastMessage.Add(msg)

	return nil
}

// valid websocket connection
// return userId if connection validated
func validConn(conn *websocket.Conn) (string, error) {
	userId := conn.Request().URL.Query().Get("userId")
	if userId == "" {
		return "", errors.New("invalid client connected")
	}

	return userId, nil
}
