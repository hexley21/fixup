package encryption

type Encryptor interface {
	Encrypt(value []byte) ([]byte, error)
}

type Decryptor interface {
	Decrypt(value []byte) ([]byte, error)
}
