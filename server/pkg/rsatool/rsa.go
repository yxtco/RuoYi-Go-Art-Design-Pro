package rsatool

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"
)

// Decrypt 使用私钥字符串解密Base64编码的密文
func Decrypt(ciphertextBase64, privateKeyString string) (string, error) {
	// 解码Base64编码的密文
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", err
	}

	// 解析私钥字符串（去除所有空白字符，兼容YAML折叠格式）
	cleanKey := strings.Join(strings.Fields(privateKeyString), "")
	privateKeyBlock, _ := pem.Decode([]byte("-----BEGIN PRIVATE KEY-----\n" + cleanKey + "\n-----END PRIVATE KEY-----"))
	if privateKeyBlock == nil {
		return "", fmt.Errorf("无法解析私钥字符串")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return "", err
	}

	// 使用私钥解密
	plainText, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey.(*rsa.PrivateKey), ciphertext)
	if err != nil {
		return "", err
	}

	// 返回明文
	return string(plainText), nil
}

// Encrypt 使用公钥字符串加密明文并返回Base64编码的密文
func Encrypt(plainText, publicKeyString string) (string, error) {
	// 解析公钥字符串（去除所有空白字符，兼容YAML折叠格式）
	cleanKey := strings.Join(strings.Fields(publicKeyString), "")
	publicKeyBlock, _ := pem.Decode([]byte("-----BEGIN PUBLIC KEY-----\n" + cleanKey + "\n-----END PUBLIC KEY-----"))
	if publicKeyBlock == nil {
		return "", fmt.Errorf("无法解析公钥字符串")
	}

	publicKeyInterface, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return "", err
	}

	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("无法转换为RSA公钥")
	}

	// 使用公钥加密
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(plainText))
	if err != nil {
		return "", err
	}

	// 返回Base64编码的密文
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
