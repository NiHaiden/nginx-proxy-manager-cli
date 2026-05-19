package main

import (
	"os"

	"github.com/NiHaiden/nginx-proxy-manager-cli/internal/npmctl"
)

func main() {
	cli := npmctl.NewCLI(os.Stdin, os.Stdout, os.Stderr, npmctl.NewOSSecretStore())
	os.Exit(cli.Run(os.Args[1:]))
}
