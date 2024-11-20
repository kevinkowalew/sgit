package cmd

import (
	"context"
	"fmt"
	"os"
	"sgit/internal/cmd/clone"
	del "sgit/internal/cmd/delete"
	"sgit/internal/cmd/ls"

	"github.com/spf13/cobra"
)

var (
	langs, states *string
	forks         *bool
)

var cmd = &cobra.Command{
	Use:   "sgit",
	Short: "git synchronization made simple",
	Long:  "git synchronization made simple",
}

func Execute() {
	assert("GITHUB_TOKEN")
	assert("GITHUB_USERNAME")
	assert("CODE_HOME_DIR")

	ctx := context.Background()
	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func assert(envVar string) {
	if _, ok := os.LookupEnv(envVar); !ok {
		panic("Unset environment variable: " + envVar)
	}
}

func init() {
	cmd.AddCommand(ls.Cmd)
	cmd.AddCommand(clone.Cmd)
	cmd.AddCommand(del.Cmd)
}
