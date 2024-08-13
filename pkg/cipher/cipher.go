package cipher

type Cipher interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

type NonCipher struct{}

func (nc *NonCipher) Encrypt(data []byte) ([]byte, error) {
	return data, nil
}

func (nc *NonCipher) Decrypt(data []byte) ([]byte, error) {
	return data, nil
}
