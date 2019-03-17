package main

import (
	"fmt"
	"log"
	"strings"

	"go.bug.st/serial.v1"
)

type score struct {
	blue int
	red  int
}

func readFromSerial(messages chan score) {
	mode := &serial.Mode{
		BaudRate: 115200,
	}
	port, err := serial.Open("/dev/ttyp5", mode)
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

func parseMessage(message string) score {
	message = strings.TrimRight(message, "\r\n")
	s := score{red: 0, blue: 0}
	if message == "GOAL_RED_1" {
		s.red++
	}
	if message == "GOAL_BLUE_1" {
		s.blue++
	}
	return s
}

func main() {
	messages := make(chan score)
	go readFromSerial(messages)

	for msg := range messages {
		fmt.Println(msg)
	}

}
