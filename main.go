package main

import (
	"log"
	"net/http"
	"strings"

	"go.bug.st/serial.v1"
)

var (
	scoreKeeper = ScoreKeeperBuilder()
)

/*
	This function reads from a serial device and relays parsed messages to the channel given in input
	Serial configurations will need to be tweaked accordingly and maybe read from a configuration file
*/
func readFromSerial(messages chan score) {
	mode := &serial.Mode{
		BaudRate: 115200,
	}
	port, err := serial.Open("/dev/ttyp7", mode)
	if err != nil {
		log.Fatal(err)
	}

	buff := make([]byte, 100)

	defer port.Close()

	for {
		n, err := port.Read(buff)
		if err != nil {
			log.Fatal(err)
			break
		}

		if n != 0 {
			messages <- parseMessage(string(buff[:n]))
		}

	}
}

/*
	Parse an input string and returns the corresponding score
*/
func parseMessage(message string) score {
	message = strings.TrimRight(message, "\r\n")
	s := score{Red: 0, Blue: 0}
	if message == "GOAL_RED_1" {
		s.Red++
	}
	if message == "GOAL_BLUE_1" {
		s.Blue++
	}
	return s
}

func startWebserver() {
	// serve static files from the client folder
	fs := http.FileServer(http.Dir("client"))
	http.Handle("/", fs)

	// websocket endpoint
	http.HandleFunc("/ws", serveWs)

	log.Println("Listening...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func main() {
	messages := make(chan score)

	go startWebserver()
	go readFromSerial(messages)

	// for each new parsed message update the score
	for msg := range messages {
		scoreKeeper.UpdateScore(msg)
	}
}
