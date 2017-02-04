// HTTP-triggered randomized sound player, simulating how a dog would bark at
// a door.

// This file implements the FLAC directory scanner and sound player.

// (C)2017 by BJ Black <bj@wjblack.com>, WTFPL licensed--see COPYING

package woofie

import (
	"github.com/gordonklaus/portaudio"
	"github.com/mewkiz/flac"
	"github.com/mewkiz/flac/meta"
	"github.com/wjblack/goflacook"
	"fmt"
	"io/ioutil"
	"math/rand"
	"regexp"
	"time"
)

// Sound represents a single FLAC sample and metadata.  This does not include
// the sample data until played (just the meta).
type Sound struct {
	filepath string
	metadata meta.StreamInfo
}

// NewSample loads the metadata from a filename.
func NewSample(filepath string) (*Sound, error) {
	stream, err := flac.ParseFile(filepath)
	if err != nil { return nil, err }
	stream.Close()
	return &Sound{filepath, *stream.Info}, nil
}

// String dumps info about the sample to a string.
func (s *Sound) String() string {
	total := float32(s.metadata.NSamples) / float32(s.metadata.SampleRate)
	return fmt.Sprintf("%d %d-bit samples/sec, %0.2f secs",
		s.metadata.NSamples, s.metadata.SampleRate, total)
}

// Play fires up pulseaudio and plays a FLAC sample based on goflacook.
func (s *Sound) Play() error {

	// Init pulseaudio
	buffer := make([]int32, 1)
	stream, err := portaudio.OpenDefaultStream(0,
		int(s.metadata.NChannels), float64(s.metadata.SampleRate),
		len(buffer), &(buffer))
	if err != nil {
		logger.Printf("Err: %s, restarting portaudio.\n", err.Error())
		portaudio.Terminate()
		time.Sleep(100*time.Millisecond)
		portaudio.Initialize()
		time.Sleep(100*time.Millisecond)
		stream, err = portaudio.OpenDefaultStream(0,
			int(s.metadata.NChannels),
			float64(s.metadata.SampleRate),
			len(buffer), &(buffer))
	}
	if err != nil { return err }
	outputter := goflacook.NewOutputter(
		func(st *flac.Stream, buf []int32) error {
			buffer = buf
			return stream.Write()
		})

	// Start playing it.  Basically load the FLAC and run decode through
	// Outputter's MainLoop()
	if err != nil { return err }
	err = stream.Start()
	if err != nil { return err }
	err = outputter.Init(s.filepath)
	if err != nil { return err }
	outputter.MainLoop()
	stream.Close()
	return nil
}

// Sounds represents all available FLAC files from the WoofDir.
type Sounds []*Sound

// NewSounds constructs the whole list of sounds from a directory, scanning for
// files that end in .FLAC and whose metadata can be parsed.
func NewSounds(dirpath string) (*Sounds, error) {

	// Seed the RNG for later
	rand.Seed(time.Now().Unix())

	// Open the dir and read all ents in it.  To end up in the slice,
	// the file must end in ".flac" and process through NewSample OK.
	ret := make(Sounds, 0)
	ents, err := ioutil.ReadDir(dirpath)
	if err != nil { return nil, err }
	flacre := regexp.MustCompile("\\.flac$")
	for _, ent := range ents {
		if flacre.MatchString(ent.Name()) {
			filepath := fmt.Sprintf("%s/%s", dirpath, ent.Name())
			sound, err := NewSample(filepath)
			if err != nil { return nil, err }
			ret = append(ret, sound)
		}
	}
	return &ret, nil
}

// PlayRandom plays one random sound from the pile.
func (s *Sounds) PlayRandom() error {
	samp := (*s)[rand.Intn(len(*s))]
	return samp.Play()
}
