//go:build !js

package runtime

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game"
)

// frameData holds pixel data and timing information
type frameData struct {
	pixels    []byte
	deltaTime float64
}

// audioData holds audio samples
type audioData struct {
	samples [][2]float64
}

// Recorder handles recording canvas frames to video
type Recorder struct {
	name            string
	ffmpegCmd       *exec.Cmd
	ffmpegStdin     *os.File
	canvas          *opengl.Canvas
	recording       bool
	videoFilePath   string
	audioFilePath   string
	finalFilePath   string
	targetFrameRate int // Target FPS for video file
	frameChan       chan frameData
	audioChan       chan audioData
	done            chan struct{}
	audioDone       chan struct{}
	encodeErr       chan error
	audioFile       *os.File
	audioSampleRate int

	// Timing
	framesWritten      int       // Track total frames written
	framesCaptured     int       // Track total frames captured
	startTime          time.Time // Recording start time
	audioStartTime     time.Time // When audio capture actually started
	elapsedTime        float64   // Total elapsed time
	timeAccumulator    float64   // Tracks fractional frames for duplication
	audioSamples       int       // Track total audio samples written
	firstFrameCaptured bool      // Track if we've captured the first video frame
}

// NewRecorder creates a new recorder for the given canvas
func NewRecorder(name string, canvas *opengl.Canvas, frameRate int32) *Recorder {
	return &Recorder{
		name:            name,
		canvas:          canvas,
		targetFrameRate: int(frameRate),
		audioSampleRate: 48000, // Match beep's sample rate
	}
}

// IsRecording returns true if currently recording
func (r *Recorder) IsRecording() bool {
	return r.recording
}

// UpdateCanvas updates the canvas reference (useful when canvas is recreated)
// Must be called from the main thread only
func (r *Recorder) UpdateCanvas(canvas *opengl.Canvas) {
	r.canvas = canvas
}

