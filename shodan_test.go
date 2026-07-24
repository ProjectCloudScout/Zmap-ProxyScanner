package main

import "testing"

func TestBuildProxyURL(t *testing.T) {
	got := buildProxyURL("socks5", "127.0.0.1", 1080)
	want := "socks5://127.0.0.1:1080"
	if got != want {
		t.Fatalf("buildProxyURL() = %q, want %q", got, want)
	}
}

func TestCandidateFromMatchUsesProductAndPort(t *testing.T) {
	match := shodanMatch{Port: 8080, Product: "http proxy", IP: 2130706433}
	candidate := candidateFromMatch(match)
	if candidate.Address != "127.0.0.1" {
		t.Fatalf("candidate address = %q, want %q", candidate.Address, "127.0.0.1")
	}
	if candidate.Protocol != "http" {
		t.Fatalf("candidate protocol = %q, want %q", candidate.Protocol, "http")
	}
	if candidate.Port != 8080 {
		t.Fatalf("candidate port = %d, want %d", candidate.Port, 8080)
	}
}

func TestFormatSummaryIncludesCounts(t *testing.T) {
	got := formatSummary(3, 7, 2, 1, 0, 1, 4)
	want := "live | imported=3 checked=7 success=2 status_err=1 proxy_err=0 timeout=1 threads=4"
	if got != want {
		t.Fatalf("formatSummary() = %q, want %q", got, want)
	}
}

func TestResolveQueryFromArgsUsesPositionalQuery(t *testing.T) {
	got := resolveQueryFromArgs([]string{"port:8080", `org:"Comcast"`}, "")
	want := `port:8080 org:"Comcast"`
	if got != want {
		t.Fatalf("resolveQueryFromArgs() = %q, want %q", got, want)
	}
}

func TestResolveQueryFromArgsStopsAtFlagTokens(t *testing.T) {
	got := resolveQueryFromArgs([]string{"port:8080", `org:"Comcast"`, "-l", "-o", "output.txt"}, "")
	want := `port:8080 org:"Comcast"`
	if got != want {
		t.Fatalf("resolveQueryFromArgs() = %q, want %q", got, want)
	}
}

func TestResolveQueryFromArgsPrefersFlagQuery(t *testing.T) {
	got := resolveQueryFromArgs([]string{"fallback"}, "real query")
	want := "real query"
	if got != want {
		t.Fatalf("resolveQueryFromArgs() = %q, want %q", got, want)
	}
}

func TestBuildShodanQueryKeepsUserProvidedQuery(t *testing.T) {
	got := buildShodanQuery(8080, "", `org:"Comcast"`)
	want := `org:"Comcast"`
	if got != want {
		t.Fatalf("buildShodanQuery() = %q, want %q", got, want)
	}
}
func TestShouldContinueShodanPaginationStopsAtLimit(t *testing.T) {
	if shouldContinueShodanPagination(100000, 200000, 100000) {
		t.Fatal("pagination should stop once the hard limit is reached")
	}
	if !shouldContinueShodanPagination(10, 200, 100000) {
		t.Fatal("pagination should continue while below the hard limit")
	}
}
func TestBuildShodanQueryUsesProxyAndProtocol(t *testing.T) {
	got := buildShodanQuery(80, "socks5", "")
	want := "proxy socks5 port:80"
	if got != want {
		t.Fatalf("buildShodanQuery() = %q, want %q", got, want)
	}
}

func TestBuildShodanQueryPreservesUserProvidedQuery(t *testing.T) {
	got := buildShodanQuery(80, "", `port:8080 org:"Comcast"`)
	want := `port:8080 org:"Comcast"`
	if got != want {
		t.Fatalf("buildShodanQuery() = %q, want %q", got, want)
	}
}
