package ls

import (
	"fmt"
	"sgit/internal/interactor"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	langs, states *string
	forks         *bool

	Cmd = &cobra.Command{
		Use:   "ls",
		Short: "list repos and their corresponding states",
		Long:  "list repos and their corresponding states",
		RunE:  run,
	}
)

func init() {
	langs = Cmd.PersistentFlags().StringP("langs", "l", "", "comma-separated list of languages to target")
	states = Cmd.PersistentFlags().StringP("states", "s", "", "comma-separated list of states to target")
	forks = Cmd.PersistentFlags().BoolP("forks", "f", false, "target forked or non-forked repos")
}

func run(cmd *cobra.Command, args []string) error {
	var forksFlag *bool
	if cmd.Flags().Changed("forks") {
		forksFlag = forks
	}

	filter, err := interactor.NewFilter(*langs, *states, forksFlag)
	if err != nil {
		return fmt.Errorf("interactor.NewFilter failed: %w", err)
	}

	i := interactor.New()

	langToRepoStatePairs, err := i.GetRepoStates(cmd.Context(), *filter)
	if err != nil {
		return fmt.Errorf("interactor.GetRepoStates failed: %w", err)
	}

	rainbow := []color.Attribute{
		color.FgBlue, color.FgMagenta, color.FgCyan,
	}
	z := 0
	for lang, rsp := range langToRepoStatePairs {
		for _, state := range rsp {
			d := color.New(
				rainbow[z%(len(rainbow)-1)],
				color.Bold,
			)
			d.Print(lang + " ")

			d = color.New(color.FgWhite)
			d.Print(state.Path + " ")

			switch state.State {
			case interactor.UpToDate:
				d = color.New(color.FgGreen, color.Bold)
			case interactor.UncommittedChanges:
				d = color.New(color.FgYellow, color.Bold)
			case interactor.NotCloned:
				d = color.New(color.FgRed, color.Bold)
			default:
				d = color.New(color.FgHiMagenta, color.Bold)
			}
			d.Print(state.State.String())

			if state.Repo.Fork {
				d = color.New(color.FgHiCyan)
				d.Println(" Fork")
			} else {
				d.Println()
			}

		}
		z += 1
	}

	return nil
}
