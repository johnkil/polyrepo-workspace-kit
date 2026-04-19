package main

import (
	"os"

	"github.com/johnkil/polyrepo-workspace-kit/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
