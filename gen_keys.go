package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

var aesKey = []byte("xianqu_ai_model_enc_key_2024!!!!")

func encrypt(plainText string) string {
	block, _ := aes.NewCipher(aesKey)
	aesGCM, _ := cipher.NewGCM(block)
	nonce := make([]byte, aesGCM.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	cipherText := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText)
}

func main() {
	fmt.Printf("DS: %s\n", encrypt("sk-deepseek-dummy-key-123456789"))
	fmt.Printf("GPT: %s\n", encrypt("sk-gpt3.5-dummy-key-987654321"))
	fmt.Printf("GLM: %s\n", encrypt("sk-glm4-dummy-key-555555555"))
}
