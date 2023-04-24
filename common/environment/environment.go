package environment

import "os"

const (
	ENV   = "ENV"
	LOCAL = "local"
)

func IsLocalEnvironment() bool {
	environment := os.Getenv(ENV)
	if len(environment) == 0 || environment == LOCAL {
		return true
	}
	return false
}
