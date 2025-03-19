package signer

type signer struct {
}

type Signer interface{}

func NewSigner() Signer {
	return &signer{}
}
