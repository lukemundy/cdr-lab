//go:build openid

package openid

// consumerPages gives the browser tasks for the pages of the data holder, which all have the path of the authorisation
// endpoint, so each task and each command for an element that can be absent is optional.
type consumerPages struct {
	authorise      string
	customerNumber string
}

const (
	continueButton = "//button[@value='continue']"
	// errorPageTitle matches the heading of the page that the data holder shows when it does not redirect to the
	// client. The heading starts with the brand name of the data holder, and the suite finds the pattern in the text.
	errorPageTitle = "cannot continue"
)

// callback matches only the redirect URI, not an authorisation URL that has the redirect URI as a parameter.
const callback = redirectURI + "*"

func task(name, match string, commands ...[]any) map[string]any {
	list := make([]any, len(commands))
	for i, command := range commands {
		list[i] = command
	}

	return map[string]any{"task": name, "match": match, "optional": true, "commands": list}
}

func verifyComplete() map[string]any {
	return task("Verify Complete", callback, []any{"wait", "id", "submission_complete", 10})
}

// consent does the flow with the fixed OTP; on the error page of the data holder, only the error page task does
// something.
func (p consumerPages) consent() map[string]any {
	return map[string]any{
		"match": p.authorise,
		"tasks": []any{
			task("Customer number", p.authorise,
				[]any{"text", "id", "identifier", p.customerNumber, "optional"},
				[]any{"click", "xpath", continueButton, "optional"},
			),
			task("One-time password", p.authorise,
				[]any{"text", "id", "otp", fixedOTP, "optional"},
				[]any{"click", "xpath", continueButton, "optional"},
			),
			task("Consent", p.authorise,
				[]any{"click", "name", "account", "optional"},
				[]any{"click", "xpath", "//button[@value='confirm']", "optional"},
			),
			task("Error page", p.authorise,
				[]any{"wait", "xpath", "//h1", 10, errorPageTitle, "update-image-placeholder-optional"},
			),
			verifyComplete(),
		},
	}
}

func (p consumerPages) cancel() map[string]any {
	return map[string]any{
		"match": p.authorise,
		"tasks": []any{
			task("Cancel", p.authorise, []any{"click", "xpath", "//button[@value='cancel']"}),
			verifyComplete(),
		},
	}
}

func (p consumerPages) visitOnly() map[string]any {
	return map[string]any{
		"match":       p.authorise,
		"match-limit": 1,
		"tasks":       []any{task("Visit only", p.authorise, []any{"wait", "id", "identifier", 10})},
	}
}
