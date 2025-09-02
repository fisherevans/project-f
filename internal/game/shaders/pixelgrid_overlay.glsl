#version 330 core

// from pixel
in vec4  vColor;
in vec2  vTexCoords;
in float vIntensity;
in vec4  vClipRect;

out vec4 fragColor;

// from pixel
uniform vec4 uColorMask;
uniform vec4 uTexBounds;
uniform sampler2D uTexture;

// knobs
uniform float uPixelSize; // fake pixel dimension in screen pixels
uniform float uScanlineDarken; // 0..1 how dark scanlines are (every other row)
uniform float uGridDarkenX; // 0..1 how dark vertical lines are
uniform float uGridDarkenY; // 0..1 how dark horizontal lines are
uniform float uSubpixelTint; // 0..1 tint strength per RGB

vec3 getSubpixelTintOld(int px, int py, int psize) {
    if (psize == 1) {
        return vec3(1);
    }

    // diagonal RGB layout
    if (psize % 2 == 0) {
        int lx = px / psize / 2;
        int ly = py / psize / 2;
        if (lx == 0 && ly == 0) return vec3(1, 0, 0); // Red
        if (lx == 1 && ly == 0) return vec3(0, 1, 0); // Green
        if (lx == 0 && ly == 1) return vec3(0, 1, 0); // Green
        return vec3(0, 0, 1); // Blue
    }

    // else, vertical lines
    int stripe = px % 3;
    if (stripe == 0) return vec3(1, 0, 0); // Red
    if (stripe == 1) return vec3(0, 1, 0); // Green
    return vec3(0, 0, 1); // Blue
}

vec3 getSubpixelTint(int px, int py, int psize) {
    if (psize <= 1) {
        return vec3(1.0);
    }

    // Even sizes: 2x2 quadrant (TL=R, TR=G, BL=G, BR=B), scaled to psize
    if ((psize & 1) == 0) {
        int halfPsize = psize / 2; // size of each quadrant
        int lx = px / halfPsize; // 0 = left half, 1 = right half
        int ly = py / halfPsize; // 0 = top half, 1 = bottom half
        if (lx == 0 && ly == 0) return vec3(1, 0, 0); // top-left: red
        if (lx == 1 && ly == 0) return vec3(0, 1, 0); // top-right: green
        if (lx == 0 && ly == 1) return vec3(0, 1, 0); // bottom-left: green
        return vec3(0, 0, 1); // bottom-right: blue
    }

    // Odd sizes: 3 vertical bands (R,G,B) across the fake pixel width
    int band = (px * 3) / psize;   // maps [0..psize-1] -> {0,1,2}
    if (band == 0) return vec3(1, 0, 0); // left: red
    if (band == 1) return vec3(0, 1, 0); // middle: green
    return vec3(0, 0, 1); // right: blue
}

void main() {
    // top section is stolen from the base pixel shader
    if ((vClipRect != vec4(0)) &&
        (gl_FragCoord.x < vClipRect.x || gl_FragCoord.y < vClipRect.y ||
         gl_FragCoord.x > vClipRect.z || gl_FragCoord.y > vClipRect.w)) {
        discard;
    }

    vec4 baseColor;
    if (vIntensity == 0.0) {
        baseColor = vColor;
    } else {
        vec2 texCoord = (vTexCoords - uTexBounds.xy) / uTexBounds.zw;
        vec4 texSample = texture(uTexture, texCoord);
        baseColor = mix(vColor, vColor * texSample, vIntensity);
    }

    // this is where the custom shader begins
    if (uPixelSize <= 1.0) {
        fragColor = baseColor * uColorMask;
        return;
    }

    // simulate "fake" pixels
    int psize = int(uPixelSize + 0.5);
    int px = int(mod(gl_FragCoord.x, uPixelSize));
    int py = int(mod(gl_FragCoord.y, uPixelSize));

    // Scanline darkening
    if (mod(floor(gl_FragCoord.y / uPixelSize), 2.0) == 1.0) {
        baseColor.rgb *= 1.0 - uScanlineDarken;
    }

    // Grid darkening per direction
    if (px == 0) {
        baseColor.rgb *= 1.0 - uGridDarkenX;
    }
    if (py == 0) {
        baseColor.rgb *= 1.0 - uGridDarkenY;
    }

    // Subpixel tinting
    vec3 tint = getSubpixelTint(px, py, psize);
    vec3 tinted = baseColor.rgb * tint;
    baseColor.rgb = mix(baseColor.rgb, tinted, uSubpixelTint);

    fragColor = baseColor * uColorMask;
}