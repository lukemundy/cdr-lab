//go:build conformance

package register

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These participants of services/register/seed/example.json cannot change back, so the test passes only one time
// after a load.
const (
	statusChangeRecipient = "9a99b5d7-fd53-49e7-8a1c-152fcd8f380b"
	statusChangeWeb       = "89fcb232-f453-4f9a-9ea9-5655d9203c3c"
	statusChangeMobile    = "b442771c-9c80-4b07-8249-f4aedfe8a2e0"
	statusChangeBeta      = "385e5475-0448-46b9-a740-cec9213b95d9"
	statusChangeLegacy    = "a0dfca7e-4688-473d-b2a9-a3c2657879f0" // under an INACTIVE brand
	surrenderRecipient    = "416448ea-d963-4391-89a7-6de950b6bc91"
	surrenderWeb          = "7bae25e7-ec64-445e-9e10-e9717bd2363d"
)

const (
	recipientStatusList = "data recipient statuses"
	softwareProductList = "software product statuses"
	dataRecipientList   = "data recipients"
)

type view struct {
	etags map[string]string
	// Keyed by API, then by ID.
	recipients map[string]map[string]string
	products   map[string]map[string]string
}

func observe(t *testing.T) view {
	t.Helper()

	v := view{
		etags:      map[string]string{},
		recipients: map[string]map[string]string{},
		products:   map[string]map[string]string{},
	}
	recipientStatuses := call(t, dataRecipientsStatusPath, v3())
	v.etags[recipientStatusList] = recipientStatuses.Header.Get("ETag")
	v.recipients[recipientStatusList] = statusMap(t, recipientStatuses.body, "legalEntityId")

	productStatuses := call(t, softwareProductsStatusPath, v3())
	v.etags[softwareProductList] = productStatuses.Header.Get("ETag")
	v.products[softwareProductList] = statusMap(t, productStatuses.body, "softwareProductId")

	recipients := call(t, dataRecipientsPath, xv("4"))
	v.etags[dataRecipientList] = recipients.Header.Get("ETag")

	var body struct {
		Data []object `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recipients.body, &body))
	v.recipients[dataRecipientList] = map[string]string{}

	v.products[dataRecipientList] = map[string]string{}
	for _, dr := range body.Data {
		v.recipients[dataRecipientList][dr.string(t, "legalEntityId")] = dr.string(t, "status")
		for _, brand := range dr.objects(t, "dataRecipientBrands") {
			for _, sp := range brand.objects(t, "softwareProducts") {
				v.products[dataRecipientList][sp.string(t, "softwareProductId")] = sp.string(t, "status")
			}
		}
	}

	return v
}

func statusMap(t *testing.T, body []byte, idField string) map[string]string {
	t.Helper()

	var list struct {
		Data []map[string]string `json:"data"`
	}
	require.NoError(t, json.Unmarshal(body, &list))

	got := map[string]string{}
	for _, item := range list.Data {
		got[item[idField]] = item["status"]
	}

	return got
}

func (v view) shows(recipients, products map[string]string) bool {
	for _, list := range []string{recipientStatusList, dataRecipientList} {
		for id, want := range recipients {
			if v.recipients[list][id] != want {
				return false
			}
		}
	}

	for _, list := range []string{softwareProductList, dataRecipientList} {
		for id, want := range products {
			if v.products[list][id] != want {
				return false
			}
		}
	}

	return true
}

type statusChange struct {
	// action is the path of the operator action.
	action string
	// recipients and products give the statuses that the public APIs must show after the action, by ID.
	recipients, products map[string]string
	changedAPIs          []string
}

// change checks that the public APIs show a status change within the
// propagation time, with a new ETag on each changed API.
func change(t *testing.T, c statusChange) {
	t.Helper()
	before := observe(t)
	status, p := operate(t, c.action)
	require.Equal(t, http.StatusNoContent, status, p.Detail)

	deadline := time.Now().Add(propagation)

	after := observe(t)
	for !after.shows(c.recipients, c.products) && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)

		after = observe(t)
	}

	require.True(
		t,
		after.shows(c.recipients, c.products),
		"not visible within %s: data recipients %v, software products %v",
		propagation,
		after.recipients,
		after.products,
	)

	for _, api := range c.changedAPIs {
		assert.NotEqual(t, before.etags[api], after.etags[api], "the ETag of %s did not change", api)
	}
}

// The operator interface only makes each status change: the subject is the CDS status rules and ETags of the public
// APIs.
func TestThePublicAPIsShowEachStatusChangeWithItsCascadeAndANewETag(t *testing.T) {
	steps := []struct {
		name string
		run  func(t *testing.T)
	}{
		{name: "deactivate a software product", run: func(t *testing.T) {
			change(t, statusChange{
				action:      productAction(statusChangeRecipient, statusChangeWeb, "deactivate"),
				products:    map[string]string{statusChangeWeb: "INACTIVE"},
				changedAPIs: []string{softwareProductList, dataRecipientList},
			})
		}},
		{name: "activate a software product", run: func(t *testing.T) {
			change(t, statusChange{
				action:      productAction(statusChangeRecipient, statusChangeWeb, "activate"),
				products:    map[string]string{statusChangeWeb: "ACTIVE"},
				changedAPIs: []string{softwareProductList, dataRecipientList},
			})
		}},
		{name: "remove a software product", run: func(t *testing.T) {
			change(t, statusChange{
				action:      productAction(statusChangeRecipient, statusChangeBeta, "remove"),
				products:    map[string]string{statusChangeBeta: "REMOVED"},
				changedAPIs: []string{softwareProductList, dataRecipientList},
			})
		}},
		{name: "suspend makes active software products inactive", run: func(t *testing.T) {
			change(t, statusChange{
				action:     recipientAction(statusChangeRecipient, "suspend"),
				recipients: map[string]string{statusChangeRecipient: "SUSPENDED"},
				products: map[string]string{
					statusChangeWeb:    "INACTIVE",
					statusChangeMobile: "INACTIVE",
					statusChangeBeta:   "REMOVED",
					statusChangeLegacy: "INACTIVE",
				},
				changedAPIs: []string{recipientStatusList, softwareProductList, dataRecipientList},
			})
		}},
		{name: "reinstate does not change software products", run: func(t *testing.T) {
			change(t, statusChange{
				action:      recipientAction(statusChangeRecipient, "reinstate"),
				recipients:  map[string]string{statusChangeRecipient: "ACTIVE"},
				products:    map[string]string{statusChangeWeb: "INACTIVE", statusChangeMobile: "INACTIVE"},
				changedAPIs: []string{recipientStatusList, dataRecipientList},
			})
		}},
		{name: "activate a software product after reinstatement", run: func(t *testing.T) {
			change(t, statusChange{
				action:      productAction(statusChangeRecipient, statusChangeWeb, "activate"),
				products:    map[string]string{statusChangeWeb: "ACTIVE"},
				changedAPIs: []string{softwareProductList, dataRecipientList},
			})
		}},
		{name: "revoke removes software products", run: func(t *testing.T) {
			change(t, statusChange{
				action:     recipientAction(statusChangeRecipient, "revoke"),
				recipients: map[string]string{statusChangeRecipient: "REVOKED"},
				products: map[string]string{
					statusChangeWeb:    "REMOVED",
					statusChangeMobile: "REMOVED",
					statusChangeLegacy: "REMOVED",
				},
				changedAPIs: []string{recipientStatusList, softwareProductList, dataRecipientList},
			})
		}},
		{name: "a revoked data recipient can be reinstated", run: func(t *testing.T) {
			change(t, statusChange{
				action:      recipientAction(statusChangeRecipient, "reinstate"),
				recipients:  map[string]string{statusChangeRecipient: "ACTIVE"},
				products:    map[string]string{statusChangeWeb: "REMOVED"},
				changedAPIs: []string{recipientStatusList, dataRecipientList},
			})
		}},
		{name: "surrender removes software products", run: func(t *testing.T) {
			change(t, statusChange{
				action:      recipientAction(surrenderRecipient, "surrender"),
				recipients:  map[string]string{surrenderRecipient: "SURRENDERED"},
				products:    map[string]string{surrenderWeb: "REMOVED"},
				changedAPIs: []string{recipientStatusList, softwareProductList, dataRecipientList},
			})
		}},
	}
	for _, step := range steps {
		if !t.Run(step.name, step.run) {
			t.Fatalf(
				"stopped at %q: the next steps depend on it. This test passes only one time after the lab loads the "+
					"seed file: run mise run reset, then mise run up",
				step.name,
			)
		}
	}
}
