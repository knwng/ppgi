package algorithms

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"

	"log"
	"math/big"
)

func HostGenerateRSAKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	return privKey, &privKey.PublicKey, nil
}

func HostOfflineHash(msgs []string, privKey *rsa.PrivateKey, ta [][]byte) {
	for i, msg := range msgs {
		hi := sha256.Sum256([]byte(msg))
		hiEnc := decryptRSA(privKey, bytesToBigInt(hi[:]))
		hiHashed := md5.Sum(hiEnc.Bytes())
		ta[i] = hiHashed[:]
	}
}

func ClientBlinding(msgs []string, pubKey *rsa.PublicKey, yb []*big.Int, rands []*big.Int) {
	for i, msg := range msgs {
		hi := sha256.Sum256([]byte(msg))
		r, err := rand.Int(rand.Reader, pubKey.N)
		if err != nil {
			log.Fatal(err)
		}
		rands[i] = r

		rEnc := encryptRSA(pubKey, r)
		yb[i] = getMod(big.NewInt(0).Mul(bytesToBigInt(hi[:]), rEnc), pubKey.N)
	}
}

func HostBlindSigning(yb []*big.Int, privKey *rsa.PrivateKey, zb []*big.Int) {
	for i, e := range yb {
		zb[i] = decryptRSA(privKey, e)
	}
}

func ClientUnblinding(zb []*big.Int, pubKey *rsa.PublicKey, rands []*big.Int, tb [][]byte) {
	for i, e := range zb {
		tb_array := md5.Sum(getDivMod(e, rands[i], pubKey.N).Bytes())
		tb[i] = tb_array[:]
	}
}

func CompareIds(ta [][]byte, tb [][]byte) [][2]int {
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
