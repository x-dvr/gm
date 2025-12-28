/*
Copyright © 2025 DENIS RODIN <denis.rodin@proton.me>
*/
package toolchain

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
)

const goDevHost = "go.dev"

func GetLatestVersion() (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://%s/VERSION?m=text", goDevHost))
	if err != nil {
		return "", fmt.Errorf("get latest Go version: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("get latest Go version: HTTP code %d", resp.StatusCode)
	}
	version, err := bufio.NewReader(resp.Body).ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("extract latest Go version: %w", err)
	}

	return strings.TrimSpace(version), nil
}
