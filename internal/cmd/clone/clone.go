package clone

import (
	"fmt"

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
	Run:   run,
}

func init() {
	langs = Cmd.PersistentFlags().StringP("langs", "l", "", "comma-separated list of languages to target")
	states = Cmd.PersistentFlags().StringP("states", "s", "", "comma-separated list of states to target")
	forks = Cmd.PersistentFlags().BoolP("forks", "f", false, "target forked or non-forked repos")
}

func run(cmd *cobra.Command, args []string) {
	// TODO: implement me
	fmt.Println("TODO: implement me")
}
