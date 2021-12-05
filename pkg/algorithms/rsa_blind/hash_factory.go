package rsa_blind

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
)

type Hasher interface {
	Sum([]byte) []byte
}

type MD5Hash struct {}

func (s *MD5Hash) Sum(msg []byte) []byte {
	hash := md5.Sum(msg)
	return hash[:]
}

type SHA256Hash struct {}

func (s *SHA256Hash) Sum(msg []byte) []byte {
	hash := sha256.Sum256(msg)
	return hash[:]
}

type SHA224Hash struct {}

func (s *SHA224Hash) Sum(msg []byte) []byte {
	hash := sha256.Sum224(msg)
	return hash[:]
}

type SHA512Hash struct {}

func (s *SHA512Hash) Sum(msg []byte) []byte {
	hash := sha512.Sum512(msg)
	return hash[:]
}

func getHasher(name string) Hasher {
	switch name {
	case "md5":
		return &MD5Hash{}
	case "sha256":
		return &SHA256Hash{}
	case "sha224":
		return &SHA224Hash{}
	case "sha512":
		return &SHA512Hash{}
	default:
		return nil
	}
}
