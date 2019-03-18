package main

import (
	"fmt"
	"sync"
)

type score struct {
	Blue int `json:"blue"`
	Red  int `json:"red"`
}

type ScoreKeeper struct {
	Blue        int `json:"blue"`
	Red         int `json:"red"`
	subscribers []chan score
	mutex       sync.Mutex
}

func (currentScore *ScoreKeeper) ResetScore() {
	currentScore.Red = 0
	currentScore.Blue = 0
	go currentScore.broadcast()
}

func (currentScore *ScoreKeeper) UpdateScore(addScore score) {
	currentScore.Red += addScore.Red
	currentScore.Blue += addScore.Blue
	go currentScore.broadcast()
}

/*
	Send message to each subscriber with current score
*/
func (currentScore *ScoreKeeper) broadcast() {
	for _, subscriber := range currentScore.subscribers {
		subscriber <- score{Red: currentScore.Red, Blue: currentScore.Blue}
	}
}

/*
	Create new channel, add to subscriber list synchronously and then send current score to the channel (async)
*/
func (currentScore *ScoreKeeper) Subscribe() chan score {
	c := make(chan score)
	currentScore.mutex.Lock()
	currentScore.subscribers = append(currentScore.subscribers, c)
	currentScore.mutex.Unlock()

	go func() {
		c <- score{Red: currentScore.Red, Blue: currentScore.Blue}
	}()

	return c
}

/*
	Search the channel c between the subscribers synchronously, close it and remove it
*/
func (currentScore *ScoreKeeper) Unsubscribe(c chan score) {
	go func() {
		var foundIndex int
		var found bool = false

		currentScore.mutex.Lock()
		for index, subscriber := range currentScore.subscribers {
			if subscriber == c {
				found = true
				foundIndex = index
				close(c)
			}
		}

		if found {
			currentScore.subscribers = append(currentScore.subscribers[:foundIndex], currentScore.subscribers[foundIndex+1:]...)
		}
		currentScore.mutex.Unlock()

		fmt.Printf("Removed index: %d, %d remaining subscribers.\n", foundIndex, len(currentScore.subscribers))
	}()
}

/*
	Scorekeeper "class" constructor
*/
func ScoreKeeperBuilder() ScoreKeeper {
	return ScoreKeeper{0, 0, []chan score{}, sync.Mutex{}}
}
