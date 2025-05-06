package main

import (
	"fmt"
	"os"

	"github.com/venafi/csm-opa-plugin/setup"
)

func main() {
	cmd := setup.SetupRootCommand(nil)

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
