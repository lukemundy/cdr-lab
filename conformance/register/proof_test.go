//go:build conformance

package register

import "testing"

// A test that fails on purpose, to prove that the cluster job fails when a test fails. It is removed before the merge.
func TestThatFailsOnPurpose(t *testing.T) {
	t.Fatal("this test fails on purpose")
}
