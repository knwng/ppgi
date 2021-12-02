package algorithms

import (
	"math/big"
	"testing"
)

func TestRSAIntersection(t *testing.T) {
	bits := 4096
	privKey, pubKey, err := HostGenerateRSAKeyPair(bits)
	if err != nil {
		t.Fatal(err)
	}

	hostA := []string{"21022219911301911", "640111191119381029", "1732819483", "184"}
	hostB := []string{"640111191119381029", "1732819483", "3728172745"}

	// server offline compute
	ta := make([][]byte, len(hostA))
	HostOfflineHash(hostA, privKey, ta)

	// client offline compute
	yb := make([]*big.Int, len(hostB))
	rands := make([]*big.Int, len(hostB))
	ClientBlinding(hostB, pubKey, yb, rands)

	// server sign
	zb := make([]*big.Int, len(yb))
	HostBlindSigning(yb, privKey, zb)

	// client unblinding
	tb := make([][]byte, len(hostB))
	ClientUnblinding(zb, pubKey, rands, tb)

	cmp_ret := CompareIds(ta, tb)
	for _, idx := range cmp_ret {
		t.Logf("equal: a: %s, b: %s", hostA[idx[0]], hostB[idx[1]])
	}
}
