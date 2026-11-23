// Package charttest renders the chart of the working directory with helm template, for the chart tests.
package charttest

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

type Object struct {
	Kind     string `yaml:"kind"`
	Metadata struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec struct {
		Replicas *int `yaml:"replicas"`
		Strategy struct {
			Type string `yaml:"type"`
		} `yaml:"strategy"`
		Ports    []Port `yaml:"ports"`
		Template struct {
			Spec PodSpec `yaml:"spec"`
		} `yaml:"template"`
	} `yaml:"spec"`
}

type PodSpec struct {
	AutomountServiceAccountToken *bool `yaml:"automountServiceAccountToken"`
	SecurityContext              struct {
		RunAsNonRoot   bool `yaml:"runAsNonRoot"`
		SeccompProfile struct {
			Type string `yaml:"type"`
		} `yaml:"seccompProfile"`
	} `yaml:"securityContext"`
	InitContainers []Container `yaml:"initContainers"`
	Containers     []Container `yaml:"containers"`
	Volumes        []Volume    `yaml:"volumes"`
}

type Volume struct {
	Name   string `yaml:"name"`
	Secret struct {
		SecretName string `yaml:"secretName"`
	} `yaml:"secret"`
	ConfigMap struct {
		Name string `yaml:"name"`
	} `yaml:"configMap"`
}

type Container struct {
	Name  string   `yaml:"name"`
	Image string   `yaml:"image"`
	Args  []string `yaml:"args"`
	Env   []struct {
		Name      string `yaml:"name"`
		Value     string `yaml:"value"`
		ValueFrom struct {
			SecretKeyRef struct {
				Name string `yaml:"name"`
				Key  string `yaml:"key"`
			} `yaml:"secretKeyRef"`
		} `yaml:"valueFrom"`
	} `yaml:"env"`
	Ports           []Port `yaml:"ports"`
	SecurityContext struct {
		AllowPrivilegeEscalation *bool `yaml:"allowPrivilegeEscalation"`
		ReadOnlyRootFilesystem   bool  `yaml:"readOnlyRootFilesystem"`
		Capabilities             struct {
			Drop []string `yaml:"drop"`
		} `yaml:"capabilities"`
	} `yaml:"securityContext"`
	LivenessProbe  Probe         `yaml:"livenessProbe"`
	ReadinessProbe Probe         `yaml:"readinessProbe"`
	VolumeMounts   []VolumeMount `yaml:"volumeMounts"`
}

type VolumeMount struct {
	Name      string `yaml:"name"`
	MountPath string `yaml:"mountPath"`
	SubPath   string `yaml:"subPath"`
	ReadOnly  bool   `yaml:"readOnly"`
}

type Port struct {
	Name          string `yaml:"name"`
	Port          int    `yaml:"port"`
	ContainerPort int    `yaml:"containerPort"`
}

type Probe struct {
	HTTPGet struct {
		Path string `yaml:"path"`
		Port int    `yaml:"port"`
	} `yaml:"httpGet"`
}

// Vars gives the value of each variable, and secret:key for a variable from a Secret.
func (c Container) Vars() map[string]string {
	vars := map[string]string{}

	for _, e := range c.Env {
		if ref := e.ValueFrom.SecretKeyRef; ref.Name != "" {
			vars[e.Name] = ref.Name + ":" + ref.Key
		} else {
			vars[e.Name] = e.Value
		}
	}

	return vars
}

// AllowlistVars gives the names of the Host allowlist variables, in order.
func (c Container) AllowlistVars() []string {
	var names []string

	for _, e := range c.Env {
		if strings.HasSuffix(e.Name, "_ALLOWED_HOSTS") {
			names = append(names, e.Name)
		}
	}

	slices.Sort(names)

	return names
}

// Render uses the release name "test" and ci/test-values.yaml. The error is the stderr of helm.
func Render(t *testing.T, set ...string) ([]Object, error) {
	t.Helper()

	return render(t, ".", nil, set)
}

