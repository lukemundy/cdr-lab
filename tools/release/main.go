// Command release gives the facts that the publications from main and the releases need. Run it from the root of
// the repository.
//
//	go run ./tools/release image-commit -service register -image ghcr.io/lukemundy/cdr-lab/register
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "release: %v\n", err)

		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "image-commit" {
		return errors.New("usage: release image-commit -service <service> -image <repository> [-rev <commit>]")
	}

	flags := flag.NewFlagSet("image-commit", flag.ContinueOnError)
	service := flags.String("service", "", "the service directory name, for example register")
	image := flags.String("image", "", "the image repository of the service, for example ghcr.io/owner/repo/register")
	rev := flags.String("rev", "HEAD", "the commit of main")
	inputsFile := flags.String("inputs", ".github/image-inputs.json", "the file of the image inputs of each service")

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	inputs, err := readInputs(*inputsFile)
	if err != nil {
		return err
	}

	pathspecs, ok := inputs[*service]
	if !ok {
		return fmt.Errorf("%s gives no image inputs for the service %q", *inputsFile, *service)
	}

	reg, err := newRegistry(http.DefaultClient, *image)
	if err != nil {
		return err
	}

	commit, err := imageCommit(ctx, gitRepo{dir: "."}, func(tag string) (bool, error) { return reg.has(ctx, tag) },
		*rev, pathspecs)
	if err != nil {
		return fmt.Errorf("%s: %w", *service, err)
	}

	_, err = fmt.Fprintln(stdout, commit)

	return err
}
