package runtime

import (
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

// Recorder handles recording canvas frames to video
type Recorder struct {
	name            string
	ffmpegCmd       *exec.Cmd
	ffmpegStdin     *os.File
	canvas          *opengl.Canvas
	recording       bool
	filePath        string
	targetFrameRate int   // Target FPS for video file
	frameChan       chan frameData
	done            chan struct{}
	encodeErr       chan error
	
	// Timing
	framesWritten   int // Track total frames written
	framesCaptured  int // Track total frames captured
	startTime       time.Time // Recording start time
	elapsedTime     float64   // Total elapsed time
	timeAccumulator float64 // Tracks fractional frames for duplication
}

// NewRecorder creates a new recorder for the given canvas
func NewRecorder(name string, canvas *opengl.Canvas, frameRate int32) *Recorder {
	return &Recorder{
		name:            name,
		canvas:          canvas,
		targetFrameRate: int(frameRate),
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

	// Create filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	desktopPath := filepath.Join(homeDir, "Desktop")
	r.filePath = filepath.Join(desktopPath, fmt.Sprintf("project-f_%s_%s.mp4", r.name, timestamp))

	// Get canvas dimensions
	b := r.canvas.Bounds()
	w, h := int(b.W()), int(b.H())

	// Start ffmpeg process that reads raw RGBA frames from stdin
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
		r.filePath,
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

	r.recording = true
	r.frameChan = make(chan frameData, 240) // Buffer up to 240 frames (~4 seconds at 60fps)
	r.done = make(chan struct{})
	r.encodeErr = make(chan error, 1)
	r.framesWritten = 0
	r.framesCaptured = 0
	r.startTime = time.Now()
	r.elapsedTime = 0
	r.timeAccumulator = 0

	// Start encoding goroutine
	go r.encodeLoop()

	log.Info().Msgf("Started recording to: %s (target %d FPS)", r.filePath, r.targetFrameRate)
	return nil
}

// Stop ends the recording and closes the file
func (r *Recorder) Stop() error {
	if !r.recording {
		return fmt.Errorf("not recording")
	}

	r.recording = false

	// Signal encoding goroutine to finish and wait for it
	close(r.frameChan)
	<-r.done

	// Check if there was an encoding error
	select {
	case err := <-r.encodeErr:
		if err != nil {
			log.Warn().Err(err).Msg("Error during frame encoding")
		}
	default:
	}

	// Close stdin to signal ffmpeg we're done
	if r.ffmpegStdin != nil {
		r.ffmpegStdin.Close()
		r.ffmpegStdin = nil
	}

	// Wait for ffmpeg to finish
	if r.ffmpegCmd != nil {
		if err := r.ffmpegCmd.Wait(); err != nil {
			log.Warn().Err(err).Msg("ffmpeg exited with error")
		}
		r.ffmpegCmd = nil
	}

	actualFPS := float64(r.framesWritten) / r.elapsedTime
	log.Info().Msgf("Stopped recording. Captured: %d frames, Wrote: %d frames, Duration: %.2fs, Actual Video FPS: %.2f", 
		r.framesCaptured, r.framesWritten, r.elapsedTime, actualFPS)
	log.Info().Msgf("Saved to: %s", r.filePath)
	game.DebugNotification("Recording saved: %s", filepath.Base(r.filePath))
	
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
		game.DebugNotification("Stopping %s recording...", r.name)
		err = r.Stop()
	} else {
		game.DebugNotification("Starting %s recording...", r.name)
		err = r.Start()
	}
	if err != nil {
		game.DebugNotification("Recording error (%s): %v", r.name, err)
	}
}
