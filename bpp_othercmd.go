package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func showVersion() error  {
	fmt.Printf("version: %s\n", version)
	return nil
}

func clean(ctx context.Context) error{
	output := filepath.Join(workDir, OutputBuildpack)
	_ = os.RemoveAll(output)
	return nil
}