// Start begins recording to a timestamped file on the desktop
func (r *Recorder) Start() error {
	if r.recording {
		return fmt.Errorf("already recording")
	}

	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create filenames with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	desktopPath := filepath.Join(homeDir, "Desktop")

	// Ensure the directory exists
	if err := os.MkdirAll(desktopPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	baseFilename := fmt.Sprintf("primortal_%s_%s", r.name, timestamp)
	r.videoFilePath = filepath.Join(desktopPath, baseFilename+"_video.mp4")
	r.audioFilePath = filepath.Join(desktopPath, baseFilename+"_audio.wav")
	r.finalFilePath = filepath.Join(desktopPath, baseFilename+".mp4")

	// Get canvas dimensions
	b := r.canvas.Bounds()
	w, h := int(b.W()), int(b.H())

	// Start ffmpeg process for video only (audio will be merged later)
	r.ffmpegCmd = exec.Command("ffmpeg",
		"-f", "rawvideo",
		"-pixel_format", "rgba",
		"-video_size", fmt.Sprintf("%dx%d", w, h),
		"-framerate", fmt.Sprintf("%d", r.targetFrameRate),
		"-i", "pipe:0",
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "18",
		"-pix_fmt", "yuv420p",
		"-y",
		r.videoFilePath,
	)

	// Get stdin pipe
	stdin, err := r.ffmpegCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	r.ffmpegStdin = stdin.(*os.File)

	// Start ffmpeg
	if err := r.ffmpegCmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Create audio WAV file
	audioFile, err := os.Create(r.audioFilePath)
	if err != nil {
		r.ffmpegCmd.Process.Kill()
		return fmt.Errorf("failed to create audio file: %w", err)
	}
	r.audioFile = audioFile

	// Write WAV header (will update later with correct size)
	if err := r.writeWAVHeader(); err != nil {
		audioFile.Close()
		r.ffmpegCmd.Process.Kill()
		return fmt.Errorf("failed to write WAV header: %w", err)
	}

	r.recording = true
	r.frameChan = make(chan frameData, 240) // Buffer up to 240 frames (~4 seconds at 60fps)
	r.audioChan = make(chan audioData, 480) // Buffer up to 480 audio chunks
	r.done = make(chan struct{})
	r.audioDone = make(chan struct{})
	r.encodeErr = make(chan error, 1)
	r.framesWritten = 0
	r.framesCaptured = 0
	r.audioSamples = 0
	r.startTime = time.Now()
	r.elapsedTime = 0
	r.timeAccumulator = 0
	r.firstFrameCaptured = false

	// Start encoding goroutines
	go r.encodeLoop()
	go r.audioEncodeLoop()

	// Note: Audio capture will be started on first frame capture to ensure sync

	log.Info().Msgf("Started recording to: %s (target %d FPS)", r.finalFilePath, r.targetFrameRate)
	return nil
}

// Stop ends the recording and closes the file
func (r *Recorder) Stop() error {
	if !r.recording {
		return fmt.Errorf("not recording")
	}

	r.recording = false

	// Stop audio capture
	r.stopAudioCapture()

	// Signal encoding goroutines to finish and wait for them
	close(r.frameChan)
	close(r.audioChan)
	<-r.done
	<-r.audioDone

	// Check if there was an encoding error
	select {
	case err := <-r.encodeErr:
		if err != nil {
			log.Warn().Err(err).Msg("Error during frame encoding")
		}
	default:
	}

	// Close stdin to signal ffmpeg we're done with video
	if r.ffmpegStdin != nil {
		r.ffmpegStdin.Close()
		r.ffmpegStdin = nil
	}

	// Wait for ffmpeg to finish video encoding
	if r.ffmpegCmd != nil {
		if err := r.ffmpegCmd.Wait(); err != nil {
			log.Warn().Err(err).Msg("ffmpeg exited with error")
		}
		r.ffmpegCmd = nil
	}

	// Update WAV header with correct size
	if r.audioFile != nil {
		if err := r.updateWAVHeader(); err != nil {
			log.Warn().Err(err).Msg("Failed to update WAV header")
		}
		r.audioFile.Close()
		r.audioFile = nil
	}

	// Merge audio and video
	if err := r.mergeAudioVideo(); err != nil {
		log.Error().Err(err).Msg("Failed to merge audio and video")
		game.DebugNotificationf("Recording saved (video only): %s", filepath.Base(r.videoFilePath))
		return err
	}

	actualFPS := float64(r.framesWritten) / r.elapsedTime
	log.Info().Msgf("Stopped recording. Captured: %d frames, Wrote: %d frames, Audio samples: %d, Duration: %.2fs, Actual Video FPS: %.2f",
		r.framesCaptured, r.framesWritten, r.audioSamples, r.elapsedTime, actualFPS)
	log.Info().Msgf("Saved to: %s", r.finalFilePath)
	game.DebugNotificationf("Recording saved: %s", filepath.Base(r.finalFilePath))

	return nil
}

// CaptureFrame captures the current canvas frame and queues it for encoding
// deltaTime is the time elapsed since the last frame in seconds
func (r *Recorder) CaptureFrame(deltaTime float64) error {
	if !r.recording {
		return nil
	}

	// Check for encoding errors
	select {
	case err := <-r.encodeErr:
		return err
	default:
	}

	r.framesCaptured++
	r.elapsedTime += deltaTime
	r.timeAccumulator += deltaTime

	// Start audio capture on first frame to ensure sync
	if !r.firstFrameCaptured {
		r.firstFrameCaptured = true
		r.audioStartTime = time.Now() // Mark when audio actually starts
		delay := time.Since(r.startTime).Seconds()
		log.Info().Msgf("Audio capture starting %.3fs after recording began (startTime=%v, audioStartTime=%v)", 
			delay, r.startTime, r.audioStartTime)
		if err := r.startAudioCapture(); err != nil {
			log.Warn().Err(err).Msg("Failed to start audio capture")
		}
	}

	// Only capture frame when enough time has passed for target frame rate
	targetFrameTime := 1.0 / float64(r.targetFrameRate) // e.g., 1/60 = 0.01666s
	if r.timeAccumulator < targetFrameTime {
		return nil // Skip this frame
	}

	// Reset accumulator (keep fractional remainder for accuracy)
	r.timeAccumulator -= targetFrameTime

	src := r.canvas.Pixels()

	// Make a copy of the pixel data (we need to copy before the next frame)
	frameCopy := make([]byte, len(src))
	copy(frameCopy, src)

	// Try to send to channel, drop frame if buffer is full
	select {
	case r.frameChan <- frameData{pixels: frameCopy, deltaTime: deltaTime}:
		// Frame queued successfully
	default:
		// Buffer full, drop frame
		log.Warn().Msg("Dropped frame - encoding too slow")
	}

	return nil
}

// encodeLoop runs in a goroutine and processes frames from the queue
func (r *Recorder) encodeLoop() {
	defer close(r.done)

	b := r.canvas.Bounds()
	w, h := int(b.W()), int(b.H())

	for frame := range r.frameChan {
		if len(frame.pixels) != w*h*4 {
			r.encodeErr <- fmt.Errorf("unexpected pixel buffer size: got %d, want %d", len(frame.pixels), w*h*4)
			return
		}

		// Flip vertically (OpenGL coordinates are bottom-to-top)
		flipped := r.flipVertical(frame.pixels, w, h)

		// Write raw RGBA frame to ffmpeg stdin
		_, err := r.ffmpegStdin.Write(flipped)
		if err != nil {
			r.encodeErr <- fmt.Errorf("failed to write frame: %w", err)
			return
		}

		r.framesWritten++
	}

	log.Info().Msgf("Encoding complete. Total frames written: %d", r.framesWritten)
}

// flipVertical flips pixel data vertically (OpenGL is bottom-to-top, video is top-to-bottom)
func (r *Recorder) flipVertical(src []byte, w, h int) []byte {
	row := w * 4
	dst := make([]byte, len(src))
	for y := 0; y < h; y++ {
		srcOff := (h - 1 - y) * row
		dstOff := y * row
		copy(dst[dstOff:dstOff+row], src[srcOff:srcOff+row])
	}
	return dst
}

// Toggle starts or stops recording
func (r *Recorder) Toggle() {
	var err error
	if r.recording {
		game.DebugNotificationf("Stopping %s recording...", r.name)
		err = r.Stop()
	} else {
		game.DebugNotificationf("Starting %s recording...", r.name)
		err = r.Start()
	}
	if err != nil {
		game.DebugNotificationf("Recording error (%s): %v", r.name, err)
	}
}

// startAudioCapture sets up the audio tee callback
func (r *Recorder) startAudioCapture() error {
	sys := game.GetAudioSystem()
	if sys == nil || sys.MasterTee == nil {
		return fmt.Errorf("audio system not initialized")
	}

	sys.MasterTee.SetCallback(func(samples [][2]float64) {
		if !r.recording {
			return
		}

		// Send samples to audio encoding goroutine
		select {
		case r.audioChan <- audioData{samples: samples}:
		default:
			// Drop audio if buffer is full
			log.Warn().Msg("Dropped audio samples - encoding too slow")
		}
	})

	return nil
}

// stopAudioCapture removes the audio tee callback
func (r *Recorder) stopAudioCapture() {
	sys := game.GetAudioSystem()
	if sys != nil && sys.MasterTee != nil {
		sys.MasterTee.SetCallback(nil)
	}
}

// audioEncodeLoop processes audio samples and writes them to WAV file
func (r *Recorder) audioEncodeLoop() {
	defer close(r.audioDone)

	for audio := range r.audioChan {
		// Convert float64 samples to int16 PCM
		for _, sample := range audio.samples {
			// Clamp and convert left channel
			left := int16(sample[0] * 32767)
			if err := binary.Write(r.audioFile, binary.LittleEndian, left); err != nil {
				log.Warn().Err(err).Msg("Failed to write left audio sample")
				return
			}

			// Clamp and convert right channel
			right := int16(sample[1] * 32767)
			if err := binary.Write(r.audioFile, binary.LittleEndian, right); err != nil {
				log.Warn().Err(err).Msg("Failed to write right audio sample")
				return
			}

			r.audioSamples++
		}
	}

	log.Info().Msgf("Audio encoding complete. Total samples written: %d", r.audioSamples)
}

// writeWAVHeader writes a WAV file header (will be updated with correct size later)
func (r *Recorder) writeWAVHeader() error {
	// WAV header structure (44 bytes)
	// We'll write placeholder values and update them when we know the final size

	// RIFF header
	r.audioFile.Write([]byte("RIFF"))
	binary.Write(r.audioFile, binary.LittleEndian, uint32(0)) // Placeholder for file size
	r.audioFile.Write([]byte("WAVE"))

	// fmt chunk
	r.audioFile.Write([]byte("fmt "))
	binary.Write(r.audioFile, binary.LittleEndian, uint32(16)) // fmt chunk size
	binary.Write(r.audioFile, binary.LittleEndian, uint16(1))  // PCM format
	binary.Write(r.audioFile, binary.LittleEndian, uint16(2))  // 2 channels (stereo)
	binary.Write(r.audioFile, binary.LittleEndian, uint32(r.audioSampleRate))
	binary.Write(r.audioFile, binary.LittleEndian, uint32(r.audioSampleRate*2*2)) // byte rate
	binary.Write(r.audioFile, binary.LittleEndian, uint16(4))                     // block align (2 channels * 2 bytes)
	binary.Write(r.audioFile, binary.LittleEndian, uint16(16))                    // bits per sample

	// data chunk
	r.audioFile.Write([]byte("data"))
	binary.Write(r.audioFile, binary.LittleEndian, uint32(0)) // Placeholder for data size

	return nil
}

// updateWAVHeader updates the WAV header with the correct file size
func (r *Recorder) updateWAVHeader() error {
	// Calculate sizes
	dataSize := uint32(r.audioSamples * 2 * 2) // samples * channels * bytes_per_sample
	fileSize := dataSize + 36                  // data size + header size (minus 8 for RIFF header)

	// Seek to RIFF size field (offset 4)
	if _, err := r.audioFile.Seek(4, 0); err != nil {
		return err
	}
	if err := binary.Write(r.audioFile, binary.LittleEndian, fileSize); err != nil {
		return err
	}

	// Seek to data size field (offset 40)
	if _, err := r.audioFile.Seek(40, 0); err != nil {
		return err
	}
	if err := binary.Write(r.audioFile, binary.LittleEndian, dataSize); err != nil {
		return err
	}

	return nil
}

// mergeAudioVideo merges the audio and video files using ffmpeg
func (r *Recorder) mergeAudioVideo() error {
	log.Info().Msg("Merging audio and video...")

	videoDuration := float64(r.framesWritten) / float64(r.targetFrameRate)
	audioDuration := float64(r.audioSamples) / float64(r.audioSampleRate)
	durationDiff := videoDuration - audioDuration
	
	// The duration difference represents the true gap, but we need to account for
	// the fact that some of it may be due to audio ending early vs starting late.
	// Use 70% of the duration difference as a heuristic to balance the offset.
	audioDelay := durationDiff * 0.7
	
	log.Info().Msgf("Video duration: %.3fs, Audio duration: %.3fs, Duration diff: %.3fs, Applying offset: %.3fs (70%%)", 
		videoDuration, audioDuration, durationDiff, audioDelay)

	// Use itsoffset BEFORE the audio input to delay it
	cmd := exec.Command("ffmpeg",
		"-i", r.videoFilePath,
		"-itsoffset", fmt.Sprintf("%.3f", audioDelay),
		"-i", r.audioFilePath,
		"-c:v", "copy",
		"-c:a", "aac",
		"-map", "0:v:0",  // Map video from first input
		"-map", "1:a:0",  // Map audio from second input (with offset)
		"-shortest",
		"-y",
		r.finalFilePath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Str("output", string(output)).Msg("ffmpeg merge failed")
		return fmt.Errorf("failed to merge audio and video: %w", err)
	}
	
	log.Info().Msgf("ffmpeg merge output: %s", string(output))

	// Clean up temporary files
	os.Remove(r.videoFilePath)
	os.Remove(r.audioFilePath)

	log.Info().Msgf("Merge complete: %s", r.finalFilePath)
	return nil
}
