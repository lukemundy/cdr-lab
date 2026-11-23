// Package compose runs Docker Compose as a separate program, in the lab directory (ADR 0038).
package compose

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Exec runs the program argv[0] with the arguments argv[1:].
type Exec func(ctx context.Context, dir string, env []string, stdout, stderr io.Writer, argv []string) error

func OSExec(ctx context.Context, dir string, env []string, stdout, stderr io.Writer, argv []string) error {
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...) //nolint:gosec // G204: the user gives the Compose command.
	cmd.Dir, cmd.Env, cmd.Stdout, cmd.Stderr = dir, env, stdout, stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", strings.Join(argv, " "), err)
	}

	return nil
}

// Command gives the Compose command from the value of CDRLAB_COMPOSE (CLI-0136).
func Command(value string) []string {
	if fields := strings.Fields(value); len(fields) > 0 {
		return fields
	}

	return []string{"docker", "compose"}
}

// SetEnv gives env with each KEY=value of set, in place of an earlier value of the same key.
func SetEnv(env []string, set ...string) []string {
	out := slices.DeleteFunc(slices.Clone(env), func(kv string) bool {
		key, _, _ := strings.Cut(kv, "=")

		return slices.ContainsFunc(set, func(s string) bool { return strings.HasPrefix(s, key+"=") })
	})

	return append(out, set...)
}

type Project struct {
	Command        []string
	Dir            string
	Env            []string
	Stdout, Stderr io.Writer
	Exec           Exec
}

type UpOptions struct {
	Build       bool
	WaitTimeout time.Duration
	Services    []string
}

// Up starts the services and waits until each is healthy or has completed. When the wait fails, the error names
// each service that is not healthy (CLI-0153).
func (p Project) Up(ctx context.Context, opts UpOptions) error {
	build := "--no-build"
	if opts.Build {
		build = "--build"
	}

	seconds := strconv.Itoa(int(math.Ceil(opts.WaitTimeout.Seconds())))
	args := append([]string{"up", "--detach", build, "--wait", "--wait-timeout", seconds}, opts.Services...)

	err := p.run(ctx, p.Stdout, args...)
	if err == nil {
		return nil
	}

	names, psErr := p.unhealthy(ctx)
	if psErr != nil || len(names) == 0 {
		return err
	}

	return fmt.Errorf("these services are not healthy: %s: %w", strings.Join(names, ", "), err)
}

type DownOptions struct {
	Volumes bool
}

// Down stops and removes the containers of each profile (CLI-0154, CLI-0155).
func (p Project) Down(ctx context.Context, opts DownOptions) error {
	args := []string{"--profile", "*", "down", "--remove-orphans"}
	if opts.Volumes {
		args = append(args, "--volumes")
	}

	return p.run(ctx, p.Stdout, args...)
}

func (p Project) run(ctx context.Context, stdout io.Writer, args ...string) error {
	return p.Exec(ctx, p.Dir, p.Env, stdout, p.Stderr, append(slices.Clone(p.Command), args...))
}

// container is an object of compose ps --format json, which gives one object on each line.
type container struct {
	Service  string `json:"Service"`
	State    string `json:"State"`
	Health   string `json:"Health"`
	ExitCode int    `json:"ExitCode"`
}

func (p Project) unhealthy(ctx context.Context) ([]string, error) {
	var out bytes.Buffer
	if err := p.run(ctx, &out, "ps", "--all", "--format", "json"); err != nil {
		return nil, err
	}

	var names []string

	lines := bufio.NewScanner(&out)
	for lines.Scan() {
		var c container
		if err := json.Unmarshal(lines.Bytes(), &c); err != nil {
			return nil, err
		}

		if !c.ok() {
			names = append(names, c.Service)
		}
	}

	slices.Sort(names)

	return slices.Compact(names), lines.Err()
}

// ok is the condition of compose up --wait: healthy, or running with no health check, or completed with status 0.
func (c container) ok() bool {
	switch c.State {
	case "running":
		return c.Health == "" || c.Health == "healthy"
	case "exited":
		return c.ExitCode == 0
	default:
		return false
	}
}
