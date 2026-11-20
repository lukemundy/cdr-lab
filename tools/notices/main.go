// Command notices writes the licence notices of the Go standard library and of each third-party module that the
// named packages link, for the image of a service. Run it with the Go toolchain and the module cache of the build.
//
//	go run ./tools/notices ./services/register/cmd/register > THIRD-PARTY-NOTICES.txt
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "notices: %v\n", err)

		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("name the packages of the binary, for example ./services/register/cmd/register")
	}

	ctx := context.Background()

	env, err := exec.CommandContext(ctx, "go", "env", "GOROOT", "GOVERSION").Output()
	if err != nil {
		return fmt.Errorf("go env: %w", err)
	}

	root, version, _ := strings.Cut(strings.TrimSpace(string(env)), "\n")

	//nolint:gosec // The arguments are package patterns after "--", so go list cannot read them as flags.
	listed, err := exec.CommandContext(ctx, "go", append([]string{"list", "-deps", "-json", "--"}, args...)...).Output()
	if err != nil {
		return fmt.Errorf("go list: %w", err)
	}

	pkgs, err := decodePackages(listed)
	if err != nil {
		return err
	}

	return write(stdout, pkgs, toolchain{Root: root, Version: version})
}

func decodePackages(listed []byte) ([]listedPackage, error) {
	var pkgs []listedPackage

	dec := json.NewDecoder(bytes.NewReader(listed))

	for {
		var p listedPackage

		err := dec.Decode(&p)
		if errors.Is(err, io.EOF) {
			return pkgs, nil
		}

		if err != nil {
			return nil, fmt.Errorf("read the output of go list: %w", err)
		}

		pkgs = append(pkgs, p)
	}
}
