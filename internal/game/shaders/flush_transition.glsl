#version 330 core

in vec4  vColor;
in vec2  vTexCoords;
in float vIntensity;
in vec4  vClipRect;

out vec4 fragColor;

uniform vec4 uColorMask;
uniform vec4 uTexBounds;
uniform sampler2D uTexture;           // To texture (Overworld)
uniform sampler2D uBackgroundTexture;  // From texture (Combat)
uniform float uProgress;

vec2 atlasUV(vec2 tc, vec4 win){ return (tc - win.xy) / win.zw; }

float rand(vec2 co) {
    return fract(sin(dot(co.xy ,vec2(12.9898,78.233))) * 43758.5453);
}

void main() {
    if ((vClipRect != vec4(0,0,0,0)) &&
        (gl_FragCoord.x < vClipRect.x || gl_FragCoord.y < vClipRect.y ||
         gl_FragCoord.x > vClipRect.z || gl_FragCoord.y > vClipRect.w)) {
        discard;
    }

    vec2 uv = atlasUV(vTexCoords, uTexBounds);
    vec2 res = vec2(240.0, 160.0);
    vec2 pixelPos = floor(uv * res);
    
    // Grid-aligned tiles (16x16)
    vec2 tileSize = vec2(16.0);
    vec2 tileID = floor(pixelPos / tileSize);
    vec2 pixelInTile = mod(pixelPos, tileSize);
    
    vec4 color = vec4(0.0, 0.0, 0.0, 1.0); // Pure black separation

    if (uProgress < 0.45) {
        // Phase 1: Buffer Drain (Combat exit)
        float p1 = uProgress / 0.45;
        
        // Stagger and move tiles off-screen
        float stagger = rand(tileID) * 0.2;
        float tileP = clamp((p1 - stagger) / 0.8, 0.0, 1.0);
        
        if (tileP < 1.0) {
            // Shrink and Slide Down
            float shrink = 1.0 - tileP;
            float fall = floor(tileP * tileP * res.y);
            
            // Snap to whole pixels
            vec2 offsetInTile = floor((pixelInTile - (tileSize * 0.5)) / max(shrink, 0.0001) + (tileSize * 0.5));
            
            if (offsetInTile.x >= 0.0 && offsetInTile.x < tileSize.x &&
                offsetInTile.y >= 0.0 && offsetInTile.y < tileSize.y) {
                
                vec2 samplePixel = (tileID * tileSize) + offsetInTile;
                samplePixel.y += fall; // Move DOWN (sample from HIGHER)
                
                if (samplePixel.y < res.y) {
                    vec2 sampleUV = (samplePixel + 0.5) / res;
                    color = texture(uBackgroundTexture, sampleUV);
                }
            }
        }
    } else if (uProgress > 0.6) {
        // Phase 3: Overworld Restore
        float p3 = (uProgress - 0.6) / 0.4;
        
        // Tile Fill Reconstruction (Top-to-Bottom)
        // 160 / 16 = 10 tiles high (indices 0 to 9)
        float rowThreshold = (9.0 - tileID.y) / 10.0;
        
        // Slight horizontal variation
        float jitter = (rand(vec2(tileID.x, 1.0)) - 0.5) * 0.1;
        
        if (p3 > clamp(rowThreshold + jitter, 0.0, 1.0)) {
            color = texture(uTexture, uv);
        }
    }
    // Phase 2 (0.45 to 0.6) is naturally black because color defaults to black
    
    fragColor = color * uColorMask * vColor;
}
