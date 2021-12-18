package rsa_blind

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"math/big"
	"errors"
	"fmt"
)

type RSAStep string

const (
	StepHostSendPubKey = "HostSendPubkey"
	StepHostHash = "HostHash"
	StepHostBlindSign = "HostBlindSign"
	StepClientReceivedPubKey = "ClientReceivedPubkey"
	StepClientBlind = "ClientBlind"
	StepClientUnblind = "ClientUnblind"
)

type RSABlindIntersect struct {
	firstHash 	Hasher
	secondHash 	Hasher
	privKey		*rsa.PrivateKey
	pubKey		*rsa.PublicKey
}

func NewRSABlindIntersect(bits int, firstHash, secondHash, role string) (*RSABlindIntersect, error) {
	if role == "server" {
		privKey, pubKey, err := generateRSAKeyPair(bits)
		if err != nil {
			return nil, err
		}

		return &RSABlindIntersect{
			firstHash: getHasher(firstHash),
			secondHash: getHasher(secondHash),
			privKey: privKey,
			pubKey: pubKey,
		}, nil
	} else if role == "client" {
		return &RSABlindIntersect{
			firstHash: getHasher(firstHash),
			secondHash: getHasher(secondHash),
		}, nil
	} else {
		return nil, errors.New(fmt.Sprintf("Unsupported role: %s", role))
	}
}

func generateRSAKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	return privKey, &privKey.PublicKey, nil
}

func (s RSABlindIntersect) HasPubKey() bool {
	return s.pubKey != nil
}

func (s RSABlindIntersect) GetPubKey() ([]byte, int) {
	return s.pubKey.N.Bytes(), s.pubKey.E
}

func (s *RSABlindIntersect) SetPubKey(n []byte, e int) {
	s.pubKey = &rsa.PublicKey{
		N: big.NewInt(0).SetBytes(n),
		E: e,
	}
}

func (s *RSABlindIntersect) HostOfflineHash(msgs []string) [][]byte {
	ta := make([][]byte, len(msgs))
	for i, msg := range msgs {
		hi := s.firstHash.Sum([]byte(msg))
		hiEnc := decryptRSA(s.privKey, bytesToBigInt(hi[:]))
		hiHashed := s.secondHash.Sum(hiEnc.Bytes())
		ta[i] = hiHashed[:]
	}
	return ta
}

func (s *RSABlindIntersect) ClientBlinding(msgs []string) ([]*big.Int, []*big.Int, error) {
	yb := make([]*big.Int, len(msgs))
	rands := make([]*big.Int, len(msgs))

	for i, msg := range msgs {
		hi := s.firstHash.Sum([]byte(msg))
		r, err := rand.Int(rand.Reader, s.pubKey.N)
		if err != nil {
			return nil, nil, errors.New(fmt.Sprintf("Getting Random num failed, err: %s", err))
		}
		rands[i] = r

		rEnc := encryptRSA(s.pubKey, r)
		yb[i] = getMod(big.NewInt(0).Mul(bytesToBigInt(hi[:]), rEnc), s.pubKey.N)
	}
	return yb, rands, nil
}

func (s *RSABlindIntersect) HostBlindSigning(yb []*big.Int) []*big.Int {
	zb := make([]*big.Int, len(yb))
	for i, e := range yb {
		zb[i] = decryptRSA(s.privKey, e)
	}
	return zb
}

func (s *RSABlindIntersect) ClientUnblinding(zb []*big.Int, rands []*big.Int) [][]byte {
	tb := make([][]byte, len(zb))
	for i, e := range zb {
		tb_array := s.secondHash.Sum(getDivMod(e, rands[i], s.pubKey.N).Bytes())
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
