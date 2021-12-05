package rsa_blind

import (
	"crypto/rand"
	"crypto/rsa"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRSAUnofficial(t *testing.T) {
	privkey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Fatal(err)
	}

	a := 111
	b := 231
	c := a * b

	encA := encryptRSA(&privkey.PublicKey, big.NewInt(int64(a)))
	encB := encryptRSA(&privkey.PublicKey, big.NewInt(int64(b)))

	encC := big.NewInt(0).Mul(encA, encB)

	decC := decryptRSA(privkey, encC)

	t.Logf("c: {%d}, decC: {%d}", c, decC.Int64())
	assert.Equal(t, int64(c), decC.Int64())
}
