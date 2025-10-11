package textbox

import (
	"math"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"github.com/gopxl/pixel/v2/ext/text"

	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

type Instance struct {
	resources.FontInstance
	text *text.Text
	imd  *imdraw.IMDraw
	cfg  tbcfg.Config
}

func NewInstance(font resources.FontInstance, cfg tbcfg.Config) *Instance {
	return &Instance{
		FontInstance: font,
		text:         text.New(pixel.ZV, font.Atlas),
		imd:          imdraw.New(nil),
		cfg:          cfg,
	}
}

func (tb *Instance) GetConfig() tbcfg.Config {
	return tb.cfg
}

type characterRenderParams struct {
	drawDelta  pixel.Vec
	foreground pixel.RGBA
}

func (tb *Instance) Render(target pixel.Target, matrix pixel.Matrix, content *Content, opts ...tbcfg.ConfigOpt) {
	tb.text.Clear()

	// per render cfg overrides
	originalConfig := tb.cfg
	defer func() {
		tb.cfg = originalConfig
	}()
	for _, opt := range opts {
		opt(&tb.cfg)
	}

	pageLines := content.pageLines()
	renderLineCount := len(pageLines)
	if tb.cfg.LinesPerPage != 0 {
		renderLineCount = tb.cfg.LinesPerPage
	}

	var width, height int
	switch tb.cfg.ExpandMode {
	case tbcfg.ExpandFull:
		width = tb.cfg.BoxWidth
		height = tb.cfg.BoxHeight
	case tbcfg.ExpandFit:
		width = content.width
		height = content.height
	}

	switch tb.cfg.Origin {
	case gfx.BottomLeft:
		// do nothing
	case gfx.BottomRight:
		matrix = matrix.Moved(gfx.IVec(-width, 0))
	case gfx.TopRight:
		matrix = matrix.Moved(gfx.IVec(-width, -height))
	case gfx.Centered:
		matrix = matrix.Moved(gfx.IVec(-width/2, -height/2))
	case gfx.TopLeft:
		matrix = matrix.Moved(gfx.IVec(0, -height))
	case gfx.LeftCenter:
		matrix = matrix.Moved(gfx.IVec(0, -height/2))
	case gfx.TopCenter:
		matrix = matrix.Moved(gfx.IVec(-width/2, -height))
	case gfx.RightCenter:
		matrix = matrix.Moved(gfx.IVec(-width, -height/2))
	case gfx.BottomCenter:
		matrix = matrix.Moved(gfx.IVec(-width/2, 0))
	default:
		panic("invalid origin")
	}

	switch tb.cfg.VAlignment {
	case tbcfg.VAlignTop:
		matrix = matrix.Moved(gfx.IVec(0, height-content.height))
	case tbcfg.VAlignMiddle:
		matrix = matrix.Moved(pixel.V(0, math.Ceil(float64(height-content.height)/2.0)))
	case tbcfg.VAlignBottom:
		// do nothing
		//matrix = matrix.Moved(gfx.IVec(0, -height))
	}

	tb.imd.Clear()

	scrollDy := int((content.scrollPosition - float64(content.startLine)) * float64(tb.Metadata.LetterHeight+tb.effectiveLineSpacing()))

	for lineId, line := range pageLines {
		lineTypingProgress := 0
		y := float64(((renderLineCount - 1 - lineId) * (tb.Metadata.LetterHeight + tb.effectiveLineSpacing())) + tb.Metadata.TailHeight + scrollDy)
		var x int
		alignment := tb.cfg.HAlignment
		if content.alignmentOverride != nil {
			alignment = *content.alignmentOverride
		}
		switch alignment {
		case tbcfg.HAlignLeft:
			x = 0
		case tbcfg.HAlignCenter:
			x = (width - line.width) / 2
		case tbcfg.HAlignRight:
			x = width - line.width
		}
		tb.text.Dot = pixel.V(float64(x), y)
		var underlineColor pixel.RGBA
		var underlineStart, underlineEnd *pixel.Vec
		drawUnderline := func() {
			if underlineStart == nil || underlineEnd == nil {
				return
			}
			tb.imd.Color = underlineColor
			tb.imd.Push(*underlineStart, *underlineEnd)
			tb.imd.Line(1)
		}
		for _, c := range line.characters {
			if lineTypingProgress >= line.typingDone {
				break
			}
			// effect
			renderParams := &characterRenderParams{
				foreground: tb.cfg.Foreground,
			}
			if c.style.color != nil {
				renderParams.foreground = c.style.color.foreground
			}
			for _, effect := range c.style.effects {
				effect.Apply(renderParams)
			}
			tb.text.Dot = tb.text.Dot.Add(renderParams.drawDelta)
			// underline
			if c.style.underline != nil {
				// todo if outline, add 1px left and right to underline
				listCharStart := matrix.Project(tb.text.Dot).Add(pixel.V(-1, -1))
				if underlineStart == nil {
					underlineStart = &listCharStart
					if c.style.underline.colorDefined {
						underlineColor = c.style.underline.color
					} else {
						underlineColor = renderParams.foreground
					}
				}
				extraLength := 0.0
				if c.style.shadow != nil {
					extraLength = 1
				}
				listCharEnd := listCharStart.Add(pixel.V(float64(c.width)+2+extraLength, 0))
				underlineEnd = &listCharEnd
			} else if underlineStart != nil {
				drawUnderline()
			}
			type render struct {
				dx, dy int
				color  pixel.RGBA
			}
			var renders []render
			// outline
			if c.style.outline != nil {
				for dx := -1; dx <= 1; dx++ {
					for dy := -1; dy <= 1; dy++ {
						if dx == 0 && dy == 0 {
							continue
						}
						renders = append(renders, render{
							dx:    dx,
							dy:    dy,
							color: c.style.outline.color,
						})
					}
				}
			}
			// shadow
			if c.style.shadow != nil {
				for dx := 0; dx <= 1; dx++ {
					for dy := 0; dy >= -1; dy-- {
						if dx == 0 && dy == 0 {
							continue
						}
						renders = append(renders, render{
							dx:    dx,
							dy:    dy,
							color: c.style.shadow.color,
						})
					}
				}
			}
			// actual text
			renders = append(renders, render{
				color: renderParams.foreground,
			})
			origDot := tb.text.Dot
			for _, r := range renders {
				tb.text.Dot = origDot.Add(pixel.V(float64(r.dx), float64(r.dy)))
				tb.text.Color = r.color
				tb.text.WriteByte(c.c)
			}
			// undo effect delta
			tb.text.Dot = tb.text.Dot.Sub(renderParams.drawDelta)
			x += c.width
			lineTypingProgress += c.typingWeight
		}
		drawUnderline()
	}

	tb.imd.Draw(target)
	tb.text.Draw(target, matrix)
}

func (tb *Instance) effectiveLineSpacing() int {
	return tb.Metadata.LineSpacing + tb.cfg.ExtraLineSpacing
}
