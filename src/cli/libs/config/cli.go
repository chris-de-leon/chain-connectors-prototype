package config

type (
	CliConfig struct {
		Chains map[string]ChainConfig `json:"chains"`
	}
)

func (c *CliConfig) ChainNames() []string {
	chains := []string{}
	for k := range c.Chains {
		chains = append(chains, k)
	}
	return chains
}
