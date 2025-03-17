package utils

import (
	"LinkHUB/config"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// padKey 使用PKCS7填充确保密钥长度为16字节
func padKey(key []byte) []byte {
	if len(key) >= 16 {
		return key[:16]
	}
	padding := make([]byte, 16-len(key))
	for i := range padding {
		padding[i] = byte(16 - len(key))
	}
	return append(key, padding...)
}

// EncryptUserID 加密用户ID
func EncryptUserID(userID string) (string, error) {
	key := padKey([]byte(config.GetConfig().JWT.Secret))
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(userID))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(userID))

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// DecryptUserID 解密用户ID
func DecryptUserID(encrypted string) (string, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	key := padKey([]byte(config.GetConfig().JWT.Secret))
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}
