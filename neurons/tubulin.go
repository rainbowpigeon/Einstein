package neurons

import (
	"encoding/base32"
	"fmt"
	"strings"
)

const paddingChar = '1'
const zBase32Charset = "ybndrfg8ejkmcpqxot1uwisza345h769" // length 32, ref: https://philzimmermann.com/docs/human-oriented-base-32-encoding.txt

func Decode32(encoded string) ([]byte, error) {
	encoded = strings.ToUpper(encoded)
	dec, err := base32.StdEncoding.WithPadding(paddingChar).DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("Incorrect string encoding: %s", err)
	}
	return dec, nil
}

func Encode32(plain []byte) string {
	encoded := base32.StdEncoding.WithPadding(paddingChar).EncodeToString(plain)
	encoded = strings.ToLower(encoded)
	return encoded
}

func DecodeZ32(encoded string) ([]byte, error) {
	return base32.NewEncoding(zBase32Charset).DecodeString(encoded)
}

func EncodeZ32(plain []byte) string {
	return base32.NewEncoding(zBase32Charset).EncodeToString(plain)
}

func DecodeZ32Decrypt(encoded string) ([]byte, error) {
	decoded, err := DecodeZ32(encoded)
	if err != nil {
		return nil, err
	}
	decrypted, err := Decrypt(decoded)
	if err != nil {
		return nil, err
	}
	return decrypted, nil
}

func EncryptEncodeZ32(plain []byte) string {
	return EncodeZ32(Encrypt(plain))
}
