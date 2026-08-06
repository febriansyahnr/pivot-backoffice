package random

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
)

func GenerateOTP(maxDigits uint32) string {
	bi, _ := rand.Int(
		rand.Reader,
		big.NewInt(int64(math.Pow(10, float64(maxDigits)))),
	)
	return fmt.Sprintf("%0*d", maxDigits, bi)
}
