//go:build js && wasm

package resources

func init() {
	// iOS WebKit WKWebView has strict per-tab GPU memory limits. A 4096x4096
	// RGBA atlas consumes 64 MB of GPU memory; 2048x2048 uses 16 MB. The
	// default atlas sprite area fits comfortably in 2048 so this is safe.
	maxSpriteAtlasSize = 2048
}
