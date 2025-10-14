#version 330 core

// pixel frame uniforms
in vec4  vColor;
in vec2  vTexCoords;
in float vIntensity;
in vec4  vClipRect;

out vec4 fragColor;

uniform vec4 uColorMask;
uniform vec4 uTexBounds;      // (x,y,w,h) in texels, same as base
uniform sampler2D uTexture;

// Transition uniforms
uniform float uTimeElapsed;        // Time elapsed since transition started (in seconds)
uniform float uTransitionDuration; // Total duration of the transition (in seconds)
uniform float uStableTime;         // Time elapsed in stable state (for ambient effects)
uniform sampler2D uBackgroundTexture; // Background texture (old content)
uniform float uTransitionDirection; // 1.0 = expand from center, -1.0 = collapse to center

// Tunable effect parameters
uniform float uScanlineIntensity;  // Intensity of scan line effect (0.0 - 1.0)
uniform float uVignetteIntensity;  // Intensity of vignette effect (0.0 - 1.0)

vec2 uv() {
    return (vTexCoords - uTexBounds.xy) / uTexBounds.zw;
}

vec4 sampleBackground()  { return texture(uBackgroundTexture,  uv()) * vColor; }
vec4 sampleTexture() { return texture(uTexture, uv()) * vColor; }

// Subtle scan line effect for stable state (can both lighten and darken)
float subtleScanline(vec2 uv, float time) {
    // Variable scroll speed using sin wave for unpredictability
    float scrollSpeed = sin(time * 0.1) * 0.15 + 0.1; // Varies between 0.05 and 0.35 (much slower)
    float scrollOffset = time * scrollSpeed * 0.08;
    
    // Create scan lines that scroll down the screen
    // Use different frequencies for light and dark lines
    float scan1 = sin((uv.y + scrollOffset) * 100.0);
    float scan2 = sin((uv.y + scrollOffset * 0.6) * 50.0);
    float scan3 = sin((uv.y + scrollOffset * 1.3) * 75.0); // Additional layer for variation
    
    // Combine for a complex pattern that can go positive or negative
    // Normalize to ensure good mix of positive and negative values
    float combined = (scan1 * 0.5 + scan2 * 0.3 + scan3 * 0.2);
    
    // Create a slow wave pattern for fade intensity across the screen
    // This creates vertical bands of varying intensity (larger wave = fewer bands)
    float fadeWave = sin(uv.y * 1.5 + time * 0.15) * 0.5 + 0.5; // Larger, slower vertical wave
    
    // Global fade in and out over time
    float globalFade = sin(time * 0.5) * 0.5 + 0.5;
    
    // Combine the wave pattern with global fade for variation
    // Increased influence of fadeWave to make it more noticeable
    float fadeIntensity = mix(globalFade * 0.2, globalFade, fadeWave * 0.8);
    
    return combined * fadeIntensity * uScanlineIntensity;
}

// Subtle phosphor glow effect
float phosphorGlow(vec2 uv, float time) {
    // Simulates the subtle phosphor persistence of CRT monitors
    float glow = sin(time * 3.0) * 0.01 + 0.99;
    return glow;
}

