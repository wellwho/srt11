package commands

import (
	"testing"
	"time"
)

func overlapAudioFile(voice string, offset, duration time.Duration) AudioFile {
	return AudioFile{
		Item: Item{
			Model: Model{model: voice},
		},
		Offset:   offset,
		Duration: duration,
	}
}

func TestFindOverlaps_SameVoiceAcrossADifferentVoiceCue(t *testing.T) {
	// Hana #0 / Matko #1 / Hana #2 -- Hana #0's audio spills 500ms past
	// Hana #2's start, even though Matko #1 sits in between.
	files := []AudioFile{
		overlapAudioFile("hana", 0, 3*time.Second),                     // ends at 3s
		overlapAudioFile("matko", 1*time.Second, 1*time.Second),        // ends at 2s
		overlapAudioFile("hana", 2500*time.Millisecond, 1*time.Second), // starts at 2.5s
	}

	got := findOverlaps(files, 0)
	if len(got) != 1 {
		t.Fatalf("want 1 overlap, got %+v", got)
	}
	if got[0].First != 0 || got[0].Second != 2 {
		t.Fatalf("want overlap between cue 0 and cue 2, got %+v", got[0])
	}
	if got[0].Duration != 500*time.Millisecond {
		t.Fatalf("want 500ms overlap, got %s", got[0].Duration)
	}
}

func TestFindOverlaps_SameVoiceWithinTolerance(t *testing.T) {
	// Same layout as above, but only a 100ms spill against a 120ms tolerance.
	files := []AudioFile{
		overlapAudioFile("hana", 0, 2100*time.Millisecond),      // ends at 2.1s
		overlapAudioFile("matko", 1*time.Second, 1*time.Second), // ends at 2s
		overlapAudioFile("hana", 2*time.Second, 1*time.Second),  // starts at 2s
	}

	got := findOverlaps(files, 120*time.Millisecond)
	if len(got) != 0 {
		t.Fatalf("want no overlaps within tolerance, got %+v", got)
	}
}

func TestFindOverlaps_AdjacentSameVoiceStillFlagged(t *testing.T) {
	files := []AudioFile{
		overlapAudioFile("hana", 0, 1200*time.Millisecond),     // ends at 1.2s
		overlapAudioFile("hana", 1*time.Second, 1*time.Second), // starts at 1s
	}

	got := findOverlaps(files, 0)
	if len(got) != 1 || got[0].First != 0 || got[0].Second != 1 {
		t.Fatalf("want overlap between cue 0 and cue 1, got %+v", got)
	}
	if got[0].Duration != 200*time.Millisecond {
		t.Fatalf("want 200ms overlap, got %s", got[0].Duration)
	}
}

func TestFindOverlaps_CrossVoiceNotFlagged(t *testing.T) {
	files := []AudioFile{
		overlapAudioFile("hana", 0, 1200*time.Millisecond),      // ends at 1.2s
		overlapAudioFile("matko", 1*time.Second, 1*time.Second), // starts at 1s -- different voice
	}

	got := findOverlaps(files, 0)
	if len(got) != 0 {
		t.Fatalf("want no cross-voice overlap flagged, got %+v", got)
	}
}
