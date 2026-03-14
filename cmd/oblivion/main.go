// Command oblivion is the CLI entrypoint.
package main

import (
	"fmt"
	"os"

	"github.com/wallfacer/oblivion/internal/app"
)

func main() {
	if err := app.Run(app.Config{Output: os.Stdout}); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
