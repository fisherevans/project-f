package adventure

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/resources"
)

var light = atlas.GetSprite("lights/amber_linear")

var lightShader = `
#version 330 core
uniform sampler2D u_texture; // Pixel injects the canvas' own texture (lights)
uniform sampler2D u_scene;   // we'll bind scene.Texture() here

in vec2  vTexCoords;
out vec4 fragColor;

void main() {
    vec4 L = texture(u_texture, vTexCoords);  // lights canvas
    vec4 S = texture(u_scene,   vTexCoords);  // scene canvas

    // If your lights are colored with alpha as intensity:
    vec3 lit = S.rgb * mix(vec3(1.0), L.rgb, L.a);  // darken by color; no light => 1.0
    fragColor = vec4(lit, S.a);
}
`

type LightSystem struct {
	lights []*Light
}

func NewLightSystem() *LightSystem {
	return &LightSystem{}
}

func (s *LightSystem) Add(light *Light) {
	s.lights = append(s.lights, light)
}

func (s *LightSystem) Render(target pixel.Target, matrix pixel.Matrix) {
	for _, l := range s.lights {
		renderLocation := l.RenderMapLocation().Scaled(resources.MapTileSize.Float())
		l.Render(target, matrix.Moved(renderLocation))
	}
}

type Light struct {
	MapLocation
}

func (l *Light) Render(target pixel.Target, matrix pixel.Matrix) {
	light.Draw(target, pixel.IM.Scaled(pixel.ZV, 2).Chained(matrix))
}

func (l *Light) RenderMapLocation() pixel.Vec {
	return pixel.V(float64(l.X), float64(l.Y))
}
