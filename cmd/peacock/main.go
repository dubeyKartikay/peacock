package main

import (
	"fmt"
	"os"

	"github.com/dubeyKartikay/peacock/internal/cli"
)

func main() {
	if err := cli.Execute(os.Stdin); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
