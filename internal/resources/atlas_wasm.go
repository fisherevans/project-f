//go:build js && wasm

package resources

func init() {
	// iOS WebKit WKWebView has strict per-tab GPU memory limits. A 4096x4096
	// RGBA atlas consumes 64 MB of GPU memory; 3072x3072 uses 36 MB. The
	// default atlas sprite area (3.5M px) fits well within 3072 (9.4M px capacity).
	maxSpriteAtlasSize = 3072
}
