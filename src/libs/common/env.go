package common

import (
	"log"
	"os"
)

type EnvVar struct {
	Key string
}

type PopulatedEnvVar struct {
	Key string
	Val string
}

func NewEnvVar(key string) *EnvVar {
	return &EnvVar{Key: key}
}

func (envvar *EnvVar) AssertExists() *PopulatedEnvVar {
	val, exists := os.LookupEnv(envvar.Key)
	if !exists {
		log.Fatalf("environment variable '%s' must be set", envvar.Key)
	}
	return &PopulatedEnvVar{
		Key: envvar.Key,
		Val: val,
	}
}

func (envvar *PopulatedEnvVar) AssertNonEmpty() string {
	if envvar.Val == "" {
		log.Fatalf("environment variable '%s' must not be empty", envvar.Key)
	}
	return envvar.Val
}
