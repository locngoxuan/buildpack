package main

import (
	"context"
	"fmt"
	"os"
)

func showVersion() error {
	fmt.Printf("version: %s\n", version)
	return nil
}

func clean(ctx context.Context) error {
	_ = os.RemoveAll(outputDir)
	return nil
}
