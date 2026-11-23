package cdrlab_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/charts/internal/charttest"
)

type object = map[string]any

var lintArgs = []string{
	"--values", "ci/lint-values.yaml",
	"--api-versions", "gateway.networking.k8s.io/v1/TLSRoute",
	"--api-versions", "gateway.networking.k8s.io/v1/HTTPRoute",
}

// rendered keeps each render of the package by its arguments, because a render takes about one second. The tests do
// not change the objects.
var rendered = map[string][]object{}

func render(t *testing.T, args ...string) []object {
	t.Helper()

	key := strings.Join(args, "\x00")
	if objects, ok := rendered[key]; ok {
		return objects
	}

	objects, err := charttest.RenderDir(t, ".", args...)
	require.NoError(t, err)

	rendered[key] = objects

	return objects
}

func renderLint(t *testing.T, args ...string) []object {
	t.Helper()

	return render(t, append(slices.Clone(lintArgs), args...)...)
}

func renderErr(t *testing.T, args ...string) error {
	t.Helper()

	_, err := charttest.RenderDir(t, ".", args...)
	require.Error(t, err, args)

	return err
}

func set(values ...string) []string {
	var args []string
	for _, v := range values {
		args = append(args, "--set", v)
	}

	return args
}

func field(o any, path ...any) any {
	for _, p := range path {
		switch k := p.(type) {
		case string:
			m, ok := o.(map[string]any)
			if !ok {
				return nil
			}

			o = m[k]
		case int:
			l, ok := o.([]any)
			if !ok || k >= len(l) {
				return nil
			}

			o = l[k]
		}
	}

	return o
}

func str(o any, path ...any) string {
	v := field(o, path...)
	if v == nil {
		return ""
	}

	return fmt.Sprint(v)
}

func list(o any, path ...any) []any {
	l, _ := field(o, path...).([]any)
	return l
}

func strs(o any, path ...any) []string {
	var out []string
	for _, v := range list(o, path...) {
		out = append(out, fmt.Sprint(v))
	}

	return out
}

func ofKind(objects []object, kind string) []object {
	var found []object

	for _, o := range objects {
		if o["kind"] == kind {
			found = append(found, o)
		}
	}

	return found
}

func named(t *testing.T, objects []object, kind, name string) object {
	t.Helper()

	for _, o := range ofKind(objects, kind) {
		if str(o, "metadata", "name") == name {
			return o
		}
	}

	require.FailNow(t, "no "+kind+" "+name)

	return nil
}

func podSpec(o object) any {
	if o["kind"] == "CronJob" {
		return field(o, "spec", "jobTemplate", "spec", "template", "spec")
	}

	return field(o, "spec", "template", "spec")
}

func workloads(objects []object) []object {
	var found []object

	for _, kind := range []string{"Deployment", "Job", "CronJob"} {
		found = append(found, ofKind(objects, kind)...)
	}

	return found
}

func container(t *testing.T, o object, name string) any {
	t.Helper()

	spec := podSpec(o)
	for _, c := range append(list(spec, "initContainers"), list(spec, "containers")...) {
		if str(c, "name") == name {
			return c
		}
	}

	require.FailNow(t, "no container "+name)

	return nil
}

// env gives the value of each variable, and secret:key for a variable from a Secret.
func env(c any) map[string]string {
	vars := map[string]string{}

	for _, e := range list(c, "env") {
		if ref := field(e, "valueFrom", "secretKeyRef"); ref != nil {
			vars[str(e, "name")] = str(ref, "name") + ":" + str(ref, "key")
		} else {
			vars[str(e, "name")] = str(e, "value")
		}
	}

	return vars
}

func envNames(c any) []string {
	var names []string
	for _, e := range list(c, "env") {
		names = append(names, str(e, "name"))
	}

	slices.Sort(names)

	return names
}

// mountedSecrets gives the Secret of each volume that a container of the pod mounts.
func mountedSecrets(o object) []string {
	spec := podSpec(o)
	volumes := map[string]string{}

	for _, v := range list(spec, "volumes") {
		if s := str(v, "secret", "secretName"); s != "" {
			volumes[str(v, "name")] = s
		}
	}

	var secrets []string

	for _, c := range append(list(spec, "initContainers"), list(spec, "containers")...) {
		for _, m := range list(c, "volumeMounts") {
			if s, ok := volumes[str(m, "name")]; ok && !slices.Contains(secrets, s) {
				secrets = append(secrets, s)
			}
		}
	}

	slices.Sort(secrets)

	return secrets
}
