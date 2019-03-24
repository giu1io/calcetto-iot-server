package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/viper"
	"go.bug.st/serial.v1"
)

var (
	scoreKeeper = ScoreKeeperBuilder()
	port        serial.Port
	portOpen    bool
)

/*
	This function reads from a serial device and relays parsed messages to the channel given in input
	Serial configurations will need to be tweaked accordingly and maybe read from a configuration file
*/
func readFromSerial(messages chan score) {
	mode := &serial.Mode{
		BaudRate: 115200,
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
	if err := http.ListenAndServe(":"+viper.GetString("httpServerPort"), nil); err != nil {
		log.Fatal(err)
	}
}

func initizeConfigurations() {
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

func updateScore(messages chan score) {
	for msg := range messages {
		scoreKeeper.UpdateScore(msg)
	}
}

func main() {
	initizeConfigurations()

	messages := make(chan score)

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
