package main

import (
	"context"
	"fmt"
	"os"

	"github.com/lwmacct/251202-mdtoc/internal/appcmd/root"
)

func main() {
	if err := root.Command.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
