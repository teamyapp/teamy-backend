package resolver

import "math/rand"

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	sequence := make([]rune, n)
	for i := range sequence {
		randomIndex := rand.Intn(len(letters))
		sequence[i] = letters[randomIndex]
	}

	return string(sequence)
}
