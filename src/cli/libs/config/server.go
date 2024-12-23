package config

import (
	"strconv"
	"strings"
)

type (
	ServerConfig struct {
		Host string `json:"host"`
		Port int64  `json:"port"`
	}
)

func (c *ServerConfig) Url() string {
	return strings.Join([]string{c.Host, strconv.FormatInt(c.Port, 10)}, ":")
}
