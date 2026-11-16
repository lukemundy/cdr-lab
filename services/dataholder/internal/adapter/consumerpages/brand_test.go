package consumerpages_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// eachPage gives each page of the flow, and the refused pages, of a data holder with the brand name.
func eachPage(t *testing.T, bank string) map[string]page {
	t.Helper()

	pages := map[string]page{}

	b := newBrowserOfBank(t, bank)
	pages["customer number"] = b.get()
	pages["form refused"] = b.post(url.Values{"action": {"cancel"}})
	pages["one-time password"] = b.identify(knownCustomer)
	pages["consent"] = b.enter(b.mail.otps[0])

	refusal := newBrowserOfBank(t, bank)
	refusal.consent()

	accountOfADifferentCustomer := "account-9"
	pages["consent account refused"] = refusal.confirm("account-1", accountOfADifferentCustomer)

	rec := httptest.NewRecorder()
	b.pages.Refused(
		rec,
		httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/authorize", nil),
		http.StatusBadRequest,
	)
	pages["refused"] = page{status: rec.Code, header: rec.Header(), body: rec.Body.String()}

	return pages
}

// Verifies: DH-0835, DH-0837, DH-0841
func TestEachPageShowsTheLogoAndTheNameOfTheBankAndTheFooter(t *testing.T) {
	for name, p := range eachPage(t, bankName) {
		t.Run(name, func(t *testing.T) {
			assert.Contains(t, p.body, `<html lang="en-AU" `)
			assert.Contains(t, p.body, `<link rel="stylesheet" href="/static/consumer.css">`)
			assert.Contains(t, p.body, `<link rel="stylesheet" href="/brand/styles.css">`)
			assert.Contains(t, p.body, `src="/brand/logo.svg" alt="Coastal Bank"`)
			assert.Contains(t, p.body, "<footer")
			assert.Contains(t, p.body, "Coastal Bank is not a real bank and this is just a demonstration. "+
				"CDR Lab has no affiliation with any real bank nor the ACCC, DSB, or any real CDR participant.")
			assert.NotContains(t, strings.ToLower(p.body), "this bank")
			assert.NotContains(t, p.body, "<a ", "the pages have no links")
			assert.NotContains(t, p.body, "<script")
			assert.NotContains(t, p.body, "style=", "the CSP permits no inline styles")
		})
	}
}

func TestADifferentBrandShowsOnTheSamePagesWithItsNameEscaped(t *testing.T) {
	pages := eachPage(t, "Harbour <Test> & Bank")

	assert.Contains(t, pages["consent account refused"].body,
		"Harbour &lt;Test&gt; &amp; Bank cannot share an account that you selected.")

	for name, p := range pages {
		t.Run(name, func(t *testing.T) {
			assert.Contains(t, p.body, `alt="Harbour &lt;Test&gt; &amp; Bank"`)
			assert.Contains(t, p.body, "Harbour &lt;Test&gt; &amp; Bank is not a real bank")
			assert.NotContains(t, p.body, "Coastal Bank")
		})
	}
}

// Verifies: DH-2600, DH-2601
func TestEachPageOfTheFlowGivesItsStep(t *testing.T) {
	pages := eachPage(t, bankName)

	assert.Contains(t, pages["customer number"].body, "Step 1 of 3")
	assert.Contains(t, pages["one-time password"].body, "Step 2 of 3")
	assert.Contains(t, pages["consent"].body, "Step 3 of 3")
	assert.NotContains(t, pages["form refused"].body, "Step ")
	assert.NotContains(t, pages["refused"].body, "Step ")
}

func TestEachPageHasTheNameOfTheBankInItsTitle(t *testing.T) {
	pages := eachPage(t, bankName)

	assert.Contains(t, pages["customer number"].body, "<title>Share your data - Coastal Bank</title>")
	assert.Contains(t, pages["refused"].body, "<title>Coastal Bank cannot continue</title>")
}

// Verifies: UI-0029
func TestEachPageUsesTheComfortableDensity(t *testing.T) {
	for name, p := range eachPage(t, bankName) {
		t.Run(name, func(t *testing.T) {
			assert.Contains(t, p.body, `<html lang="en-AU" data-density="comfortable">`)
		})
	}
}

// Verifies: DH-2602, UI-0364, UI-0365
func TestAPageOfAFlowHasAHeroWithTheDataRecipientAndTheMarkOfTheBank(t *testing.T) {
	pages := eachPage(t, bankName)

	for _, name := range []string{"customer number", "one-time password", "consent", "form refused"} {
		t.Run(name, func(t *testing.T) {
			body := pages[name].body
			assert.Contains(t, body, `<span class="cp-initial">B</span>`, "the initial of the data recipient")
			assert.Contains(t, body, `<img class="cp-mark" src="/brand/mark.svg" alt="">`)
			assert.Contains(
				t,
				body,
				"<strong>Budget &lt;App&gt; &amp; Co</strong> asks for your data from Coastal Bank.",
			)
		})
	}

	assert.NotContains(t, pages["refused"].body, "cp-hero", "the refused page has no data recipient")
}

// Verifies: UI-0371
func TestTheStepOfAPageHasAProgressBarThatIsOnlyDecoration(t *testing.T) {
	pages := eachPage(t, bankName)

	assert.Contains(t, pages["customer number"].body, `<div class="cp-progress cp-progress-1" aria-hidden="true">`)
	assert.Contains(t, pages["consent"].body, `<div class="cp-progress cp-progress-3" aria-hidden="true">`)
}

func TestTheInitialOfTheDataRecipientIsItsFirstLetterInUpperCase(t *testing.T) {
	b := newBrowser(t)
	b.registrations.orgName = "ödeme Budget"

	assert.Contains(t, b.get().body, `<span class="cp-initial">Ö</span>`)
}
