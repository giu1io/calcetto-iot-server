package main

import (
	"fmt"
	"sync"
	"time"
)

// ScoreUpdate response sent back
type ScoreUpdate struct {
	CurrentScore Score     `json:"currentScore"`
	LastScore    LastScore `json:"lastScore"`
}

// Score keep track of the scores
type Score struct {
	Blue int `json:"blue"`
	Red  int `json:"red"`
}

// LastScore keeps track of the last game score
type LastScore struct {
	Score       Score     `json:"score"`
	Timestamp   time.Time `json:"timestamp"`
	Displayable bool      `json:"displayable"`
}

// ScoreKeeper contains score keeping logic
type ScoreKeeper struct {
	Blue        int `json:"blue"`
	Red         int `json:"red"`
	subscribers []chan ScoreUpdate
	mutex       sync.Mutex
	lastScore   LastScore
}

// ResetScore resets the score
func (currentScore *ScoreKeeper) ResetScore() {
	currentScore.Red = 0
	currentScore.Blue = 0
	go currentScore.broadcast()
}

// UpdateScore increments the current score
func (currentScore *ScoreKeeper) UpdateScore(addScore Score) {
	currentScore.Red += addScore.Red
	currentScore.Blue += addScore.Blue

	if currentScore.Red >= 10 || currentScore.Blue >= 10 {
		currentScore.lastScore = LastScore{Score{currentScore.Blue, currentScore.Red}, time.Now(), true}
		currentScore.Red = 0
		currentScore.Blue = 0
	}

	go currentScore.broadcast()
}

/*
	Send message to each subscriber with current score
*/
func (currentScore *ScoreKeeper) broadcast() {
	for _, subscriber := range currentScore.subscribers {
		score := Score{Red: currentScore.Red, Blue: currentScore.Blue}
		subscriber <- ScoreUpdate{CurrentScore: score, LastScore: currentScore.lastScore}
	}
}

// Subscribe Create new channel, add to subscriber list synchronously and then send current score to the channel (async)
func (currentScore *ScoreKeeper) Subscribe() chan ScoreUpdate {
	c := make(chan ScoreUpdate)
	currentScore.mutex.Lock()
	currentScore.subscribers = append(currentScore.subscribers, c)
	currentScore.mutex.Unlock()

	go func() {
		score := Score{Red: currentScore.Red, Blue: currentScore.Blue}
		c <- ScoreUpdate{CurrentScore: score, LastScore: currentScore.lastScore}
	}()

	return c
}

// Unsubscribe Search the channel c between the subscribers synchronously, close it and remove it
func (currentScore *ScoreKeeper) Unsubscribe(c chan ScoreUpdate) {
	go func() {
		var foundIndex int
		var found = false

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

// ScoreKeeperBuilder Scorekeeper "class" constructor
func ScoreKeeperBuilder() ScoreKeeper {
	return ScoreKeeper{0, 0, []chan ScoreUpdate{}, sync.Mutex{}, LastScore{}}
}
