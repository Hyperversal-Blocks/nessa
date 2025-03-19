package config

type NodeConfig struct{}

type ServerConfig struct {
	HOST    string `json:"HOST"`
	PORT    string `json:"PORT"`
	CORSAGE int    `json:"CORS_AGE"`
}
