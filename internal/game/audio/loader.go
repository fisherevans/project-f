package audio

import (
	"io"
	"os"
	"strings"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/flac"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/vorbis"
	"github.com/gopxl/beep/v2/wav"
)

// openDecode opens and decodes an audio file based on its extension
func openDecode(f io.ReadCloser, extension string) (beep.StreamSeekCloser, beep.Format, error) {

	// Pick decoder by extension
	switch extension {
	case ".wav":
		s, fmt, err := wav.Decode(f)
		if err != nil {
			return nil, beep.Format{}, err
		}
		return s, fmt, nil
	case ".mp3":
		s, fmt, err := mp3.Decode(f)
		if err != nil {
			return nil, beep.Format{}, err
		}
		return s, fmt, nil
	case ".ogg":
		s, fmt, err := vorbis.Decode(f)
		if err != nil {
			f.Close()
			return nil, beep.Format{}, err
		}
		return s, fmt, nil
	case ".flac":
		s, fmt, err := flac.Decode(f)
		if err != nil {
			f.Close()
			return nil, beep.Format{}, err
		}
		return s, fmt, nil
	default:
		f.Close()
		return nil, beep.Format{}, os.ErrInvalid
	}
}

// Preload a short SFX into RAM (WAV/OGG/MP3/FLAC all ok)
func (a *System) Preload(name, extension string, reader io.ReadCloser) error {
	src, fmt, err := openDecode(reader, strings.ToLower(extension))
	if err != nil {
		return err
	}
	defer src.Close()
	a.mu.Lock()
	a.cache[name] = newSourceBuffer(src, fmt)
	a.mu.Unlock()
	return nil
}

func newSourceBuffer(src beep.StreamSeekCloser, fmt beep.Format) *beep.Buffer {
	str := beep.Streamer(src)
	if int(fmt.SampleRate) != targetSR {
		str = beep.Resample(4, fmt.SampleRate, beep.SampleRate(targetSR), str) // quality 4 is fine for SFX
	}
	buf := beep.NewBuffer(beep.Format{SampleRate: targetSR, NumChannels: fmt.NumChannels, Precision: 2})
	buf.Append(str)
	return buf
}
