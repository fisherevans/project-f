#version 330 core

in vec4  vColor;
in vec2  vTexCoords;
in float vIntensity;
in vec4  vClipRect;

out vec4 fragColor;

uniform vec4 uColorMask;
uniform vec4 uTexBounds;      // (x,y,w,h) in texels, same as base
uniform sampler2D uTexture;

// Maximum number of palette levels supported by this shader
const int MAX_LEVELS = 16;

// === XBit palette controls ===
uniform int   uThresholdCount;                 // number of palette colors (1..MAX_LEVELS), darkest -> lightest
uniform vec3  uPalette[MAX_LEVELS];            // palette colors ordered darkest -> lightest
uniform float uThresholds[MAX_LEVELS - 1];     // brightness thresholds [0,1], ascending; length = uThresholdCount - 1
uniform float uPaletteEpsilon;                 // tolerance for palette color match (to preserve already-quantized pixels)

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

bool isPaletteColor(vec3 color) {
    int n = uThresholdCount;
    for (int i = 0; i < MAX_LEVELS; ++i) {
        if (i >= n) break;
        float d = distance(color, uPalette[i]);
        if (d <= uPaletteEpsilon) {
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

    vec4 src = sampleScene();
    float Y = clamp(luminance(src.rgb), 0.0, 1.0);

    // If the source pixel already equals a palette color (within epsilon), keep it as-is
    vec3 outRGB;
    if (isPaletteColor(src.rgb)) {
        outRGB = src.rgb;
    } else if (uThresholdCount <= 0) {
        // No palette provided: pass-through
        outRGB = src.rgb;
    } else {
        // Pick palette index by counting thresholds <= brightness
        int idx = 0;
        int limit = (uThresholdCount > 0) ? (uThresholdCount - 1) : 0;
        for (int i = 0; i < MAX_LEVELS - 1; ++i) {
            if (i >= limit) break;
            if (Y >= clamp(uThresholds[i], 0.0, 1.0)) {
                idx++;
            }
        }
        idx = clamp(idx, 0, max(uThresholdCount - 1, 0));
        outRGB = uPalette[idx];
    }

    // Apply alpha from source and color mask
    vec4 outC = vec4(outRGB, 1.0) * src.a;  // Apply source alpha to the final color
    fragColor = outC * uColorMask;
}