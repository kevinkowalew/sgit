package tui

import (
	"sgit/internal/cmd"

	"github.com/fatih/color"
)

type TUI struct {
}

func New() *TUI {
	return &TUI{}
}

func (t TUI) Handle(repos []cmd.GithubRepository) error {
	langToRepo := make(map[string][]cmd.GithubRepository, 0)
	for _, r := range repos {
		e, ok := langToRepo[r.Language]
		if !ok {
			langToRepo[r.Language] = []cmd.GithubRepository{r}
			continue
		}
		langToRepo[r.Language] = append(e, r)
	}

	rainbow := []color.Attribute{
		color.FgYellow, color.FgBlue, color.FgMagenta, color.FgCyan,
	}

	i := 0
	for lang, repos := range langToRepo {

		for _, repo := range repos {
			d := color.New(
				rainbow[i%(len(rainbow)-1)],
				color.Bold,
			)
			d.Print(lang + " ")

			d = color.New(color.FgWhite)
			d.Print(repo.Name() + ": ")

			if repo.State == cmd.UpToDate {
				d = color.New(color.FgGreen, color.Bold)
				d.Println(repo.State.String())
			} else {
				d = color.New(color.FgRed, color.Bold)
				d.Println(repo.State.String())
			}

		}
		i += 1
	}

	return nil
}
