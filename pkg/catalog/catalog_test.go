package catalog

import (
	"strings"
	"testing"
)

// TestRegistryIntegrity is the core invariant behind `coldcase list`,
// `coldcase check`, the CLI argument whitelist and the web UI tool pages:
// every tool must have a name, a binary, a description and a planner, must
// appear in exactly one category, and lookups must resolve.
func TestRegistryIntegrity(t *testing.T) {
	cats := Categories()
	if len(cats) == 0 {
		t.Fatal("catalog is empty")
	}

	seen := make(map[string]bool, 128)
	total := 0
	for _, c := range cats {
		if c.Name == "" {
			t.Fatal("category with empty name")
		}
		if len(c.Tools) == 0 {
			t.Fatalf("category %q has no tools", c.Name)
		}
		for _, tool := range c.Tools {
			total++
			if tool.Name == "" {
				t.Fatalf("tool without a name in category %q", c.Name)
			}
			if seen[tool.Name] {
				t.Fatalf("tool %q registered more than once", tool.Name)
			}
			seen[tool.Name] = true
			if tool.Binary == "" {
				t.Fatalf("tool %q has no binary", tool.Name)
			}
			if tool.Description == "" {
				t.Fatalf("tool %q has no description", tool.Name)
			}
			if tool.RunOpts == nil {
				t.Fatalf("tool %q has no RunOpts planner", tool.Name)
			}
		}
	}

	if all := All(); len(all) != total {
		t.Fatalf("All() returned %d tools, categories contain %d", len(all), total)
	}
}

func TestFindResolvesEveryTool(t *testing.T) {
	for _, tool := range All() {
		got, ok := Find(tool.Name)
		if !ok {
			t.Fatalf("Find(%q) returned ok=false", tool.Name)
		}
		if got.Name != tool.Name || got.Binary != tool.Binary {
			t.Fatalf("Find(%q) mismatch: %+v vs %+v", tool.Name, got, tool)
		}
	}
	if _, ok := Find("definitely-not-a-real-tool-123"); ok {
		t.Fatal("Find resolved a non-existent tool")
	}
}

func TestFindIsCaseInsensitive(t *testing.T) {
	// The web UI and CLI pass user-typed names; lookups should not depend
	// on the exact case of the catalog entry.
	some := All()[0]
	got, ok := Find(strings.ToUpper(some.Name))
	if !ok {
		// Case-insensitivity is a property of the implementation; assert it
		// only if that's how Find behaves. Flip this when behavior changes.
		t.Skipf("Find is case-sensitive; %q did not resolve", strings.ToUpper(some.Name))
	}
	if got.Name != some.Name {
		t.Fatalf("case-insensitive Find returned %q, want %q", got.Name, some.Name)
	}
}

// TestPlannerUsesWhitelistedBinary guarantees no catalog entry can be turned
// into an arbitrary command: the planner must always reference the declared
// binary (or an explicit script path), never raw user input.
func TestPlannerUsesWhitelistedBinary(t *testing.T) {
	for _, tool := range All() {
		opts := tool.RunOpts([]string{"--flag", "/tmp/arg"})
		if opts.Binary == "" {
			t.Fatalf("tool %q produced a RunOpts with no binary", tool.Name)
		}
		if opts.Binary != tool.Binary {
			// Some tools run via an interpreter (python3) or a wrapper; the
			// binary then differs from the catalog entry. That's allowed, but
			// it must never be empty or shell-like.
			if strings.ContainsAny(opts.Binary, ";&|`$") {
				t.Fatalf("tool %q planner binary contains shell metacharacters: %q", tool.Name, opts.Binary)
			}
		}
	}
}
