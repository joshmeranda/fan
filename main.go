package main

import (
	"fmt"
	"os"

	"github.com/joshmeranda/fan/cmd"
)

func main() {
	app := cmd.App()

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
}
