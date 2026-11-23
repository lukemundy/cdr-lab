package compose_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/compose"
)

type call struct {
	dir  string
	env  []string
	argv []string
}

type fake struct {
	calls []call
	ps    string
	err   error
}

func (f *fake) exec(_ context.Context, dir string, env []string, stdout, _ io.Writer, argv []string) error {
	f.calls = append(f.calls, call{dir: dir, env: env, argv: argv})

	if argv[len(argv)-1] == "json" {
		_, _ = io.WriteString(stdout, f.ps)
		return nil
	}

	return f.err
}

func project(f *fake, command string) compose.Project {
	return compose.Project{
		Command: compose.Command(command), Dir: "lab-dir", Env: []string{"A=1"},
		Stdout: io.Discard, Stderr: io.Discard, Exec: f.exec,
	}
}

// Verifies: CLI-0136
func TestCommandIsDockerComposeUnlessTheVariableGivesACommand(t *testing.T) {
	assert.Equal(t, []string{"docker", "compose"}, compose.Command(""))
	assert.Equal(t, []string{"docker-compose"}, compose.Command("docker-compose"))
	assert.Equal(t, []string{"podman", "compose"}, compose.Command(" podman  compose "))
}

func TestUpRunsComposeInTheLabDirectoryAndWaits(t *testing.T) {
	f := &fake{}

	err := project(f, "docker-compose").Up(context.Background(), compose.UpOptions{
		Build: true, WaitTimeout: 60 * time.Second, Services: []string{"register"},
	})
	require.NoError(t, err)

	require.Len(t, f.calls, 1)
	assert.Equal(t, "lab-dir", f.calls[0].dir)
	assert.Equal(t, []string{"A=1"}, f.calls[0].env)
	assert.Equal(t, []string{
		"docker-compose", "up", "--detach", "--build", "--wait", "--wait-timeout", "60", "register",
	}, f.calls[0].argv)
}

// Verifies: CLI-0138
func TestUpWithNoBuildUsesTheImagesThatAreThere(t *testing.T) {
	f := &fake{}

	require.NoError(t, project(f, "").Up(context.Background(), compose.UpOptions{WaitTimeout: 1500 * time.Millisecond}))

	assert.Equal(t, []string{"docker", "compose", "up", "--detach", "--no-build", "--wait", "--wait-timeout", "2"},
		f.calls[0].argv)
}

// Verifies: CLI-0153
func TestUpThatFailsNamesEachServiceThatIsNotHealthy(t *testing.T) {
	f := &fake{
		err: errors.New("exit status 1"),
		ps: `{"Service":"postgres","State":"running","Health":"healthy","ExitCode":0}
{"Service":"register","State":"running","Health":"starting","ExitCode":0}
{"Service":"register-seed","State":"exited","Health":"","ExitCode":0}
{"Service":"register-migrate","State":"exited","Health":"","ExitCode":1}
{"Service":"mail","State":"running","Health":"","ExitCode":0}
{"Service":"wattle-bank","State":"created","Health":"","ExitCode":0}
`,
	}

	err := project(f, "").Up(context.Background(), compose.UpOptions{WaitTimeout: time.Minute})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "register, register-migrate, wattle-bank")
	assert.NotContains(t, err.Error(), "postgres")
	assert.Equal(t, []string{"docker", "compose", "ps", "--all", "--format", "json"}, f.calls[1].argv)
}

// Verifies: CLI-0154
func TestDownRemovesTheContainersOfEachProfileAndKeepsTheVolumes(t *testing.T) {
	f := &fake{}

	require.NoError(t, project(f, "").Down(context.Background(), compose.DownOptions{}))

	assert.Equal(t, []string{"docker", "compose", "--profile", "*", "down", "--remove-orphans"}, f.calls[0].argv)
}

// Verifies: CLI-0155
func TestDownWithVolumesAlsoRemovesTheVolumes(t *testing.T) {
	f := &fake{}

	require.NoError(t, project(f, "").Down(context.Background(), compose.DownOptions{Volumes: true}))

	assert.Equal(t, []string{"docker", "compose", "--profile", "*", "down", "--remove-orphans", "--volumes"},
		f.calls[0].argv)
}

func TestSetEnvReplacesAValueAndAddsAValue(t *testing.T) {
	env := compose.SetEnv([]string{"A=1", "COMPOSE_PROFILES=x", "B=2"}, "COMPOSE_PROFILES=telemetry", "C=")

	assert.Equal(t, []string{"A=1", "B=2", "COMPOSE_PROFILES=telemetry", "C="}, env)
}
