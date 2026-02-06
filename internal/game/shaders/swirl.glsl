#version 330 core

// ==== Pixel base varyings ====
in vec4  vColor;
in vec2  vTexCoords;
in float vIntensity;
in vec4  vClipRect;

out vec4 fragColor;

// ==== Pixel base uniforms ====
uniform vec4 uColorMask;
uniform vec4 uTexBounds;      // atlas window for uTexture
uniform sampler2D uTexture;   // scene/source being drawn

// ==== Swirl controls ====
uniform vec2  uCenter;    // swirl center in normalized atlas UV (0..1), e.g. vec2(0.5,0.5)
uniform float uRadius;    // 0..1 of the maximum safe radius (see below)
uniform float uSwirl;     // radians at center (decreases toward edge)
uniform float uFalloff;   // 0..1, 0=hard edge, 1=soft edge
uniform float uProgress;  // 0..1 overall strength (optional; set 1.0 if unused)

// Helpers
vec2 atlasUV(vec2 tc, vec4 win){ return (tc - win.xy) / win.zw; }

void main() {
    // Pixel clip test
    if ((vClipRect != vec4(0,0,0,0)) &&
        (gl_FragCoord.x < vClipRect.x || gl_FragCoord.y < vClipRect.y ||
         gl_FragCoord.x > vClipRect.z || gl_FragCoord.y > vClipRect.w)) {
        discard;
    }

    // Base scene sampling path (Pixel-style)
    vec2 uv = atlasUV(vTexCoords, uTexBounds);

    // --- Swirl with guaranteed edge preservation ---
    float r = uRadius;

    // Vector from center, normalized by r
    vec2  d   = uv - uCenter;
    float len = length(d);

    // Default: no distortion
    vec2 warped = uv;

    if (r > 0.0 && len < r && uProgress > 0.0 && uSwirl != 0.0) {
        // Distance 0..1 within the swirl region
        float t = 1.0 - (len / r);                 // 1 at center → 0 at edge
        // Soft falloff near edge
        float f = (uFalloff <= 0.0) ? t : smoothstep(0.0, 1.0, pow(t, 1.0 + 2.0*uFalloff));
        
        // Smooth the progress to avoid jarring start/stop
        float smoothP = smoothstep(0.0, 1.0, uProgress);
        // Scale by overall progress
        f *= smoothP;

        // To avoid tearing at edges, we fade the effect based on distance to nearest edge
        float distToEdge = min(min(uv.x, 1.0 - uv.x), min(uv.y, 1.0 - uv.y));
        float edgeFade = smoothstep(0.0, 0.04, distToEdge); 

        // Rotation angle decreases toward edge (max at center)
        float a = uSwirl * f * edgeFade;

        // Rotate d by angle a
        float s = sin(a), c = cos(a);
        vec2  rot = vec2(c*d.x - s*d.y, s*d.x + c*d.y);

        warped = uCenter + rot;
    }

    // Sample scene using warped UV (clamped to avoid precision nicks)
    vec2 safeUV = clamp(warped, 0.0, 1.0);

    vec4 tex = texture(uTexture, safeUV);

    // Re-apply Pixel’s vertex color/intensity modulation
    vec4 outCol;
    if (vIntensity == 0.0) {
        outCol = uColorMask * vColor;
    } else {
        outCol  = vec4(0.0);
        outCol += (1.0 - vIntensity) * vColor;
        outCol += vIntensity * vColor * tex;
        outCol *= uColorMask;
    }

    fragColor = outCol;
}