package gotest

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

var seedGenerated bool

// GenerateRandomSeed - generates a random seed across the entire test
func GenerateRandomSeed() {

	if !seedGenerated {
		rand.Seed(time.Now().Unix())
		seedGenerated = true
	}
}

// GeneratePort - generates a port
func GeneratePort() int {

	GenerateRandomSeed()

	port, err := strconv.Atoi(fmt.Sprintf("1%d", rand.Intn(9999)))
	if err != nil {
		panic(err)
	}

	return port
}

// RandomInt - generates a random int
func RandomInt(min, max int) int {

	GenerateRandomSeed()

	return min + rand.Intn(max+1)
}
