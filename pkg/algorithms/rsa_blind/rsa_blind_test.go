package rsa_blind

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestRSAIntersection(t *testing.T) {
	bits := 4096
	firstHash := "sha256"
	secondHash := "md5"
	client, err := NewRSABlindIntersect(bits, firstHash, secondHash, "client")
	assert.NoError(t, err)
	server, err := NewRSABlindIntersect(bits, firstHash, secondHash, "server")
	assert.NoError(t, err)

	n, e := server.GetPubKey()
	client.SetPubKey(n, e)

	hostA := []string{"21022219911301911", "640111191119381029", "1732819483", "184", "97561890571"}
	hostB := []string{"640111191119381029", "1732819483", "3728172745", "97561890571"}

	target := [][2]int{{1, 0}, {2, 1}, {4, 3}}

	// server offline compute
	ta := server.HostOfflineHash(hostA)

	// client offline compute
	yb, rands, err := client.ClientBlinding(hostB)
	assert.NoError(t, err)

	// server sign
	zb := server.HostBlindSigning(yb)

	// client unblinding
	tb := client.ClientUnblinding(zb, rands)

	cmp_ret := server.CompareIds(ta, tb)
	for _, idx := range cmp_ret {
		t.Logf("equal: a: %s, b: %s", hostA[idx[0]], hostB[idx[1]])
	}
	assert.Equal(t, target, cmp_ret)
}
