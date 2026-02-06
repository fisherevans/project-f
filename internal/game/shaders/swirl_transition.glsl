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
uniform sampler2D uTexture;   // To texture (new content)

// ==== Swirl Transition uniforms ====
uniform sampler2D uBackgroundTexture; // From texture (old content)
uniform vec2  uCenter;    // swirl center in normalized atlas UV (0..1)
uniform float uRadius;    // 0..1 of the maximum safe radius
uniform float uSwirl;     // max radians at center
uniform float uFalloff;   // 0..1, 0=hard edge, 1=soft edge
uniform float uProgress;  // 0..1 overall transition progress

// Helpers
vec2 atlasUV(vec2 tc, vec4 win){ return (tc - win.xy) / win.zw; }

void main() {
    // Pixel clip test
    if ((vClipRect != vec4(0,0,0,0)) &&
        (gl_FragCoord.x < vClipRect.x || gl_FragCoord.y < vClipRect.y ||
         gl_FragCoord.x > vClipRect.z || gl_FragCoord.y > vClipRect.w)) {
        discard;
    }

    // Base scene sampling path
    vec2 uv = atlasUV(vTexCoords, uTexBounds);

    // --- Swirl logic ---
    float r = uRadius;

    vec2  d   = uv - uCenter;
    float len = length(d);

    // Default: no distortion
    vec2 warped = uv;

    // Smooth the progress to avoid jarring start/stop
    float smoothP = smoothstep(0.0, 1.0, uProgress);
    // Swirl peaks at 0.5 progress
    float swirlAmount = uSwirl * 4.0 * smoothP * (1.0 - smoothP);

    if (r > 0.0 && len < r && swirlAmount != 0.0) {
        float t = 1.0 - (len / r);
        float f = (uFalloff <= 0.0) ? t : smoothstep(0.0, 1.0, pow(t, 1.0 + 2.0*uFalloff));
        
        // To avoid tearing at edges, we fade the effect based on distance to nearest edge
        float distToEdge = min(min(uv.x, 1.0 - uv.x), min(uv.y, 1.0 - uv.y));
        float edgeFade = smoothstep(0.0, 0.04, distToEdge); 

        float a = swirlAmount * f * edgeFade;

        float s = sin(a), c = cos(a);
        vec2  rot = vec2(c*d.x - s*d.y, s*d.x + c*d.y);

        warped = uCenter + rot;
    }

    vec2 safeUV = clamp(warped, 0.0, 1.0);

    // Sample both textures
    vec4 fromTex = texture(uBackgroundTexture, safeUV);
    vec4 toTex = texture(uTexture, safeUV);

    // Mix based on progress
    vec4 tex = mix(fromTex, toTex, uProgress);

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
