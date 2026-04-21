#version 330 core

// ---- Pixel-provided ----
in vec4  vColor;
in vec2  vTexCoords;
in float vIntensity;
in vec4  vClipRect;

out vec4 fragColor;

uniform vec4 uColorMask;
uniform vec4 uTexBounds;

uniform sampler2D uTexture;
uniform sampler2D uMask;

vec2 uv() {
    return (vTexCoords - uTexBounds.xy) / uTexBounds.zw;
}

vec4 sampleSource() { return texture(uTexture, uv()) * vColor; }
vec4 sampleMask()  { return texture(uMask,  uv()) * vColor; }

void main() {
    if ((vClipRect != vec4(0,0,0,0)) &&
    (gl_FragCoord.x < vClipRect.x || gl_FragCoord.y < vClipRect.y ||
    gl_FragCoord.x > vClipRect.z || gl_FragCoord.y > vClipRect.w)) {
        discard;
    }

    // If Pixel set vIntensity=0 for culled sprites, keep consistent:
    if (vIntensity == 0.0) {
        fragColor = uColorMask * vColor * 0.0;
        return;
    }

    vec4 src = sampleSource();
    vec4 mask = sampleMask();
    if (src.a == 0.0) {
        discard;
    }
    vec3 color = src.rgb + mask.rgb;
    fragColor = vec4(color, src.a);
}