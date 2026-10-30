package buildinfo

import "testing"

func TestIsDevelopment(t *testing.T) {
	old := Version
	defer func() { Version = old }()
	for v, want := range map[string]bool{"dev": true, "": true, "5.0.0-SNAPSHOT": true, "5.0.0": false, "4.0.1": false} {
		Version = v
		if got := IsDevelopment(); got != want {
			t.Errorf("%q: %v", v, got)
		}
	}
}