// RenderPackaged renders the chart that helm package gives with the appVersion, as CI publishes it.
func RenderPackaged(t *testing.T, appVersion string, set ...string) ([]Object, error) {
	t.Helper()

	dir := t.TempDir()

	//nolint:gosec // The appVersion comes from the chart tests.
	out, err := exec.CommandContext(t.Context(), "helm", "package", ".", "--app-version", appVersion,
		"--destination", dir).CombinedOutput()
	require.NoError(t, err, string(out))

	packages, err := filepath.Glob(filepath.Join(dir, "*.tgz"))
	require.NoError(t, err)
	require.Len(t, packages, 1)

	return render(t, packages[0], nil, set)
}

// RenderDir renders the chart in dir with its ci/test-values.yaml and the more arguments of helm template, and gives
// each object as a map.
func RenderDir(t *testing.T, dir string, args ...string) ([]map[string]any, error) {
	t.Helper()

	readChartFiles(t, dir)

	args = append([]string{"template", "test", dir, "--values", filepath.Join(dir, "ci", "test-values.yaml")}, args...)

	var stdout, stderr bytes.Buffer

	//nolint:gosec // The chart directory and the arguments come from the chart tests.
	cmd := exec.CommandContext(t.Context(), "helm", args...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	if err := cmd.Run(); err != nil {
		return nil, errors.New(stderr.String())
	}

	var objects []map[string]any

	dec := yaml.NewDecoder(&stdout)

	for {
		var o map[string]any

		err := dec.Decode(&o)
		if errors.Is(err, io.EOF) {
			return objects, nil
		}

		require.NoError(t, err)

		if o != nil {
			objects = append(objects, o)
		}
	}
}

func render(t *testing.T, chart string, files []string, set []string) ([]Object, error) {
	t.Helper()

	readChartFiles(t, ".")

	args := []string{"template", "test", chart, "--values", "ci/test-values.yaml"}
	for _, f := range files {
		args = append(args, "--values", f)
	}

	for _, s := range set {
		args = append(args, "--set", s)
	}

	var stdout, stderr bytes.Buffer

	//nolint:gosec // The chart, the values files and the values come from the chart tests.
	cmd := exec.CommandContext(t.Context(), "helm", args...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	if err := cmd.Run(); err != nil {
		return nil, errors.New(stderr.String())
	}

	var objects []Object

	dec := yaml.NewDecoder(&stdout)

	for {
		var o Object

		err := dec.Decode(&o)
		if errors.Is(err, io.EOF) {
			return objects, nil
		}

		require.NoError(t, err)

		if o.Kind != "" {
			objects = append(objects, o)
		}
	}
}

// readChartFiles reads each file of the chart, because go test caches a result until a file that the test opens
// changes, and it does not see the files that helm opens.
func readChartFiles(t *testing.T, dir string) {
	t.Helper()

	chart := os.DirFS(dir)

	err := fs.WalkDir(chart, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		_, err = fs.ReadFile(chart, name)

		return err
	})
	require.NoError(t, err)
}

func MustRender(t *testing.T, set ...string) []Object {
	t.Helper()

	objects, err := Render(t, set...)
	require.NoError(t, err)

	return objects
}

// MustRenderValues renders with a values file on top of ci/test-values.yaml.
func MustRenderValues(t *testing.T, file string, set ...string) []Object {
	t.Helper()

	objects, err := render(t, ".", []string{file}, set)
	require.NoError(t, err)

	return objects
}

func Only(t *testing.T, objects []Object, kind string) Object {
	t.Helper()

	var found []Object

	for _, o := range objects {
		if o.Kind == kind {
			found = append(found, o)
		}
	}

	require.Len(t, found, 1, kind)

	return found[0]
}

func Pod(t *testing.T, objects []Object) PodSpec {
	t.Helper()

	return Only(t, objects, "Deployment").Spec.Template.Spec
}

func Main(t *testing.T, objects []Object) Container {
	t.Helper()

	spec := Pod(t, objects)
	require.Len(t, spec.Containers, 1)

	return spec.Containers[0]
}

func ServicePorts(t *testing.T, objects []Object) map[string]int {
	t.Helper()

	ports := map[string]int{}
	for _, p := range Only(t, objects, "Service").Spec.Ports {
		ports[p.Name] = p.Port
	}

	return ports
}
