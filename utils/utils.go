package gotest

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"time"
)

// GeneratePort - generates a port
func GeneratePort() int {

	port, err := strconv.Atoi(fmt.Sprintf("1%d", RandomInt(1000, 8888)))
	if err != nil {
		panic(err)
	}

	return port
}

// RandomInt - generates a random int
func RandomInt(min, max int) int {

	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max+1)))
	if err != nil {
		panic(err)
	}

	return min + int(nBig.Int64())
}

// MustParseDuration - forces to parse the duration or it panics
func MustParseDuration(value string) time.Duration {

	d, err := time.ParseDuration(value)
	if err != nil {
		panic(err)
	}

	return d
}
