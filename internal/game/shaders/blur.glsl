#version 330 core

in vec4  vColor;
in vec2  vTexCoords;
in float vIntensity;
in vec4  vClipRect;

out vec4 fragColor;

uniform vec4 uColorMask;
uniform vec4 uTexBounds;
uniform sampler2D uTexture;
// === Gaussian controls ===
uniform vec2  uDirection;   // texel-space step: (1,0)=horiz, (0,1)=vert
uniform int   uRadius;      // 1..10 typical
uniform float uSigma;       // e.g., radius*0.5..radius

vec4 sampleScene(vec2 uv)
{
    // uTexBounds is in TEXELS; convert vTexCoords to 0..1 window
    vec2 t = (uv - uTexBounds.xy) / uTexBounds.zw;
    return texture(uTexture, t);
}

float gauss(float x, float s)
{
    return exp(-(x*x) / (2.0*s*s));
}

void main() {
    if ((vClipRect != vec4(0,0,0,0)) &&
        (gl_FragCoord.x < vClipRect.x || gl_FragCoord.y < vClipRect.y ||
         gl_FragCoord.x > vClipRect.z || gl_FragCoord.y > vClipRect.w))
        discard;

    // If no texture contribution, nothing to blur
    if (vIntensity == 0.0) {
        fragColor = uColorMask * vColor * 0.0;
        return;
    }

    // Base UV in the same “texel window” space as vTexCoords
    vec2 baseUV = vTexCoords;


    // Accumulate with symmetric taps
    int   r   = clamp(uRadius, 1, 10);
    float sig = max(uSigma, 0.0001);

    vec4  acc = vec4(0.0);
    float wsum = 0.0;

    // center tap
    float w0 = gauss(0.0, sig);
    acc  += sampleScene(baseUV) * w0;
    wsum += w0;

    // side taps
    for (int i = 1; i <= 10; ++i) {
        if (i > r) break;
        float w = gauss(float(i), sig);
        vec2  off = uDirection * float(i);

        acc  += sampleScene(baseUV + off) * w;
        acc  += sampleScene(baseUV - off) * w;
        wsum += 2.0 * w;
    }

    vec4 blurred = acc / wsum;

    // Pixel multiplies texture by vColor upstream; replicate that convention
    vec4 outC = blurred * vColor;

    fragColor = outC * uColorMask;
}