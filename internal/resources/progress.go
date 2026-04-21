package resources

import "sync"

// Progress describes the current asset-loading stage. Consumers read via
// CurrentProgress() from the main loading loop; the resources package pushes
// updates as files are loaded and deferred initializers run.
type Progress struct {
	Stage   string
	Current int
	Total   int
}

var (
	progressMu sync.Mutex
	progress   Progress
)

func CurrentProgress() Progress {
	progressMu.Lock()
	defer progressMu.Unlock()
	return progress
}

func SetProgress(p Progress) {
	progressMu.Lock()
	progress = p
	progressMu.Unlock()
}
