package http

import "testing"

// TestParseTraceParent_Valid verifiziert den W3C-konformen Pfad:
// gültiger Header → ok=true, korrekte trace_id und parent_id.
func TestParseTraceParent_Valid(t *testing.T) {
	t.Parallel()
	raw := "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"
	tid, pid, ok := parseTraceParent(raw)
	if !ok {
		t.Fatalf("expected ok, got false")
	}
	if tid != "0af7651916cd43dd8448eb211c80319c" {
		t.Errorf("trace_id = %q, want 0af7651916cd43dd8448eb211c80319c", tid)
	}
	if pid != "b7ad6b7169203331" {
		t.Errorf("parent_id = %q, want b7ad6b7169203331", pid)
	}
}

// TestParseTraceParent_Invalid deckt jeden formalen Verstoß ab,
// der laut W3C-Spec zu Parse-Error führt — Caller mappt das auf
// Span-Attribut mtrace.trace.parse_error=true und Server-Fallback.
func TestParseTraceParent_Invalid(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"empty":                "",
		"too short":            "00-abc",
		"too long":             "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01-extra",
		"wrong version":        "01-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
		"two parts":            "00-0af7651916cd43dd8448eb211c80319c",
		"non-hex trace_id":     "00-0af7651916cd43dd8448eb211c8031XX-b7ad6b7169203331-01",
		"non-hex parent_id":    "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169Z03331-01",
		"non-hex flags":        "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-XX",
		"upper-case trace_id":  "00-0AF7651916CD43DD8448EB211C80319C-b7ad6b7169203331-01",
		"all-zero trace_id":    "00-00000000000000000000000000000000-b7ad6b7169203331-01",
		"all-zero parent_id":   "00-0af7651916cd43dd8448eb211c80319c-0000000000000000-01",
		"trace_id wrong len":   "00-0af7651916cd43dd8448eb211c8031-b7ad6b7169203331111-01",
		"parent_id wrong len":  "00-0af7651916cd43dd8448eb211c80319c-b7ad6b71692033-0101",
	}
	for name, raw := range cases {
		if _, _, ok := parseTraceParent(raw); ok {
			t.Errorf("%s: expected parse failure, got ok=true", name)
		}
	}
}
