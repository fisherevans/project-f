module fisherevans.com/project/f

go 1.24

toolchain go1.24.8

require (
	github.com/go-gl/mathgl v1.1.0
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/gopxl/glhf/v2 v2.0.0
	github.com/gopxl/pixel/v2 v2.3.0
	github.com/lafriks/go-tiled v0.13.0
	github.com/rs/zerolog v1.33.0
	golang.design/x/clipboard v0.7.1
	golang.org/x/image v0.28.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/go-gl/gl v0.0.0-20231021071112-07e5d0ea2e71 // indirect
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20240506104042-037f3cc74f2a // indirect
	github.com/gopxl/mainthread/v2 v2.1.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/exp/shiny v0.0.0-20250606033433-dcc06ee1d476 // indirect
	golang.org/x/mobile v0.0.0-20250606033058-a2a15c67f36f // indirect
	golang.org/x/sys v0.33.0 // indirect
)

replace github.com/gopxl/pixel/v2 => ../pixel
