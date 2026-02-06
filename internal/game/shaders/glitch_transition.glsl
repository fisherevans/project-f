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

// ==== Glitch Transition uniforms ====
uniform sampler2D uBackgroundTexture; // From texture (old content)
uniform float uProgress;              // 0..1 overall transition progress
uniform float uTime;                  // Time for animation

// Helpers
vec2 atlasUV(vec2 tc, vec4 win){ return (tc - win.xy) / win.zw; }

float rand(vec2 co) {
    return fract(sin(dot(co.xy ,vec2(12.9898,78.233))) * 43758.5453);
}

float rand(float n) {
    return fract(sin(n) * 43758.5453123);
}

void main() {
    // Pixel clip test
    if ((vClipRect != vec4(0,0,0,0)) &&
        (gl_FragCoord.x < vClipRect.x || gl_FragCoord.y < vClipRect.y ||
         gl_FragCoord.x > vClipRect.z || gl_FragCoord.y > vClipRect.w)) {
        discard;
    }

    vec2 uv = atlasUV(vTexCoords, uTexBounds);
    
    // GBA resolution 240x160 for grid alignment
    vec2 res = vec2(240.0, 160.0);
    vec2 pixelUV = floor(uv * res);
    
    // Smooth the progress to avoid jarring start/stop
    float smoothP = smoothstep(0.0, 1.0, uProgress);
    
    // Transition progress: 0.0 to 1.0
    // Instability peaks in the middle
    float instability = 4.0 * smoothP * (1.0 - smoothP);
    
    // Time steps for snappy, non-analog feel
    // Using a slightly lower rate for "glitch state" updates makes it feel more readable
    float timeStep = floor(uTime * 12.0);
    
    // Sporadic bursts: glitches don't happen every frame, creating a more "unstable hardware" feel
    float burst = step(0.2, rand(timeStep)); 
    float glitchPower = instability * burst;
    
    vec2 offsetUV = uv;
    
    // 1. Chunked Row Slides (Horizontal shearing)
    // Multiple layers of row slides with different sizes and intensities for non-uniformity
    // Layer A: Small rows
    float rowGroupSmall = floor(pixelUV.y / 2.0);
    if (rand(vec2(rowGroupSmall, timeStep * 0.98)) < glitchPower * 0.3) {
        float slide = (rand(vec2(rowGroupSmall, timeStep + 11.0)) - 0.5) * 0.3 * glitchPower;
        offsetUV.x += floor(slide * res.x) / res.x;
    }
    // Layer B: Coarse rows (The requested "lines shifting horizontally")
    float rowGroupLarge = floor(pixelUV.y / 10.0);
    if (rand(vec2(rowGroupLarge, timeStep * 1.02)) < glitchPower * 0.5) {
        float slide = (rand(vec2(rowGroupLarge, timeStep + 17.0)) - 0.5) * 0.7 * glitchPower;
        offsetUV.x += floor(slide * res.x) / res.x;
    }
    
    // 2. Vertical Column Address Jumps
    // Coarser column groups for more structured vertical glitches
    float colGroup = floor(pixelUV.x / 10.0);
    if (rand(vec2(colGroup, timeStep * 1.1)) < glitchPower * 0.2) {
        float jump = (rand(vec2(colGroup, timeStep * 0.9)) - 0.5) * 0.2 * glitchPower;
        offsetUV.y += floor(jump * res.y) / res.y;
    }
    
    // 3. Block Corruption (Data chunk displacement)
    // Non-uniform block sizes across the screen
    vec2 sizeSeed = floor(pixelUV / 32.0);
    float rSize = rand(sizeSeed);
    vec2 blockSize = vec2(8.0 + floor(rSize * 24.0), 4.0 + floor(rand(sizeSeed + 0.5) * 12.0));
    vec2 blockUV = floor(uv * res / blockSize);
    if (rand(blockUV + vec2(timeStep * 0.8)) < glitchPower * 0.15) {
        vec2 blockOffset = vec2(rand(blockUV.x + timeStep), rand(blockUV.y + timeStep)) - 0.5;
        offsetUV += floor(blockOffset * 30.0 * glitchPower) / res;
    }

    // 4. Large Vertical Address Jumps (Screen tearing)
    if (rand(timeStep * 1.3) < glitchPower * 0.12) {
        float globalJump = floor((rand(timeStep * 0.4) - 0.5) * 40.0) / res.y;
        offsetUV.y += globalJump;
    }
    
    // Clamp and snap to pixel grid to ensure hard edges
    offsetUV = clamp(offsetUV, 0.0, 1.0);
    offsetUV = (floor(offsetUV * res) + 0.5) / res;
    
    // Sample both textures
    vec4 fromTex = texture(uBackgroundTexture, offsetUV);
    vec4 toTex = texture(uTexture, offsetUV);
    
    // 5. Digital Mix Logic
    // Instead of a smooth fade, use block-based and erratic swapping
    float blockSwap = rand(blockUV);
    
    // The "scanline" of the transition moves from bottom to top (or vice versa) but is jagged
    float transitionFront = smoothP + (rand(blockUV.y + timeStep) - 0.5) * 0.2 * instability;
    float mixFactor = step(blockSwap, transitionFront);
    
    // During high instability, add chaotic flickering of individual data chunks
    if (glitchPower > 0.4) {
        if (rand(blockUV + vec2(timeStep * 1.5)) < glitchPower * 0.3) {
            mixFactor = 1.0 - mixFactor;
        }
    }
    
    // 6. Black Tears (Dead data segments)
    // Interspersing black areas helps reduce "muddiness" during the swap
    float blackTear = 0.0;
    if (rand(vec2(rowGroupLarge, timeStep + 13.0)) < glitchPower * 0.1) {
        blackTear = 1.0;
    }
    // Occasionally whole blocks go black
    if (rand(blockUV + vec2(timeStep * 0.7)) < glitchPower * 0.05) {
        blackTear = 1.0;
    }

    vec4 tex = mix(fromTex, toTex, mixFactor);
    // Multiply blackTear by instability to ensure it only happens mid-transition
    tex = mix(tex, vec4(0.0, 0.0, 0.0, 1.0), blackTear * instability);

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
