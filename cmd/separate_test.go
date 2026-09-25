package cmd

import "testing"

func TestStemsToOutputType(t *testing.T) {
	cases := map[string]string{
		"vocals":                               "VOCALS",
		"instrumental":                         "INSTRUMENTAL",
		"both":                                 "BOTH",
		"vocals,instrumental":                  "BOTH",
		"vocals,drums,bass,other":              "FOUR_STEMS",
		" vocals , drums , bass , other ":      "FOUR_STEMS",
		"six-stems":                            "SIX_STEMS",
		"vocals,drums,bass,other,piano,guitar": "SIX_STEMS",
	}
	for in, want := range cases {
		if got := stemsToOutputType(in); got != want {
			t.Errorf("stemsToOutputType(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestModelToQuality(t *testing.T) {
	cases := map[string]string{
		"htdemucs_ft": "BEST",
		"HTDEMUCS":    "BALANCED",
		"fast":        "FAST",
		"unknown":     "BEST",
	}
	for in, want := range cases {
		if got := modelToQuality(in); got != want {
			t.Errorf("modelToQuality(%q) = %s, want %s", in, got, want)
		}
	}
}
