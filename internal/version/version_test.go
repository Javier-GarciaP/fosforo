package version

import "testing"

func TestReleaseVersion(t *testing.T) {
	for in, want := range map[string]string{
		"v0.8.0":                               "0.8.0",
		"v0.8.0-rc.1":                          "0.8.0-rc.1",
		"":                                     "",
		"(devel)":                              "",
		"v0.8.0+dirty":                         "",
		"v0.0.0-20261021140633-df8d393b89f2":   "",
		"v0.8.1-0.20261021140633-df8d393b89f2": "",
		"v0.8.1-0.20261021140633-df8d393b89f2+dirty": "",
	} {
		if got := releaseVersion(in); got != want {
			t.Errorf("releaseVersion(%q) = %q, quería %q", in, got, want)
		}
	}
}
