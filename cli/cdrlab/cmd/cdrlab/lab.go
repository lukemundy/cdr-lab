package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/certlist"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/compose"
)

const (
	composeFile = "compose.yaml"
	pkiDir      = ".lab/pki"
)

type labFlags struct {
	dir string
}

func newLabFlags(name string) (*labFlags, *flag.FlagSet) {
	fs := newFlags(name)
	l := &labFlags{}
	fs.StringVar(&l.dir, "dir", ".", "the lab directory")

	return l, fs
}

func (l *labFlags) project(stdout io.Writer, env []string, d deps) compose.Project {
	return compose.Project{
		Command: compose.Command(d.getenv("CDRLAB_COMPOSE")), Dir: l.dir, Env: env,
		Stdout: stdout, Stderr: d.stderr, Exec: d.compose,
	}
}

func up(ctx context.Context, args []string, stdout io.Writer, d deps) error {
	lab, fs := newLabFlags("up")
	source := fs.Bool("source", false, "use the compose project of the lab directory, and build the images")
	allBanks := fs.Bool("all-banks", false, "also start the three other preset brands (compose profile all-banks)")
	noBuild := fs.Bool("no-build", false, "use the images that are there, with no build")
	noTelemetry := fs.Bool("no-telemetry", false, "start no telemetry backend, and export no telemetry")
	waitTimeout := fs.Duration("wait-timeout", 60*time.Second,
		"the maximum time of each wait until the services are healthy")

	if err := parseFlags(fs, args); err != nil {
		return err
	}

	if !*source {
		return fmt.Errorf(
			"%w: this version has no release mode: in a checkout of the repository, use -source",
			errUsage,
		)
	}

	if *waitTimeout <= 0 {
		return fmt.Errorf("%w: -wait-timeout must be more than 0", errUsage)
	}

	list, err := certlist.Load(lab.dir)
	if err != nil {
		return err
	}

	project := lab.project(stdout, composeEnv(d.environ(), *allBanks, *noTelemetry), d)
	opts := compose.UpOptions{Build: !*noBuild, WaitTimeout: *waitTimeout}

	if err := makeCertificates(ctx, lab.dir, list, stdout, d); err != nil {
		return err
	}

	// The Register issues each CDR certificate, so it starts first.
	register := opts
	register.Services = []string{"register"}

	if err := project.Up(ctx, register); err != nil {
		return err
	}

	if err := issueCDRCertificates(ctx, lab.dir, list, stdout, d); err != nil {
		return err
	}

	opts.Services = fs.Args()

	return project.Up(ctx, opts)
}

// composeEnv sets the profiles, so that an inherited value cannot change them (CLI-0151, CLI-0152).
func composeEnv(environ []string, allBanks, noTelemetry bool) []string {
	var profiles []string
	if !noTelemetry {
		profiles = append(profiles, "telemetry")
	}

	if allBanks {
		profiles = append(profiles, "all-banks")
	}

	set := []string{"COMPOSE_PROFILES=" + strings.Join(profiles, ",")}
	if noTelemetry {
		// An empty endpoint turns the telemetry export of each service off (compose.yaml).
		set = append(set, "LAB_OTLP_ENDPOINT=")
	}

	return compose.SetEnv(environ, set...)
}

// makeCertificates does steps 2 and 3 of the start (CLI-0146).
func makeCertificates(ctx context.Context, labDir string, list certlist.List, stdout io.Writer, d deps) error {
	pki := filepath.Join(labDir, pkiDir)

	if err := caInitDir(pki, stdout, d); err != nil {
		return err
	}

	for _, p := range list.Public {
		// The containers read the keys as a different user (CLI-0149).
		args := append([]string{
			"-dir", pki, "-key-readable-by-all",
			"-cert", filepath.Join(labDir, p.Cert), "-key", filepath.Join(labDir, p.Key),
		}, p.Names...)

		if err := publicCertIssue(ctx, args, stdout, d); err != nil {
			return fmt.Errorf("public certificate %s: %w", p.Cert, err)
		}
	}

	return nil
}

// issueCDRCertificates does step 5 of the start, with the flags of cert issue (CLI-0194).
func issueCDRCertificates(ctx context.Context, labDir string, list certlist.List, stdout io.Writer, d deps) error {
	getenv := d.getenv
	d.getenv = func(key string) string {
		if key == "REGISTER_OPERATOR_TOKEN" {
			return list.Register.OperatorToken
		}

		return getenv(key)
	}

	for _, c := range list.CDR {
		args := append([]string{"-register-url", list.Register.URL, "-key-readable-by-all"}, c.Flags(labDir)...)

		if err := certIssue(ctx, args, stdout, d); err != nil {
			return fmt.Errorf("CDR certificate %s: %w", c.Cert, err)
		}
	}

	return nil
}

func down(ctx context.Context, args []string, stdout io.Writer, d deps) error {
	lab, found, err := existingLab("down", args, stdout)
	if err != nil || !found {
		return err
	}

	return lab.project(stdout, d.environ(), d).Down(ctx, compose.DownOptions{})
}

func reset(ctx context.Context, args []string, stdout io.Writer, d deps) error {
	lab, found, err := existingLab("reset", args, stdout)
	if err != nil || !found {
		return err
	}

	if err := lab.project(stdout, d.environ(), d).Down(ctx, compose.DownOptions{Volumes: true}); err != nil {
		return err
	}

	return os.RemoveAll(filepath.Join(lab.dir, pkiDir))
}

// existingLab gives found false, and says so on stdout, when the lab directory has no compose file (CLI-0157).
func existingLab(name string, args []string, stdout io.Writer) (*labFlags, bool, error) {
	lab, fs := newLabFlags(name)
	if err := parse(fs, args, nil); err != nil {
		return nil, false, err
	}

	_, err := os.Stat(filepath.Join(lab.dir, composeFile))
	if errors.Is(err, os.ErrNotExist) {
		_, err = fmt.Fprintf(stdout, "%s has no lab: no change\n", lab.dir)
		return nil, false, err
	}

	return lab, err == nil, err
}
