package rand

import (
	"crypto/rand"
	"encoding/base64"
)

const RememberTokenBytes = 32

// Bytes will help generate n random bytes
// or will return an error
// This uses crypt/rand package
func Bytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Creates a byte slice of length n,
// then returns a base64 encoded string
// of that byte slice
func String(nBytes int) (string, error){
	b, err := Bytes(nBytes)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

// RemeberToken generates remember tokens of 
// a predetermined length
func RememberToken() (string, error) {
	return String(RememberTokenBytes)
}