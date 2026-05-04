package application

import (
	"strings"
	"testing"

	"github.com/pt9912/m-trace/apps/api/hexagon/domain"
)

func TestRedactURLString_StripsQueryAndFragment(t *testing.T) {
	t.Parallel()
	in := "https://cdn.example.test/playlists/main.m3u8?token=secret&sig=abc#start"
	got := redactURLString(in)
	want := "https://cdn.example.test/playlists/main.m3u8"
	if got != want {
		t.Fatalf("redacted url mismatch: got %q want %q", got, want)
	}
}

func TestRedactURLString_RemovesUserinfo(t *testing.T) {
	t.Parallel()
	got := redactURLString("https://alice:p%40ss@cdn.example.test/seg/0001.ts")
	want := "https://cdn.example.test/seg/0001.ts"
	if got != want {
		t.Fatalf("redacted url mismatch: got %q want %q", got, want)
	}
}

func TestRedactURLString_RedactsTokenLikePathSegments(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			"long base64-ish segment (≥24 chars, ≥80% [A-Za-z0-9_-])",
			"https://cdn.example.test/" + strings.Repeat("a", 32) + "/playlist.m3u8",
			"https://cdn.example.test/:redacted/playlist.m3u8",
		},
		{
			"hex segment with even length ≥ 32",
			"https://cdn.example.test/" + strings.Repeat("ab", 16) + "/seg.ts",
			"https://cdn.example.test/:redacted/seg.ts",
		},
		{
			"jwt-like three base64url blocks separated by dots",
			"https://cdn.example.test/auth/eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ1c2VyIn0.dozjgNryP4J3jVmNHl0w5N",
			"https://cdn.example.test/auth/:redacted",
		},
		{
			// Bypass-Pin (Review-F-2): prozent-encodierte Punkte
			// (`%2E`) dürfen die Token-Heuristik nicht umgehen —
			// `redactPathSegments` dekodiert per `url.PathUnescape`,
			// bevor das Pattern matched.
			"jwt-like with percent-encoded dots",
			"https://cdn.example.test/auth/eyJhbGciOiJIUzI1NiJ9%2EeyJzdWIiOiJ1c2VyIn0%2EsigBlockHere/playlist.m3u8",
			"https://cdn.example.test/auth/:redacted/playlist.m3u8",
		},
		{
			"normal short segment stays untouched",
			"https://cdn.example.test/playlists/v1/main.m3u8",
			"https://cdn.example.test/playlists/v1/main.m3u8",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := redactURLString(tc.in)
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestRedactURLString_UnparsableValueBecomesRedactedSentinel(t *testing.T) {
	t.Parallel()
	got := redactURLString("not-a-url-at-all")
	if got != ":redacted" {
		t.Fatalf("expected :redacted, got %q", got)
	}
}

func TestRedactEventMetaURLs_TouchesAllKnownURLKeys(t *testing.T) {
	t.Parallel()
	meta := domain.EventMeta{
		"url":          "https://cdn.example.test/x.m3u8?t=1",
		"uri":          "https://cdn.example.test/y.ts?s=2",
		"manifest_url": "https://cdn.example.test/m.m3u8?secret=true",
		"segment_url":  "https://cdn.example.test/" + strings.Repeat("z", 40) + "/seg.ts",
		"media_url":    "https://user@cdn.example.test/play.m3u8",
		"network.url":  "https://cdn.example.test/n.m3u8?token=abc",
		"request.url":  "https://cdn.example.test/req.m3u8#frag",
		"response.url": "https://cdn.example.test/res.m3u8?sig=xyz",
	}
	redactEventMetaURLs(meta)
	for k, v := range meta {
		s := v.(string)
		if strings.Contains(s, "?") || strings.Contains(s, "#") || strings.Contains(s, "@") {
			t.Fatalf("meta[%q] still raw: %q", k, s)
		}
	}
	if got := meta["segment_url"].(string); !strings.Contains(got, ":redacted") {
		t.Fatalf("token segment not redacted: %q", got)
	}
}

func TestRedactEventMetaURLs_IsCaseInsensitiveOnKeys(t *testing.T) {
	t.Parallel()
	meta := domain.EventMeta{
		"URL":          "https://cdn.example.test/a.m3u8?t=1",
		"Manifest_URL": "https://cdn.example.test/b.m3u8?t=2",
		"Network.URL":  "https://cdn.example.test/c.m3u8?t=3",
	}
	redactEventMetaURLs(meta)
	for k, v := range meta {
		if strings.Contains(v.(string), "?") {
			t.Fatalf("case-insensitive key %q not redacted: %q", k, v)
		}
	}
}

func TestRedactEventMetaURLs_LeavesPreRedactedKeyUntouched(t *testing.T) {
	t.Parallel()
	const ok = "https://cdn.example.test/playlists/main.m3u8"
	meta := domain.EventMeta{
		"network.redacted_url": ok,
	}
	redactEventMetaURLs(meta)
	if meta["network.redacted_url"].(string) != ok {
		t.Fatalf("network.redacted_url must not be re-redacted, got %q", meta["network.redacted_url"])
	}
}

func TestRedactEventMetaURLs_RedactsUnknownURLishKeys(t *testing.T) {
	t.Parallel()
	meta := domain.EventMeta{
		"foo_uri":     "https://cdn.example.test/x.m3u8?token=abc",
		"some_string": "not-a-url",
		"weird":       "://bare-scheme.example.test/p?x=1",
	}
	redactEventMetaURLs(meta)
	if strings.Contains(meta["foo_uri"].(string), "?") {
		t.Fatalf("unknown URLish key not redacted: %v", meta["foo_uri"])
	}
	if meta["some_string"].(string) != "not-a-url" {
		t.Fatalf("non-url string was mutated: %v", meta["some_string"])
	}
	got := meta["weird"].(string)
	if strings.Contains(got, "?") {
		t.Fatalf("`://`-bearing weird value not redacted: %q", got)
	}
}

func TestRedactEventMetaURLs_NonStringValuesUnchanged(t *testing.T) {
	t.Parallel()
	meta := domain.EventMeta{
		"url":              42,
		"buffered_seconds": 1.8,
		"is_live":          true,
	}
	redactEventMetaURLs(meta)
	if meta["url"] != 42 || meta["buffered_seconds"] != 1.8 || meta["is_live"] != true {
		t.Fatalf("non-string values were mutated: %+v", meta)
	}
}

func TestIsAlreadyRedactedURL(t *testing.T) {
	t.Parallel()
	good := []string{
		"https://cdn.example.test/playlists/main.m3u8",
		"https://cdn.example.test/",
		"http://cdn.example.test/v1/seg.ts",
	}
	for _, g := range good {
		if !isAlreadyRedactedURL(g) {
			t.Errorf("expected redacted-OK: %q", g)
		}
	}
	bad := []string{
		"",
		"not-a-url",
		"https://cdn.example.test/x?token=abc",
		"https://cdn.example.test/x#fragment",
		"https://user:p@cdn.example.test/x",
		"https://cdn.example.test/" + strings.Repeat("a", 32) + "/x",
		// Bypass-Pin (Review-F-2): prozent-encodierter JWT in einem
		// Pfadsegment muss als nicht-redigiert erkannt werden.
		"https://cdn.example.test/eyJhbGciOiJIUzI1NiJ9%2EeyJzdWIiOiJ1c2VyIn0%2Esig",
	}
	for _, b := range bad {
		if isAlreadyRedactedURL(b) {
			t.Errorf("expected redacted-FAIL: %q", b)
		}
	}
}
