package del

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sgit/internal/interactor"
	"sgit/internal/tui"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	local  = "local"
	remote = "remote"
	both   = "both"
)

var (
	langs, states, names *string
	forks                *bool

	Cmd = &cobra.Command{
		Use:   "delete",
		Short: "delete repositories",
		Long:  "delete repositories",
		RunE:  run,
	}
)

func init() {
	langs = Cmd.PersistentFlags().StringP("lang", "l", "", "comma-separated string of languages to target")
	forks = Cmd.PersistentFlags().BoolP("fork", "f", false, "target forked or non-forked repos")
	states = Cmd.PersistentFlags().StringP("state", "s", "", "comma-separated list of states to target")
	names = Cmd.PersistentFlags().StringP("name", "n", "", "comma-separated list of repo names to target")
}

func run(cmd *cobra.Command, args []string) error {
	repos, err := getTargets(cmd, args)
	if err != nil {
		return err
	}

	if proceed := showPrompt(repos); !proceed {
		return nil
	}

	selection := showLocalRemotePrompt()
	if selection == "" {
		return nil
	}

	err = deleteRepos(cmd, repos, selection)
	return err
}

func getTargets(cmd *cobra.Command, args []string) ([]interactor.Repo, error) {
	if len(args) == 0 {
		var forksFlag *bool
		if cmd.Flags().Changed("forks") {
			forksFlag = forks
		}

		filter, err := interactor.NewFilter(*langs, *states, *names, forksFlag)
		if err != nil {
			return nil, fmt.Errorf("interactor.NewFilter failed: %w", err)
		}

		i := interactor.New()

		langToRepoStatePairs, err := i.GetRepoStates(cmd.Context(), *filter)
		if err != nil {
			return nil, fmt.Errorf("interactor.GetRepoStates failed: %w", err)
		}

		rv := make([]interactor.Repo, 0)
		for _, rsps := range langToRepoStatePairs {
			for _, rsp := range rsps {
				rv = append(rv, rsp.Repo)
			}
		}

		return rv, nil
	}

	i := interactor.New()

	type result struct {
		interactor.Repo
		err    error
		exists bool
	}

	results := make(chan result, len(args))
	errs := make([]error, 0)
	var wg sync.WaitGroup
	for _, arg := range args {
		wg.Add(1)

		go func(arg string) {
			defer wg.Done()

			repo := normalize(arg)
			lang, err := i.GetPrimaryLanguageForRepo(cmd.Context(), repo.Owner, repo.Name)
			if err != nil {
				results <- result{
					repo,
					fmt.Errorf("i.GetPrimaryLanguageForRepo failed: %w", err),
					false,
				}
				return
			}
			repo.Language = strings.ToLower(lang)

			exists, err := i.Exists(repo)
			if err != nil {
				results <- result{
					repo,
					fmt.Errorf("i.Exists failed: %w", err),
					false,
				}
				return
			}

			results <- result{
				repo,
				nil,
				exists,
			}
		}(arg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	rv := make([]interactor.Repo, 0)
	for result := range results {
		if result.err != nil {
			errs = append(errs, result.err)
			continue
		} else if result.exists {
			continue
		} else {
			rv = append(rv, result.Repo)
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return rv, nil
}

func normalize(arg string) interactor.Repo {
	p := strings.Split(arg, "/")

	name, username := "", ""
	if len(p) > 1 {
		name = strings.Split(p[len(p)-1], ".")[0]

		p := strings.Split(p[len(p)-2], ":")
		username = p[len(p)-1]
	} else {
		name = strings.Split(p[len(p)-1], ".")[0]
		username = os.Getenv("GITHUB_USERNAME")
	}

	return interactor.Repo{
		Name:  name,
		Owner: username,
		URL:   fmt.Sprintf("https://github.com/%s/%s", name, username),
	}
}

func showPrompt(repos []interactor.Repo) bool {
	if len(repos) == 0 {
		return false
	}

	tui.Output(repos)

	reader := bufio.NewReader(os.Stdin)

	proceed := false
	for {
		msg := "You're about to delete 1 repo"
		if len(repos) > 1 {
			msg = fmt.Sprintf("You're about to delete %d repos", len(repos))
		}
		msg += ", would you like to proceed? (y/n): "
		d := color.New(color.FgGreen, color.Bold)
		d.Print(msg)

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "y" {
			proceed = true
			break
		} else if input == "n" {
			break
		} else {
			fmt.Println("Invalid input. Please enter y or n.")
		}
	}

	return proceed
}

func showLocalRemotePrompt() string {
	reader := bufio.NewReader(os.Stdin)
	for {
		d := color.New(color.FgGreen, color.Bold)
		d.Print("What repo would you like to delete? (local/remote/both): ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case local:
			return input
		case remote:
			return input
		case both:
			return input
		default:
			fmt.Println("Invalid input. Please enter local, remote or both.")
		}
	}

	return ""
}

func deleteRepos(cmd *cobra.Command, repos []interactor.Repo, target string) error {
	i := interactor.New()

	tui.PrintProgress(0.0)
	complete := 0
	errs := make([]error, 0)
	results := make(chan error, len(repos))
	var wg sync.WaitGroup
	for _, repo := range repos {
		wg.Add(1)
		go func(r interactor.Repo) {
			defer wg.Done()

			errs := make([]error, 0)
			if target == remote || target == both {
				if err := i.DeleteRemote(cmd.Context(), r); err != nil {
					errs = append(errs, err)
				}
			}

			if target == local || target == both {
				if err := i.DeleteLocal(r); err != nil {
					errs = append(errs, err)
				}
			}

			if len(errs) > 0 {
				results <- errors.Join(errs...)
			} else {
				results <- nil
			}
		}(repo)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for err := range results {
		complete += 1
		tui.PrintProgress(float64(complete) / float64(len(repos)))
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil

}
