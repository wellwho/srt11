package commands

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var speakerTagRE = regexp.MustCompile(`^\s*\[([^\]]+)\]\s*([\s\S]+)$`)

func resolveSpeaker(text string) (name, dialogue string, tagged bool) {
	m := speakerTagRE.FindStringSubmatch(text)
	if m == nil {
		return "", text, false
	}
	return strings.TrimSpace(m[1]), strings.TrimSpace(m[2]), true
}

func lookupSpeakerModel(name string, config *Config) (SpeakerConfig, error) {
	sc, ok := config.Models[name]
	if !ok {
		return SpeakerConfig{}, fmt.Errorf(
			"speaker %q is tagged in the subtitles but is not configured under `models` (known speakers: %s)",
			name, knownSpeakerNames(config))
	}
	if sc.Model == "" {
		return SpeakerConfig{}, fmt.Errorf(
			"speaker %q is configured but has an empty voice ID (`model`)", name)
	}
	return sc, nil
}

func knownSpeakerNames(config *Config) string {
	names := make([]string, 0, len(config.Models))
	for k := range config.Models {
		names = append(names, k)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
