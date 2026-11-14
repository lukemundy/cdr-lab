package domain

// SupportedScopes gives the scopes that the data holder can put in a registration, a consent or a token.
func SupportedScopes() []string {
	return []string{
		"openid",
		"profile",
		"cdr:registration",
		"common:customer.basic:read",
		"bank:accounts.basic:read",
		"bank:accounts.detail:read",
		"bank:transactions:read",
	}
}
