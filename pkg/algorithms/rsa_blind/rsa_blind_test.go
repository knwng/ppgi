package rsa_blind

import (
	// "math/big"
	"testing"
)

func TestRSAIntersection(t *testing.T) {
	firstHash := "sha256"
	secondHash := "md5"
	RsaIntersect := NewRSABlindIntersect(firstHash, secondHash)

	bits := 4096
	privKey, pubKey, err := RsaIntersect.HostGenerateRSAKeyPair(bits)
	if err != nil {
		t.Fatal(err)
	}

	hostA := []string{"21022219911301911", "640111191119381029", "1732819483", "184", "97561890571"}
	hostB := []string{"640111191119381029", "1732819483", "3728172745", "97561890571"}

	// server offline compute
	ta := RsaIntersect.HostOfflineHash(hostA, privKey)

	// client offline compute
	yb, rands := RsaIntersect.ClientBlinding(hostB, pubKey)

	// server sign
	zb := RsaIntersect.HostBlindSigning(yb, privKey)

	// client unblinding
	tb := RsaIntersect.ClientUnblinding(zb, pubKey, rands)

	cmp_ret := RsaIntersect.CompareIds(ta, tb)
	for _, idx := range cmp_ret {
		t.Logf("equal: a: %s, b: %s", hostA[idx[0]], hostB[idx[1]])
	}
}
