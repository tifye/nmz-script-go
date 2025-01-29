package main

import (
	"time"

	"golang.org/x/exp/rand"
)

type clock interface {
	Future(time.Duration) time.Time
	Sleep(time.Duration)
	Scale(time.Duration) time.Duration
}

type simulatedClock struct {
	timeScale float32
}

func (s simulatedClock) Future(d time.Duration) time.Time {
	return time.Now().Add(s.Scale(d))
}

func (s simulatedClock) Sleep(d time.Duration) {
	time.Sleep(s.Scale(d))
}

func (s simulatedClock) Scale(d time.Duration) time.Duration {
	scaledNano := float32(d.Nanoseconds()) * s.timeScale
	newDur := time.Duration(scaledNano)
	return newDur
}

func randomDuration(deviation time.Duration) time.Duration {
	return time.Duration(rand.Int63n(deviation.Nanoseconds()))
}

func randomSecondDuration(duration uint, deviation uint) time.Duration {
	return time.Duration(duration)*time.Second + randomDuration(time.Duration(deviation)*time.Second)
}

func randomMilisecondDuration(duration uint, deviation uint) time.Duration {
	return time.Duration(duration)*time.Millisecond + randomDuration(time.Duration(deviation)*time.Millisecond)
}
