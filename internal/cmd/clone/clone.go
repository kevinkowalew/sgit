package clone

import (
	"errors"
	"fmt"
	"os"
	"sgit/internal/interactor"
	"strings"

	"github.com/spf13/cobra"
)

var (
	langs, states *string
	forks         *bool
)

var Cmd = &cobra.Command{
	Use:   "clone",
	Short: "clone repo(s) based on provided filters",
	Long:  "clone repo(s) based on provided filters",
	RunE:  run,
}

func init() {
	langs = Cmd.PersistentFlags().StringP("langs", "l", "", "comma-separated list of languages to target")
	states = Cmd.PersistentFlags().StringP("states", "s", "", "comma-separated list of states to target")
	forks = Cmd.PersistentFlags().BoolP("forks", "f", false, "target forked or non-forked repos")
}

func run(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("fatal: You must specify a repository to clone.")
	}

	url := args[0]
	p := strings.Split(url, "/")

	name, username := "", ""
	if len(p) > 1 {
		name = strings.Split(p[len(p)-1], ".")[0]

		p := strings.Split(p[len(p)-2], ":")
		username = p[len(p)-1]
	} else {
		name = strings.Split(p[len(p)-1], ".")[0]
		username = os.Getenv("GITHUB_USERNAME")
	}

	i := interactor.New()
	lang, err := i.GetPrimaryLanguageForRepo(cmd.Context(), username, name)
	if err != nil {
		return fmt.Errorf("interactor.GetPrimaryLanguageForRepo failed: %w", err)
	}

	if err := i.CreateDir(lang); err != nil {
		return fmt.Errorf("interactor.CreateDir failed: %w", err)
	}

	url = fmt.Sprintf("https://github.com/%s/%s", username, name)
	return i.Clone(lang, url)
}
