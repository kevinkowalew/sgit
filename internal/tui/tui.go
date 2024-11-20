package tui

import (
	"fmt"
	"sgit/internal/interactor"
	"strings"

	"github.com/fatih/color"
)

func Output(repos []interactor.Repo) {
	rainbow := []color.Attribute{
		color.FgBlue, color.FgMagenta, color.FgCyan,
	}

	langIndex := make(map[string]int, 0)
	for _, repo := range repos {
		z, ok := langIndex[repo.Language]
		if !ok {
			langIndex[repo.Language] = len(langIndex)
			z += 1
		}

		d := color.New(
			rainbow[z%(len(rainbow)-1)],
			color.Bold,
		)
		d.Print(repo.Language + " ")

		d = color.New(color.FgWhite)
		d.Println(
			fmt.Sprintf("%s/%s ", repo.Owner, repo.Name),
		)
	}
}

func PrintProgress(percentage float64) {
	width := 50
	barWidth := int(percentage * float64(width))
	bar := strings.Repeat("#", barWidth) + strings.Repeat(" ", width-barWidth)
	fmt.Printf("\r[%s] %3.0f%%", bar, percentage*100)

	if percentage == 1.0 {
		fmt.Printf("\r%s\r", strings.Repeat(" ", width+8))
	}
}
