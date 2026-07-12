package suggest

import "testing"

func TestSuggestFindsCloseMatch(t *testing.T) {
	got := Suggest("cpp", "count")
	if got != "cout" {
		t.Errorf("Suggest(cpp, count) = %q, want cout", got)
	}
}

func TestSuggestIgnoresExactMatch(t *testing.T) {
	// "nullptr" has no close neighbors in the cpp dictionary, so an exact
	// match must not produce a spurious "did you mean" for itself.
	got := Suggest("cpp", "nullptr")
	if got != "" {
		t.Errorf("expected no suggestion for exact match, got %q", got)
	}
}

func TestSuggestReturnsEmptyForFarString(t *testing.T) {
	got := Suggest("cpp", "completely_unrelated_long_name")
	if got != "" {
		t.Errorf("expected no suggestion for distant string, got %q", got)
	}
}

func TestSuggestUnknownLanguage(t *testing.T) {
	if got := Suggest("cobol", "anything"); got != "" {
		t.Errorf("expected empty suggestion for unknown language, got %q", got)
	}
}

func TestLevenshteinBasic(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"abc", "abc", 0},
		{"abc", "", 3},
		{"kitten", "sitting", 3},
		{"count", "cout", 1},
	}
	for _, c := range cases {
		if got := levenshtein(c.a, c.b); got != c.want {
			t.Errorf("levenshtein(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}
