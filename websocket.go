package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

/*
	fake function to disable same origin control
*/
func CheckOrigin(r *http.Request) bool {
	return true
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     CheckOrigin,
	}
)

const (
	// Time allowed to write the file to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Poll file for changes with this period.
	filePeriod = 10 * time.Second
)

/*
	http handler for the websocket connection
*/
func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		fmt.Println(err)
		return
	}

	go writer(ws)
	reader(ws)
}

/*
	handles new messages from the client
*/
func reader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

/*
	handles sending data to the client and checking if the client is still alive
*/
func writer(ws *websocket.Conn) {
	lastError := ""
	pingTicker := time.NewTicker(pingPeriod)
	scoreUpdated := scoreKeeper.Subscribe()

	// after function end cleanup
	defer func() {
		pingTicker.Stop()
		ws.Close()
		scoreKeeper.Unsubscribe(scoreUpdated)
	}()
	for {
		select {
		// every time a new score comes in, encode in JSON and send to client
		case lastScore := <-scoreUpdated:
			var p []byte
			var err error

			p, err = json.Marshal(lastScore)

			if err != nil {
				if s := err.Error(); s != lastError {
					lastError = s
					p = []byte(lastError)
				}
			} else {
				lastError = ""
			}

			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.TextMessage, p); err != nil {
				return
			}
		// check if the client is still alive, if not exit
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
