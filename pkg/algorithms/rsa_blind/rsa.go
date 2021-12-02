package algorithms

import (
	"math/big"
	"crypto/rsa"
	"encoding/binary"
	"bytes"
)

func encryptRSA(pub *rsa.PublicKey, data *big.Int) *big.Int {
	e := big.NewInt(int64(pub.E))
	encrypted := big.NewInt(0).Exp(data, e, pub.N)
	return encrypted
}

func decryptRSA(priv *rsa.PrivateKey, data *big.Int) *big.Int {
	decrypted := big.NewInt(0).Exp(data, priv.D, priv.N)
	return decrypted
}

func crtCoefficient(p, q *big.Int) (*big.Int, *big.Int) {
	tq := big.NewInt(0).ModInverse(p, q)
	tp := big.NewInt(0).ModInverse(q, p)
	return big.NewInt(0).Mul(tp, q), big.NewInt(0).Mul(tq, p)
}

func powModCRT(x, d, n, p, q, cp, cq *big.Int) *big.Int {
	big_one := big.NewInt(1)
	rp := big.NewInt(0).Exp(x, getMod(d, getMinus(p, big_one)), p)
	rq := big.NewInt(0).Exp(x, getMod(d, getMinus(q, big_one)), q)
	return getMod(getAdd(getMul(rp, cp), getMul(rq, cq)), n)
}

func IntToBytes(n int) []byte {
    data := int64(n)
    bytebuf := bytes.NewBuffer([]byte{})
    binary.Write(bytebuf, binary.BigEndian, data)
    return bytebuf.Bytes()
}

func BytesToInt(bys []byte) int {
    bytebuff := bytes.NewBuffer(bys)
    var data int64
    binary.Read(bytebuff, binary.BigEndian, &data)
    return int(data)
}
