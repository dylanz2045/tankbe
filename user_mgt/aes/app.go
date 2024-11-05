package aes

const (
	BlockSize = 16
)

type Aes interface {
	Encrypt(string) (string, error)
	Decrypt(string) (string, error)
}

type UserInfo struct {
}

func NewAes() Aes {
	return &UserInfo{}
}
