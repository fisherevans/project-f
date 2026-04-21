//go:build js && wasm

package runtime

import "github.com/gopxl/pixel/v2/backends/opengl"

type Recorder struct{}

func NewRecorder(name string, canvas *opengl.Canvas, frameRate int32) *Recorder {
	return &Recorder{}
}

func (r *Recorder) IsRecording() bool                      { return false }
func (r *Recorder) UpdateCanvas(canvas *opengl.Canvas)     {}
func (r *Recorder) Start() error                           { return nil }
func (r *Recorder) Stop() error                            { return nil }
func (r *Recorder) CaptureFrame(deltaTime float64) error   { return nil }
func (r *Recorder) Toggle()                                {}
