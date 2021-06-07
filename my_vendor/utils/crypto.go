package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func PKCS7Padding(plainText []byte, blockSize int) []byte {
	padding := blockSize - len(plainText)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plainText, padtext...)
}

func PKCS7UnPadding(plainText []byte) []byte {
	length := len(plainText)
	unpadding := int(plainText[length-1])
	return plainText[:(length - unpadding)]
}

func AesEncrypt(plainText, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	plainText = PKCS7Padding(plainText, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(plainText))
	blockMode.CryptBlocks(crypted, plainText)
	return string(base64.StdEncoding.EncodeToString(crypted)), nil
}

func AesDecrypt(cryptedText, key []byte) (string, error) {
	cryptedByts, err := base64.StdEncoding.DecodeString(string(cryptedText))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	plainByts := make([]byte, len(cryptedByts))
	blockMode.CryptBlocks(plainByts, cryptedByts)
	plainByts = PKCS7UnPadding(plainByts)
	return string(plainByts), nil
}

func Md5(plainText string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(plainText)))
}

func Sha256(plainText string) string {
	sum := sha256.Sum256([]byte(plainText))
	return fmt.Sprintf("%x", sum)
}

func HmacSha256(str string, secret string, raw bool) string {
	var sha []byte
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(str))
	if raw {
		sha = h.Sum(nil)
	} else {
		sha = []byte(hex.EncodeToString(h.Sum(nil)))
	}
	return base64.StdEncoding.EncodeToString(sha)
}
