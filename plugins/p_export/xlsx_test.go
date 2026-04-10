package p_export

import (
	"strings"
	"testing"
)

func TestAssignSheetNamesSanitizesAndDeduplicates(t *testing.T) {
	used := map[string]struct{}{}
	names := assignSheetNames([]string{
		"foo/bar",
		"foo:bar",
		strings.Repeat("a", 40) + "1",
		strings.Repeat("a", 40) + "2",
	}, used)

	seen := map[string]struct{}{}
	for source, sheet := range names {
		if sheet == "" {
			t.Fatalf("sheet name empty for %q", source)
		}
		if len(sheet) > 31 {
			t.Fatalf("sheet name too long for %q: %q", source, sheet)
		}
		if strings.ContainsAny(sheet, `:\/?*[]`) {
			t.Fatalf("sheet name still has invalid chars for %q: %q", source, sheet)
		}
		if _, ok := seen[sheet]; ok {
			t.Fatalf("duplicate sheet name generated: %q", sheet)
		}
		seen[sheet] = struct{}{}
	}
}
