package commands

import "time"

// cueOverlap records that two same-voice cues' audio will play at the same
// time: First's audio doesn't finish before Second's starts.
type cueOverlap struct {
	First, Second int
	Duration      time.Duration
}

// findOverlaps flags cues whose audio end spills past the start of the next
// cue spoken by the SAME voice -- not merely the next cue in the list. Voices
// share an output channel, so a different-voice cue sitting in between does
// not prevent the two same-voice clips from playing simultaneously.
//
// One backward pass is enough: nextByVoice tracks, per voice, the closest
// index seen so far that is greater than the current one.
func findOverlaps(files []AudioFile, tolerance time.Duration) []cueOverlap {
	overlaps := make([]cueOverlap, 0)
	nextByVoice := make(map[string]int)
	for i := len(files) - 1; i >= 0; i-- {
		voice := files[i].Item.Model.model
		if next, ok := nextByVoice[voice]; ok {
			spill := (files[i].Offset + files[i].Duration) - files[next].Offset
			if spill > tolerance {
				overlaps = append(overlaps, cueOverlap{First: i, Second: next, Duration: spill})
			}
		}
		nextByVoice[voice] = i
	}
	// walked backwards, so restore ascending order by First
	for l, r := 0, len(overlaps)-1; l < r; l, r = l+1, r-1 {
		overlaps[l], overlaps[r] = overlaps[r], overlaps[l]
	}
	return overlaps
}
