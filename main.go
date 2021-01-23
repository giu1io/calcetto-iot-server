package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"

	"github.com/spf13/viper"
	"go.bug.st/serial.v1"
)

type parsedMessage struct {
	Blue  int
	Red   int
	Reset bool
}

var (
	scoreKeeper  = ScoreKeeperBuilder()
	port         serial.Port
	portOpen     bool
	matchRed, _  = regexp.Compile("GOAL_RED_[0-9]{1,}")
	matchBlue, _ = regexp.Compile("GOAL_BLUE_[0-9]{1,}")
)

/*
	This function reads from a serial device and relays parsed messages to the channel given in input
	Serial configurations will need to be tweaked accordingly and maybe read from a configuration file
*/
func readFromSerial(messages chan parsedMessage) {
	mode := &serial.Mode{
		BaudRate: 9600,
	}
	var err error

	port, err = serial.Open(viper.GetString("serialDevice"), mode)
	if err != nil {
		log.Println("Serial Error")
		log.Fatal(err)
	}

	portOpen = true

	buff := make([]byte, 100)

	//defer port.Close()

	for {
		n, err := port.Read(buff)
		if err != nil && portOpen {
			log.Println("Serial Error")
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
func parseMessage(message string) parsedMessage {
	message = strings.TrimRight(message, "\r\n")
	s := parsedMessage{Red: 0, Blue: 0, Reset: false}
	if matchRed.MatchString(message) {
		s.Red++
	}
	if matchBlue.MatchString(message) {
		s.Blue++
	}
	if message == "MATCH_START\r\nStart Game" {
		s.Reset = true
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
	if err := http.ListenAndServe(":"+viper.GetString("httpServerPort"), nil); err != nil {
		log.Fatal(err)
	}
}

func initializeConfigurations() {
	viper.SetDefault("serialDevice", "/dev/ttyp7")
	viper.SetDefault("httpServerPort", "8080")
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/hookbot/")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

func updateScore(messages chan parsedMessage) {
	for msg := range messages {
		if msg.Reset {
			scoreKeeper.ResetScore()
		} else {
			scoreKeeper.UpdateScore(Score{Red: msg.Red, Blue: msg.Blue})
		}
	}
}

func main() {
	initializeConfigurations()

	messages := make(chan parsedMessage)

	go startWebserver()
	go readFromSerial(messages)

	// for each new parsed message update the score
	go updateScore(messages)

	// handle interrupts
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan struct{})
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		log.Println("Received an interrupt, stopping services...")
		if port != nil && portOpen {
			portOpen = false
			port.Close()
		}
		close(cleanupDone)
	}()
	<-cleanupDone
}
