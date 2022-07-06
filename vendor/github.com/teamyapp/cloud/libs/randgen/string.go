package randgen

import (
	"math/rand"
)

var Base62 = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func String(alphabet []rune, length int) string {
	alphabetEndIndex := len(alphabet) - 1
	result := make([]rune, length)
	for i := 0; i < length; i++ {
		randomIndex := Int(0, alphabetEndIndex)
		result[i] = alphabet[randomIndex]
	}
	return string(result)
}

func Int(min int, max int) int {
	return min + rand.Intn(max-min+1)
}
