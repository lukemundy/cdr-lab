package cdrlab_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/charts/internal/charttest"
)

// The umbrella chart builds the containers itself, so a change to a service chart must also change it.
func TestTheContainersHaveTheVariablesOfTheServiceCharts(t *testing.T) {
	for _, c := range []struct {
		desc                           string
		umbrella, register, dataholder []string
	}{
		{
			desc:       "no optional part",
			umbrella:   nil,
			register:   set("telemetry.otlpEndpoint=", "seed.configMap=seed"),
			dataholder: set("telemetry.otlpEndpoint="),
		},
		{
			desc:     "each optional part",
			umbrella: lintArgs,
			register: set("seed.configMap=seed", "pki.publicRootSecret=root", "console.enabled=true"),
			dataholder: set("pki.publicRootSecret=root", "crlMaxAge=30s", "console.enabled=true", "bankSite.enabled=true",
				"bankSite.allowedHosts={bank.test:8085}"),
		},
	} {
		umbrella := render(t, c.umbrella...)

		for _, s := range []struct {
			chart, umbrellaName string
			args                []string
		}{
			{"../register", "test-cdr-lab-register", c.register},
			{"../dataholder", "test-cdr-lab-wattle-bank", c.dataholder},
		} {
			objects, err := charttest.RenderDir(t, s.chart, s.args...)
			require.NoError(t, err)

			deployments := ofKind(objects, "Deployment")
			require.Len(t, deployments, 1)

			for _, name := range []string{"migrate", "seed", "main"} {
				assert.Equal(t,
					envNames(container(t, deployments[0], name)),
					envNames(container(t, named(t, umbrella, "Deployment", s.umbrellaName), name)),
					"%s: %s: %s", c.desc, s.chart, name)
			}
		}
	}
}
