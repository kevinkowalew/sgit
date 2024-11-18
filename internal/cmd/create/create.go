package create

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	langs, states *string
	forks         *bool

	Cmd = &cobra.Command{
		Use:   "create",
		Short: "create an a remote repo for the current directory",
		Long:  "create an a remote repo for the current directory",
		RunE:  run,
	}
)

func run(cmd *cobra.Command, args []string) error {
	fmt.Println("TODO: implement me")
	return nil
}
