package main

import (
	"fmt"
	"os"

	"github.com/gdanko/pstree/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}

	return
}
