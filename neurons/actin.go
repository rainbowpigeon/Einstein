package neurons

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"log"
)

var (
	// change key to whatever you like
	key = []byte{
		0xCE, 0x90, 0xA6, 0x9C, 0xFA, 0xBA, 0x12, 0xB8,
		0x42, 0x01, 0xB7, 0x66, 0x35, 0xE3, 0x74, 0x07}
	gcm       cipher.AEAD
	nonceSize int
)

func init() {
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatalf("Error reading key: %s\n", err.Error())
	}

	gcm, err = cipher.NewGCM(block)
	if err != nil {
		log.Fatalf("Error initializing AEAD: %s\n", err.Error())
	}
	nonceSize = gcm.NonceSize() // this should return 12
}

func getRandBytes(length int) []byte {
	b := make([]byte, length)
	rand.Read(b)
	return b
}

func Decrypt(ciphertext []byte) (plaintext []byte, err error) {
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("Ciphertext too short.")
	}
	nonce := ciphertext[:nonceSize]
	msg := ciphertext[nonceSize:]
	// gcm.Open will take in the ciphertext + authtag
	return gcm.Open(nil, nonce, msg, nil)
}

// 12 byte nonce + ciphertext + 16 byte authtag
// gcm.Overhead() should return 16 bytes for the authtag length
func Encrypt(plaintext []byte) (ciphertext []byte) {
	nonce := getRandBytes(nonceSize)
	c := gcm.Seal(nil, nonce, plaintext, nil) // last nil parameter is for optional AAD
	return append(nonce, c...)
}
