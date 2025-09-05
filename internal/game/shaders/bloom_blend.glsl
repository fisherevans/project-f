#version 330 core

// from pixel

in vec4  vColor;
in vec2  vTexCoords;
in float vIntensity;
in vec4  vClipRect;

out vec4 fragColor;

uniform vec4 uColorMask;
uniform vec4 uTexBounds;
uniform sampler2D uTexture;

// custom

// --- Inputs ---
uniform sampler2D uBase;

// --- Controls ---
uniform float uIntensity; // overall bloom strength
uniform float uBloomBias; // e.g., 1.0..2.0  (how strongly to favor darker bloom)
uniform float uSceneBias; // e.g., 0.0..2.0  (how much to reduce effect on bright scene)


vec2 uv() {
    return (vTexCoords - uTexBounds.xy) / uTexBounds.zw;
}

vec4 sampleBase()  { return texture(uBase,  uv()) * vColor; }
vec4 sampleTexture() { return texture(uTexture, uv()) * vColor; }

float luma(vec3 c){ return dot(c, vec3(0.2126, 0.7152, 0.0722)); }
float saturate(float x){ return clamp(x, 0.0, 1.0); }
vec3  safeNorm(vec3 c){ return (max(max(c.r,c.g),c.b) > 1e-6) ? normalize(c) : vec3(0.0); }

void main() {
    // clip like Pixel
    if ((vClipRect != vec4(0,0,0,0)) &&
    (gl_FragCoord.x < vClipRect.x || gl_FragCoord.y < vClipRect.y ||
    gl_FragCoord.x > vClipRect.z || gl_FragCoord.y > vClipRect.w)) {
        discard;
    }
    if (vIntensity == 0.0) {
        fragColor = uColorMask * vColor * 0.0;
        return;
    }


    vec4 bloom = sampleTexture();
    float lumaBloom = luma(clamp(bloom.rgb,  0.0, 1.0));

    vec4 scene  = sampleBase();
    float lumaScene = luma(clamp(scene.rgb, 0.0, 1.0));

    // Adaptive gain - for both bloom and seen:
    // - brighten darks
    // - prevent brights from washing out
    float gainBloom   = pow(1.0 - lumaBloom, max(uBloomBias, 0.0));
    float gainScene   = pow(1.0 - lumaScene, max(uSceneBias, 0.0));
    float k           = max(uIntensity, 0.0) * gainBloom * gainScene;

    // Hue-preserving glow vector: use bloom's hue direction * its luminance
    vec3 hue  = safeNorm(bloom.rgb);
    vec3 glow = hue * lumaBloom;

    // Additive composite with adaptive, colored glow
    vec3 outRGB = scene.rgb + glow * k;

    // Preserve scene alpha
    fragColor = vec4(outRGB, scene.a) * uColorMask;
}