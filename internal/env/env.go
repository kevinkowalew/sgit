package env

import "os"

const (
	GithubToken    = "GITHUB_TOKEN"
	GithubUsername = "GITHUB_USERNAME"
	CodeHomeDir    = "CODE_HOME_DIR"
)

func AssertExists(envVars ...string) {
	for _, envVar := range envVars {
		if _, ok := os.LookupEnv(envVar); !ok {
			panic("Unset environment variable: " + envVar)
		}
	}
}
