//go:build conformance

package dataholder

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

const waitForReady = time.Minute

// The data holder gives no data until it has the statuses of the Register.
func TestMain(m *testing.M) {
	if err := ready(os.Getenv("DATAHOLDER_PROBE_URL") + "/readyz"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func ready(url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), waitForReady)
	defer cancel()

	for {
		//nolint:gosec // G704: the URL is the probe URL of the data holder in the lab configuration.
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		if probeOK(req) {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("the data holder was not ready after %s: %s", waitForReady, url)
		case <-time.After(time.Second):
		}
	}
}

func probeOK(req *http.Request) bool {
	//nolint:gosec // G704: the URL is the probe URL of the data holder in the lab configuration.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}

	_ = resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
