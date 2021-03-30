package app


import (
	"math/rand"
	"time"
)

func RandInt() int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(100)
}
