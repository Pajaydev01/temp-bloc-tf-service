package misc

import (
	"fmt"
	math "math/rand"
	"time"
)

func GenerateRandomDigits(length int) string {
	SeededRand := math.New(math.NewSource(time.Now().UnixNano()))
	numberCharSet := "0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = numberCharSet[SeededRand.Intn(len(numberCharSet))]
	}

	randomNumbers := fmt.Sprintf("%s", string(b))

	return randomNumbers
}
