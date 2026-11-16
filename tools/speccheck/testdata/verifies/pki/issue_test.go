package pki

// Verifies: PKI-0001
func TestIssue(t *testing.T) {
	tests := []struct{ name string }{
		// Verifies: PKI-0009
		{name: "missing"},
		// Verifies: PKI-0002
		{name: "judgement"},
		// Verifies: PKI-0003
		{name: "withdrawn"},
	}
}
