package gotest

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

var randInstance *rand.Rand

// GenerateRandomSeed - generates a random seed across the entire test
func GenerateRandomSeed() {

	if randInstance == nil {
		randInstance = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
}

// GeneratePort - generates a port
func GeneratePort() int {

	GenerateRandomSeed()

	port, err := strconv.Atoi(fmt.Sprintf("1%d", RandomInt(1000, 8888)))
	if err != nil {
		panic(err)
	}

	return port
}

// RandomInt - generates a random int
func RandomInt(min, max int) int {

	GenerateRandomSeed()

	return min + randInstance.Intn(max+1)
}

// MustParseDuration - forces to parse the duration or it panics
func MustParseDuration(value string) time.Duration {

	d, err := time.ParseDuration(value)
	if err != nil {
		panic(err)
	}

	return d
}
