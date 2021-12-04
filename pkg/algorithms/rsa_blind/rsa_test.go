package rsa_blind

import (
	"testing"
	"math/big"
	"crypto/rsa"
	"crypto/rand"
)

func TestRSAUnofficial(t *testing.T) {
	privkey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Fatal(err)
	}

	a := 111
	b := 231
	c := a * b

	a_enc := encryptRSA(&privkey.PublicKey, big.NewInt(int64(a)))
	b_enc := encryptRSA(&privkey.PublicKey, big.NewInt(int64(b)))

	c_enc := big.NewInt(0).Mul(a_enc, b_enc)

	c_dec := decryptRSA(privkey, c_enc)

	t.Logf("c: {%d}, c_dec: {%d}", c, c_dec.Int64())
}
