// Test routines for the FLAC decoder.

package woofie

import (
	"testing"
)

// TestSounds loads the sample woofdir and ensures that the FLAC files parse OK
func TestSounds(t *testing.T) {
	samples, err := NewSounds("woofs")
	if err != nil { t.Error(err) }
	if len(*samples) != 3 {
		t.Error("Expected 3 samples, got ", len(*samples))
	}
	for _, sample := range *samples {
		if sample.metadata.SampleRate != 22500.0 {
			t.Error("Got wrong samplerate for ", sample.filepath,
				":  Expected 22500, got ",
				sample.metadata.SampleRate)
		}
	}
}
