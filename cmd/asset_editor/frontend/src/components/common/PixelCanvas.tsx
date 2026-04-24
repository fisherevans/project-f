import { forwardRef, useEffect, useImperativeHandle, useRef } from "react";

export interface PixelCanvasHandle {
    canvas: HTMLCanvasElement | null;
    ctx: CanvasRenderingContext2D | null;
}

interface PixelCanvasProps {
    width: number;
    height: number;
    className?: string;
    onDraw?: (ctx: CanvasRenderingContext2D, canvas: HTMLCanvasElement) => void;
    onClick?: (e: React.MouseEvent<HTMLCanvasElement>) => void;
    onMouseMove?: (e: React.MouseEvent<HTMLCanvasElement>) => void;
    onMouseLeave?: (e: React.MouseEvent<HTMLCanvasElement>) => void;
}

export const PixelCanvas = forwardRef<PixelCanvasHandle, PixelCanvasProps>(
    function PixelCanvas({ width, height, className, onDraw, onClick, onMouseMove, onMouseLeave }, ref) {
        const canvasRef = useRef<HTMLCanvasElement>(null);

        useImperativeHandle(ref, () => ({
            get canvas() {
                return canvasRef.current;
            },
            get ctx() {
                return canvasRef.current?.getContext("2d") ?? null;
            },
        }));

        useEffect(() => {
            const canvas = canvasRef.current;
            if (!canvas) return;
            const dpr = window.devicePixelRatio || 1;
            canvas.width = width * dpr;
            canvas.height = height * dpr;
            canvas.style.width = `${width}px`;
            canvas.style.height = `${height}px`;
            const ctx = canvas.getContext("2d");
            if (!ctx) return;
            ctx.scale(dpr, dpr);
            ctx.imageSmoothingEnabled = false;
            onDraw?.(ctx, canvas);
        }, [width, height, onDraw]);

        return (
            <canvas
                ref={canvasRef}
                className={className}
                style={{ imageRendering: "pixelated" }}
                onClick={onClick}
                onMouseMove={onMouseMove}
                onMouseLeave={onMouseLeave}
            />
        );
    },
);
