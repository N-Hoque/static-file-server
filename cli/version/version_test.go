package version

import "testing"

func TestVersion(t *testing.T) {
	if err := Run(); err != nil {
		t.Errorf("While running version got %v", err)
	}
}
