package helpers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// Encrypt string to base64 crypto using AES
func Encrypt(authSecret string, text string) (data string, err error) {
	key := []byte(authSecret)
	encryptPass := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		return data, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(encryptPass))

	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return data, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], encryptPass)

	// convert to base64
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// Decrypt from base64 to decrypted string
func Decrypt(authSecret string, cryptoText string) (data string, err error) {
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)
	key := []byte(authSecret)
	block, err := aes.NewCipher(key)
	if err != nil {
		return data, err
	}

	if len(ciphertext) < aes.BlockSize {
		return data, errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	stream.XORKeyStream(ciphertext, ciphertext)

	return fmt.Sprintf("%s", ciphertext), nil
}
