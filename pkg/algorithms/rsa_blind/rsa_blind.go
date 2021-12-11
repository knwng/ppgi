package rsa_blind

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"log"
	"math/big"
)

type RSABlindIntersect struct {
	firstHash Hasher
	secondHash Hasher
}

func NewRSABlindIntersect(firstHash, secondHash string) *RSABlindIntersect {
	return &RSABlindIntersect{
		firstHash: getHasher(firstHash),
		secondHash: getHasher(secondHash),
	}
}

func (s *RSABlindIntersect) HostGenerateRSAKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	return privKey, &privKey.PublicKey, nil
}

func (s *RSABlindIntersect) HostOfflineHash(msgs []string, privKey *rsa.PrivateKey) [][]byte {
	ta := make([][]byte, len(msgs))
	for i, msg := range msgs {
		hi := s.firstHash.Sum([]byte(msg))
		hiEnc := decryptRSA(privKey, bytesToBigInt(hi[:]))
		hiHashed := s.secondHash.Sum(hiEnc.Bytes())
		ta[i] = hiHashed[:]
	}
	return ta
}

func (s *RSABlindIntersect) ClientBlinding(msgs []string, pubKey *rsa.PublicKey) ([]*big.Int, []*big.Int) {
	yb := make([]*big.Int, len(msgs))
	rands := make([]*big.Int, len(msgs))

	for i, msg := range msgs {
		hi := s.firstHash.Sum([]byte(msg))
		r, err := rand.Int(rand.Reader, pubKey.N)
		if err != nil {
			log.Fatal(err)
		}
		rands[i] = r

		rEnc := encryptRSA(pubKey, r)
		yb[i] = getMod(big.NewInt(0).Mul(bytesToBigInt(hi[:]), rEnc), pubKey.N)
	}
	return yb, rands
}

func (s *RSABlindIntersect) HostBlindSigning(yb []*big.Int, privKey *rsa.PrivateKey) []*big.Int {
	zb := make([]*big.Int, len(yb))
	for i, e := range yb {
		zb[i] = decryptRSA(privKey, e)
	}
	return zb
}

func (s *RSABlindIntersect) ClientUnblinding(zb []*big.Int, pubKey *rsa.PublicKey, rands []*big.Int) [][]byte {
	tb := make([][]byte, len(zb))
	for i, e := range zb {
		tb_array := s.secondHash.Sum(getDivMod(e, rands[i], pubKey.N).Bytes())
		tb[i] = tb_array[:]
	}
	return tb
}

func (s *RSABlindIntersect) CompareIds(ta, tb [][]byte) [][2]int {
	cmp_ret := make([][2]int, 0)
	for i, m := range ta {
		for j, n := range tb {
			if bytes.Compare(m[:], n[:]) == 0 {
				cmp_ret = append(cmp_ret, [2]int{i, j})
			}
		}
	}
	return cmp_ret
}
