#version 330 core

in vec4  vColor;
in vec2  vTexCoords;
in float vIntensity;
in vec4  vClipRect;

out vec4 fragColor;

uniform vec4 uColorMask;
uniform vec4 uTexBounds;      // (x,y,w,h) in texels, same as base
uniform sampler2D uTexture;

// === Bright-pass controls ===
uniform float uThreshold;          // 0..1 (linear brightness) ~ 0.8
uniform float uThresholdIntensity; // 0..1 reduce brightness of pixel breyond threshold
uniform float uKnee;               // 0..1 (soft edge width around threshold). 0 = hard cut ~ 0.05 - 0.2
uniform float uDesaturate;         // 0..1 push bright to white (optional aesthetic) ~ 0.3

// Colors that should always pass bloom
uniform int   uAlwaysCount;       // how many colors are active
uniform vec3  uAlwaysColors[10];  // up to 10 colors to check
uniform float uAlwaysEpsilon;     // tolerance for color matching
uniform float uAlwaysIntensity;   // 0..1 reduce brightness of pixel matching uAlwaysColors

vec4 sampleScene()
{
    vec2 t = (vTexCoords - uTexBounds.xy) / uTexBounds.zw;
    return texture(uTexture, t) * vColor;
}

float luminance(vec3 c)
{
    // Rec. 709
    return dot(c, vec3(0.2126, 0.7152, 0.0722));
}

bool isAlwaysBloom(vec3 color) {
    for (int i = 0; i < uAlwaysCount; ++i) {
        float d = distance(color, uAlwaysColors[i]);
        if (d < uAlwaysEpsilon) {
            return true;
        }
    }
    return false;
}

void main() {
    if ((vClipRect != vec4(0,0,0,0)) &&
        (gl_FragCoord.x < vClipRect.x || gl_FragCoord.y < vClipRect.y ||
         gl_FragCoord.x > vClipRect.z || gl_FragCoord.y > vClipRect.w))
        discard;

    // Base path from Pixel: if no texture contribution, nothing to threshold
    if (vIntensity == 0.0) {
        fragColor = uColorMask * vColor * 0.0; // keep pass black when no texture
        return;
    }

    vec4 src = sampleScene();
    float Y = luminance(src.rgb);

    // Soft threshold via smoothstep
    float cut;
    if (isAlwaysBloom(src.rgb)) {
        cut = uAlwaysIntensity;
    } else {
        cut = (uKnee <= 0.0)
        ? step(uThreshold, Y)
        : smoothstep(uThreshold - uKnee, uThreshold, Y);
        cut = cut * uThresholdIntensity;
    }

    // Optional desaturation so bloom is “light” colored
    vec3 white = vec3(Y);
    vec3 col   = mix(src.rgb, white, clamp(uDesaturate, 0.0, 1.0));

    vec4 outC  = vec4(col * cut, src.a * cut);

    // Respect global color mask (Pixel’s convention)
    fragColor = outC * uColorMask;
}