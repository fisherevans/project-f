#version 330 core

// ==== Pixel base varyings ====
in vec4  vColor;
in vec2  vTexCoords;
in float vIntensity;
in vec4  vClipRect;

out vec4 fragColor;

// ==== Pixel base uniforms ====
uniform vec4 uColorMask;
uniform vec4 uTexBounds;
uniform sampler2D uTexture;

// ==== Transition controls ====
uniform float uTime;          // seconds
uniform float uProgress;      // 0..1
uniform int   uReverse;       // 0 = normal, 1 = reverse (use 1-p)
uniform float uCurve;         // >=1, easing exponent (2–4 strong finish)

// Max strengths reached at ramp=1
uniform float uAmpMax;        // ~0.02..0.08 (UV displacement)
uniform float uSwirlMax;      // ~1.5..4.0  (radians at edge)
uniform float uCABMax;        // ~0..0.003  (chromatic aberration)

// Wave params
uniform vec2  uFreq1;         // e.g. (8,11)
uniform vec2  uFreq2;         // e.g. (15,9)
uniform float uSpeed;         // e.g. 3.0

// --- helpers ---
float easePowOut(float t, float k){ t = clamp(t,0.0,1.0); return 1.0 - pow(1.0 - t, k); }
vec2 atlasUV(vec2 tc, vec4 win){ return (tc - win.xy) / win.zw; }

vec2 squiggle(vec2 uv, float t, float amp, vec2 f1, vec2 f2, float speed){
    float p1 = sin(uv.y * f1.x + t*speed) + sin(uv.x * f1.y - t*speed*0.73);
    float p2 = sin(uv.x * f2.x + t*speed*1.21) + sin(uv.y * f2.y - t*speed*0.57);
    return uv + amp * 0.5 * vec2(p1, p2);
}

vec2 swirl(vec2 uv, float amount){
    vec2 c = uv - 0.5;
    float r = length(c);
    float a = amount * r; // grows toward edges
    float s = sin(a), co = cos(a);
    vec2 rot = vec2(co*c.x - s*c.y, s*c.x + co*c.y);
    return rot + 0.5;
}

void main(){
    // Pixel clip
    if ((vClipRect != vec4(0,0,0,0)) &&
        (gl_FragCoord.x < vClipRect.x || gl_FragCoord.y < vClipRect.y ||
         gl_FragCoord.x > vClipRect.z || gl_FragCoord.y > vClipRect.w)) {
        discard;
    }

    // Base scene (Pixel sampling path)
    vec4 scene;
    vec2 baseUV = atlasUV(vTexCoords, uTexBounds);
    if (vIntensity == 0.0) {
        scene = uColorMask * vColor;
    } else {
        vec4 tex = texture(uTexture, baseUV);
        scene  = vec4(0.0);
        scene += (1.0 - vIntensity) * vColor;
        scene += vIntensity * vColor * tex;
        scene *= uColorMask;
    }

    // Ramp (optionally reversed)
    float p = clamp(uProgress, 0.0, 1.0);
    if (uReverse != 0) p = 1.0 - p;
    float w = easePowOut(p, max(uCurve, 1.0)); // strong finish as p→1

    // Build distorted UVs
    float amp   = uAmpMax   * w;
    float swirlAmt = uSwirlMax * w;
    float cab   = uCABMax   * w;

    vec2 uv = baseUV;
    uv = squiggle(uv, uTime, amp, uFreq1, uFreq2, uSpeed);
    uv = swirl(uv, swirlAmt);

    // Chromatic aberration (subtle RGB offset)
    vec4 texR = texture(uTexture, uv + vec2( cab, 0.0));
    vec4 texG = texture(uTexture, uv);
    vec4 texB = texture(uTexture, uv - vec2( cab, 0.0));
    vec4 texRGB = vec4(texR.r, texG.g, texB.b, texG.a);

    // Re-apply Pixel’s vIntensity/vColor mixing on the warped sample
    vec4 outCol;
    if (vIntensity == 0.0) {
        outCol = uColorMask * vColor; // untextured vertices
    } else {
        outCol  = vec4(0.0);
        outCol += (1.0 - vIntensity) * vColor;
        outCol += vIntensity * vColor * texRGB;
        outCol *= uColorMask;
    }

    fragColor = outCol;
}