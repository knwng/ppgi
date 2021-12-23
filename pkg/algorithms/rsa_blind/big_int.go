package rsa_blind

import (
	"math/big"
)

func bytesToBigInt(x []byte) *big.Int {
	return big.NewInt(0).SetBytes(x)
}

func getMinus(x, y *big.Int) *big.Int {
	return big.NewInt(0).Sub(x, y)
}

func getAdd(x, y *big.Int) *big.Int {
	return big.NewInt(0).Add(x, y)
}

func getMul(x, y *big.Int) *big.Int {
	return big.NewInt(0).Mul(x, y)
}

func getMod(x, y *big.Int) *big.Int {
	// return big.NewInt(0).Div(x, y)
	return big.NewInt(0).Mod(x, y)
}

func getDivMod(x, y, m *big.Int) *big.Int {
	// return big.NewInt(0).Mod(big.NewInt(0).Div(x, y), m)
	return getMod(getMul(x, big.NewInt(0).ModInverse(y, m)), m)
}

func BigIntsToBytesSlice(dataList []*big.Int) [][]byte {
	ret := make([][]byte, len(dataList))
	for i, data := range dataList {
		ret[i] = data.Bytes()
	}

	return ret
}

func BytesSliceToBigInts(bytesSlice [][]byte) []*big.Int {
	ret := make([]*big.Int, len(bytesSlice))
	for i, bytes := range bytesSlice {
		ret[i] = big.NewInt(0).SetBytes(bytes)
	}

	return ret
}