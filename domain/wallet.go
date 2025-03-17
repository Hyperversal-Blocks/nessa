package domain

type Wallet struct {
	Address    string `json:"address"`
	PrivateKey []byte `json:"private_key"`
	Credential string `json:"credential"`
	MerkleRoot string `json:"merkle_root"`
}
