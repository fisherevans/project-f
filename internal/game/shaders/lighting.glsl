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
uniform sampler2D uTexture;   // SCENE (the source being drawn)

// ==== Lighting uniforms ====
uniform sampler2D uLightMap;    // colored light map (same mapping as scene)
uniform float     uLightThreshold; // 0..1, where darkening switches to brightening (default 0.5)
uniform float     uDarkStrength;   // 0..1, how strongly to apply darkening
uniform float     uBrightStrength; // 0..1, how strongly to apply brightening

void main() {
    // Pixel's clip test
    if ((vClipRect != vec4(0,0,0,0)) &&
        (gl_FragCoord.x < vClipRect.x || gl_FragCoord.y < vClipRect.y ||
         gl_FragCoord.x > vClipRect.z || gl_FragCoord.y > vClipRect.w)) {
        discard;
    }

    // --- Sample SCENE exactly like Pixel base ---
    vec4 scene;
    if (vIntensity == 0.0) {
        scene = uColorMask * vColor;
    } else {
        vec2 ts = (vTexCoords - uTexBounds.xy) / uTexBounds.zw;
        scene  = vec4(0.0);
        scene += (1.0 - vIntensity) * vColor;
        scene += vIntensity * vColor * texture(uTexture, ts);
        scene *= uColorMask;
    }

    // Early out if lighting disabled
    if (uDarkStrength <= 0.0 && uBrightStrength <= 0.0) {
        fragColor = scene;
        return;
    }

    // Recompute the same normalized texcoords for sampling the light map
    vec2 ts = (vTexCoords - uTexBounds.xy) / uTexBounds.zw;
    vec3 L = texture(uLightMap, ts).rgb;   // 0..1 per channel, colored & aligned with scene
    vec3 S = scene.rgb;

    // Threshold logic (per channel)
    float T = clamp(uLightThreshold, 0.0001, 0.9999);

    // Dark side (L < T): map [0..T] -> [0..1] factor = L/T, then multiply scene
    vec3 darkFactor = clamp(L / T, 0.0, 1.0);
    vec3 darkened   = S * darkFactor;

    // Bright side (L > T): p = (L - T)/(1 - T) in [0..1], add p * (1 - S) (relative headroom)
    vec3 p          = clamp((L - T) / (1.0 - T), 0.0, 1.0);
    vec3 brightened = S + p * (1.0 - S);

    // Per-channel select: 1 where L <= T (dark), 0 where L > T (bright)
    vec3 useDark = step(L, vec3(T));

    // Apply separate strengths
    float kd = clamp(uDarkStrength,   0.0, 1.0);
    float kb = clamp(uBrightStrength, 0.0, 1.0);

    vec3 litDark   = mix(S, darkened,   kd);
    vec3 litBright = mix(S, brightened, kb);

    // Combine per-channel
    vec3 lit = mix(litBright, litDark, useDark);

    fragColor = vec4(lit, scene.a);
}