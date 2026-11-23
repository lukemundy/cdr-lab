// Command release gives the facts that the publications from main and the releases need, and checks a release. Run it
// from the root of the repository.
//
//	go run ./tools/release image-commit -service register -image ghcr.io/lukemundy/cdr-lab/register
//	go run ./tools/release plan -mode check -image-prefix ghcr.io/lukemundy/cdr-lab -github-repo lukemundy/cdr-lab
//	go run ./tools/release notes -plan plan.json -chart-digests <directory>
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const usage = "usage: release image-commit|plan|notes [flags]; release <command> -h gives the flags"

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "release: %v\n", err)

		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New(usage)
	}

	switch args[0] {
	case "image-commit":
		return runImageCommit(ctx, args[1:], stdout)
	case "plan":
		return runPlan(ctx, args[1:], stdout)
	case "notes":
		return runNotes(args[1:], stdout)
	default:
		return errors.New(usage)
	}
}

func runImageCommit(ctx context.Context, args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("image-commit", flag.ContinueOnError)
	service := flags.String("service", "", "the service directory name, for example register")
	image := flags.String("image", "", "the image repository of the service, for example ghcr.io/owner/repo/register")
	rev := flags.String("rev", "HEAD", "the commit of main")
	inputsFile := flags.String("inputs", ".github/image-inputs.json", "the file of the image inputs of each service")

	if err := flags.Parse(args); err != nil {
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

func runPlan(ctx context.Context, args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("plan", flag.ContinueOnError)
	rev := flags.String("rev", "HEAD", "the release commit; it must be checked out")
	mode := flags.String("mode", string(modeCheck), "check, rehearse or release")
	runID := flags.String("run-id", "", "the number of the workflow run; a rehearsal puts it in its names")
	imagePrefix := flags.String(
		"image-prefix",
		"",
		"the repository prefix of the images, for example ghcr.io/owner/repo",
	)
	githubRepo := flags.String("github-repo", "", "the GitHub repository, for example owner/repo")
	mainRef := flags.String("main", "origin/main", "the ref of main")
	inputsFile := flags.String("inputs", ".github/image-inputs.json", "the file of the image inputs of each image")
	template := flags.String("changelog-template", "tools/release/changelog.tera", "the cog template of the changelog")
	workflow := flags.String("ci-workflow", "ci.yml", "the file name of the CI workflow")
	api := flags.String("github-api", envOr("GITHUB_API_URL", "https://api.github.com"), "the GitHub REST API URL")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *imagePrefix == "" || *githubRepo == "" {
		return errors.New("plan needs -image-prefix and -github-repo")
	}

	inputs, err := readInputs(*inputsFile)
	if err != nil {
		return err
	}

	templatePath, err := filepath.Abs(*template)
	if err != nil {
		return err
	}

	ci := ciChecker{
		client: http.DefaultClient, api: strings.TrimSuffix(*api, "/"), repo: *githubRepo, workflow: *workflow,
		token: os.Getenv("GITHUB_TOKEN"),
	}

	p, err := makePlan(ctx, planConfig{
		repo:        gitRepo{dir: "."},
		mainRef:     *mainRef,
		mode:        releaseMode(*mode),
		runID:       *runID,
		imagePrefix: *imagePrefix,
		githubRepo:  *githubRepo,
		inputs:      inputs,
		template:    templatePath,
		ci:          ci.check,
		client:      http.DefaultClient,
	}, *rev)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")

	return enc.Encode(p)
}

func runNotes(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("notes", flag.ContinueOnError)
	planFile := flags.String("plan", "", "the output of plan")
	digestDir := flags.String("chart-digests", "", "a directory with one file for each chart, with the chart digest")

	if err := flags.Parse(args); err != nil {
		return err
	}

	data, err := os.ReadFile(*planFile)
	if err != nil {
		return fmt.Errorf("read the plan: %w", err)
	}

	var p releasePlan
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("read the plan %s: %w", *planFile, err)
	}

	digests := map[string]string{}

	if *digestDir != "" {
		for _, chart := range p.Charts {
			//nolint:gosec // The directory is a flag of a repository tool that a maintainer runs.
			digest, err := os.ReadFile(filepath.Join(*digestDir, chart))
			if err != nil {
				return fmt.Errorf("read the digest of the chart %s: %w", chart, err)
			}

			digests[chart] = strings.TrimSpace(string(digest))
		}
	}

	return writeNotes(stdout, p, digests)
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}

	return fallback
}