void main() {
    if ((vClipRect != vec4(0,0,0,0)) &&
        (gl_FragCoord.x < vClipRect.x || gl_FragCoord.y < vClipRect.y ||
         gl_FragCoord.x > vClipRect.z || gl_FragCoord.y > vClipRect.w))
        discard;

    // Calculate normalized transition progress (0 to 1)
    float progress = clamp(uTimeElapsed / max(uTransitionDuration, 0.001), 0.0, 1.0);
    bool isTransitioning = progress < 1.0 && uTimeElapsed > 0.0;
    
    // Get base UV coordinates
    vec2 uv = (vTexCoords - uTexBounds.xy) / uTexBounds.zw;
    
    vec4 color;
    
    if (isTransitioning) {
        // === TRANSITION MODE ===
        
        // Create a transition curve - starts fast, ends slower
        float transitionCurve = progress; //1.0 - pow(1.0 - progress, 2.0);
        
        // Sample both textures
        vec4 backgroundTex = sampleBackground();
        vec4 primaryTex = sampleTexture();
        
        // Sweep based on direction
        // uTransitionDirection = 1.0: expand from center (new content appears in middle, expands out)
        // uTransitionDirection = -1.0: collapse to center (new content appears at edges, moves to middle)
        float overshootBuffer = 0.025;
        
        // Calculate distance from center (0.5, 0.5)
        float distanceFromCenter = abs(uv.y - 0.5);
        
        // Row-based dithering - entire rows swap together with some randomness
        vec2 pixelPos = gl_FragCoord.xy;
        float rowIndex = floor(pixelPos.y); // 1 pixel high rows
        float rowRandom = fract(sin(rowIndex * 12.9898) * 43758.5453);
        
        // Broader offset for alternating rows to create wider dithered pattern
        float rowOffset = (rowRandom - 0.5) * 0.18; // Increased from 0.08 to 0.18 for broader effect
        
        // Transition threshold based on distance from center with row offset
        float transitionThreshold = distanceFromCenter + rowOffset;
        
        // Calculate sweep position based on direction
        float sweepRadius;
        float pixelSwap;
        
        if (uTransitionDirection > 0.0) {
            // Expand from center: sweep grows from 0 to 0.5
            sweepRadius = transitionCurve * (0.5 + overshootBuffer);
            // New content starts in middle (small distance from center)
            pixelSwap = step(transitionThreshold, sweepRadius);
        } else {
            // Collapse to center: sweep shrinks from 0.5 to 0
            sweepRadius = (1.0 - transitionCurve) * (0.5 + overshootBuffer);
            // New content starts at edges (large distance from center)
            pixelSwap = step(sweepRadius, transitionThreshold);
        }
        
        // Mix between old and new based on pixel swap
        color = mix(backgroundTex, primaryTex, pixelSwap);
        
        // Add bright sweep lines at the transition boundaries (expanding from center)
        float lineThickness = 0.01;
        // Two lines: one above center, one below center
        float lineDistanceFromRadius = abs(distanceFromCenter - sweepRadius);
        float lineBrightness = smoothstep(lineThickness, 0.0, lineDistanceFromRadius);
        
        // Make the lines bright white/cyan
        vec3 lineColor = vec3(0.8, 1.0, 1.0); // Cyan-white color
        color.rgb = mix(color.rgb, lineColor, lineBrightness * 0.5);

        // Quick flash at the start
        float flashDuration = 0.25;
        if (uTimeElapsed < flashDuration) {
            float flashIntensity = 1.0 - (uTimeElapsed / flashDuration);
            color.rgb = mix(color.rgb, vec3(1.0), flashIntensity * 0.2);
        }
        
    } else {
        // === STABLE STATE (no transition) ===
        color = sampleTexture();
    }

    // === SUSTAINED EFFECTS ===
    // Add subtle scan lines that fade in and out
    float subtleScan = subtleScanline(uv, uStableTime);
    color.rgb *= 1.0 + subtleScan;

    // Add phosphor glow effect
    color.rgb *= phosphorGlow(uv, uStableTime);

    // Very subtle occasional flicker
    float subtleFlicker = 1.0 - sin(uStableTime * 30.0) * 0.002;
    color.rgb *= subtleFlicker;

    
    // Apply tunable vignette effect (always on for that CRT feel)
    vec2 vignetteUV = uv * (1.0 - uv);
    float vignette = vignetteUV.x * vignetteUV.y * 15.0;
    vignette = pow(clamp(vignette, 0.0, 1.0), 0.4);
    // Mix between full brightness and vignette based on intensity
    color.rgb = mix(color.rgb, color.rgb * vignette, uVignetteIntensity);
    
    // Apply color mask and final output
    fragColor = color * uColorMask;
}