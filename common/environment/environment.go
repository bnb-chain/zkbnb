package environment

import "os"

const (
	ENV   = "ENV"
	LOCAL = "local"
)

func IsLocalEnvironment() bool {
	environment := os.Getenv(ENV)
	if environment == LOCAL {
		return true
	}
	return false
}
