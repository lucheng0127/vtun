package cipher

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"github.com/lucheng0127/vtun/pkg/utils"
)

type AESCipher struct {
	Key string
}

func NewAESCipher(key string) (Cipher, error) {
	if !utils.ValidateKey(key) {
		return nil, errors.New("invalid key size, 16, 24 or 32")
	}
	return &AESCipher{Key: key}, nil
}

// AES CBC can only encrypt data with blocksize,
// for data large than blocksize, it split it into
// several blocks, use last result as input to encrypt
// the next block, IV as the input for encrypt the
// first block. So length of date wait for encrypt
// should multiple blocksize, this why need padding.
func (ac *AESCipher) PKCS7Padding(data []byte) []byte {
	paddingLen := aes.BlockSize - len(data)%aes.BlockSize
	padData := bytes.Repeat([]byte{byte(paddingLen)}, paddingLen)
	return append(data, padData...)
}

func (ac *AESCipher) PKCS7Unpadding(data []byte) []byte {
	dataLen := len(data)
	unpaddingLen := int(data[dataLen-1])

	if dataLen < unpaddingLen {
		// For wrong AES key, it can't parse correct, so if dataLen < unpaddingLen, return nothing
		return make([]byte, 0)
	}

	return data[:(dataLen - unpaddingLen)]
}

func (ac *AESCipher) Encrypt(data []byte) ([]byte, error) {
	// Create cipher block
	block, err := aes.NewCipher([]byte(ac.Key))
	if err != nil {
		return make([]byte, 0), err
	}

	// Padding data first
	data = ac.PKCS7Padding(data)

	// Add IV
	cipherData := make([]byte, aes.BlockSize+len(data))
	iv := cipherData[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return make([]byte, 0), err
	}

	// Encrypt
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherData[aes.BlockSize:], data)
	return cipherData, nil
}

func (ac *AESCipher) Decrypt(data []byte) ([]byte, error) {
	if len(data) < aes.BlockSize || len(data)%aes.BlockSize != 0 {
		return make([]byte, 0), fmt.Errorf("invalid cipher data size")
	}

	// Create cipher Block
	block, err := aes.NewCipher([]byte(ac.Key))
	if err != nil {
		return make([]byte, 0), err
	}

	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]

	// Decrypt
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(data, data)

	// Unpadding
	data = ac.PKCS7Unpadding(data)
	return data, nil
}
