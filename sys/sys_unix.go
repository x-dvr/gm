//go:build !windows

/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package sys

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func SetGoEnvs() error {
	path := os.Getenv("PATH")
	if path == "" {
		return ErrNoPath
	}
	homedir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir of user: %w", err)
	}

	goPath := filepath.Join(homedir, gmDir, workspace)
	goBin := filepath.Join(goPath, "bin")
	goRoot := filepath.Join(homedir, gmDir, versions, current)
	goSDKBin := filepath.Join(goRoot, "bin")

	// Detect shell from SHELL environment variable
	shell := os.Getenv("SHELL")
	isFish := strings.HasSuffix(shell, "/fish")

	if isFish {
		// Fish shell syntax
		fmt.Printf("set -gx GOPATH %s\n", goPath)
		fmt.Printf("set -gx GOBIN %s\n", goBin)
		fmt.Printf("set -gx GOROOT %s\n", goRoot)
		fmt.Printf("set -gx PATH %s $PATH\n", goSDKBin+":"+goBin)
	} else {
		// Bash/Zsh/POSIX shell syntax
		fmt.Printf("export GOPATH=%s\n", goPath)
		fmt.Printf("export GOBIN=%s\n", goBin)
		fmt.Printf("export GOROOT=%s\n", goRoot)
		fmt.Printf("export PATH=\"%s:$PATH\"\n", goSDKBin+":"+goBin)
	}
	return nil
}
