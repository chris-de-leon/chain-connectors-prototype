package config

type (
	ChainConfig struct {
		Server *ServerConfig    `json:"server"`
		Conn   *ConnectionConfg `json:"conn"`
		Plugin *PluginConfig    `json:"plugin"`
	}
)
