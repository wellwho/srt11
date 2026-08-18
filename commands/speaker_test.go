package commands

import (
	"regexp"
	"strings"
	"testing"
)

var oldRE = regexp.MustCompile(`\[(.*?)\]\s*(.+)`)

func TestResolveSpeaker(t *testing.T) {
	cases := []struct {
		name         string
		in           string
		wantName     string
		wantDialogue string
		wantTagged   bool
	}{
		{"plain matko line", "[Matko] Stvarno?", "Matko", "Stvarno?", true},
		{"untagged is default speaker", "Ja mislim da je tako.", "", "Ja mislim da je tako.", false},
		{"no space after tag", "[Matko]Odmah!", "Matko", "Odmah!", true},
		{"spaces inside tag", "[ Matko ] Hej", "Matko", "Hej", true},
		{"leading whitespace before tag", "   [Matko] Hej", "Matko", "Hej", true},
		{
			"multiline dialogue keeps every line",
			"[Matko] Pa dobro,\nne razumijem baš.",
			"Matko", "Pa dobro,\nne razumijem baš.", true,
		},
		{
			"bracket used mid sentence is NOT a speaker",
			"Pa, [otprilike] tako nekako.",
			"", "Pa, [otprilike] tako nekako.", false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotName, gotDialogue, gotTagged := resolveSpeaker(tc.in)
			if gotName != tc.wantName || gotDialogue != tc.wantDialogue || gotTagged != tc.wantTagged {
				t.Errorf("resolveSpeaker(%q)\n  got  name=%q dialogue=%q tagged=%v\n  want name=%q dialogue=%q tagged=%v",
					tc.in, gotName, gotDialogue, gotTagged, tc.wantName, tc.wantDialogue, tc.wantTagged)
			}
		})
	}
}

func TestOldRegexBugs(t *testing.T) {
	multiline := "[Matko] Pa dobro,\nne razumijem baš."
	oldMatch := oldRE.FindStringSubmatch(multiline)
	if oldMatch == nil {
		t.Fatal("sanity: old regex should match this input")
	}
	if strings.Contains(oldMatch[2], "ne razumijem") {
		t.Fatal("expected the OLD regex to have dropped the second line, but it did not")
	}
	if _, newDialogue, _ := resolveSpeaker(multiline); !strings.Contains(newDialogue, "ne razumijem") {
		t.Errorf("new helper should preserve the second line, got %q", newDialogue)
	}

	aside := "Pa, [otprilike] tako nekako."
	if oldMatch = oldRE.FindStringSubmatch(aside); oldMatch == nil || oldMatch[1] != "otprilike" {
		t.Fatalf("expected OLD regex to mis-capture a mid-line bracket, got %#v", oldMatch)
	}
	if name, _, tagged := resolveSpeaker(aside); tagged || name != "" {
		t.Errorf("new helper should NOT treat a mid-line bracket as a speaker, got name=%q tagged=%v", name, tagged)
	}
}

func testConfig() *Config {
	return &Config{
		Default: SpeakerConfig{Model: "hana-voice-id", Name: "Hana", Speed: 1.2},
		Models: map[string]SpeakerConfig{
			"Matko":  {Model: "matko-voice-id", Name: "Matko", Speed: 1.0},
			"Broken": {Model: "", Name: "Broken", Speed: 1.0},
		},
	}
}

func TestLookupSpeakerModel(t *testing.T) {
	cfg := testConfig()

	t.Run("known speaker resolves", func(t *testing.T) {
		sc, err := lookupSpeakerModel("Matko", cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sc.Model != "matko-voice-id" {
			t.Errorf("got voice id %q, want %q", sc.Model, "matko-voice-id")
		}
	})

	t.Run("unknown speaker errors loudly", func(t *testing.T) {
		_, err := lookupSpeakerModel("Speaker2", cfg)
		if err == nil {
			t.Fatal("expected an error for an unconfigured speaker, got nil")
		}
		if !strings.Contains(err.Error(), "Speaker2") || !strings.Contains(err.Error(), "known speakers") {
			t.Errorf("error should name the bad speaker and list known ones, got: %v", err)
		}
	})

	t.Run("empty voice id errors loudly", func(t *testing.T) {
		if _, err := lookupSpeakerModel("Broken", cfg); err == nil {
			t.Fatal("expected an error for an empty voice id, got nil")
		} else if !strings.Contains(err.Error(), "empty voice ID") {
			t.Errorf("error should mention the empty voice ID, got: %v", err)
		}
	})
}

func TestKnownSpeakerNames(t *testing.T) {
	if got := knownSpeakerNames(testConfig()); got != "Broken, Matko" {
		t.Errorf("got %q, want %q", got, "Broken, Matko")
	}
}
