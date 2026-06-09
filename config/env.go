package config

import (
	"os"
	"strings"
)

var Env string

func init() {
	Env = os.Getenv("ENV")
}

const (
	Local = "LOCAL"
	Dev   = "DEV"
	UAT   = "UAT"
	Prod  = "PROD"
)

func IsLocalEnv() bool {
	return strings.ToUpper(Env) == Local
}
