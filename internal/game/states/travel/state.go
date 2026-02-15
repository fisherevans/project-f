package travel

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/interp"
	"fisherevans.com/project/f/internal/util/keyframe"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"github.com/gopxl/pixel/v2"
)

type baseState struct {
	game.BaseState
	nextIntent any

	batch                               *pixel.Batch
	shipSprite, podSprite, planetSprite pixelutil.BoundedDrawable
	bgSprites                           []*bgSprite

	group      *keyframe.Group
	fadeIn     *keyframe.Member
	launch     *keyframe.Member
	accelerate *keyframe.Member
}

func NewTravelState(intent game.TravelIntent) game.State {
	group := keyframe.NewGroup(interp.Smoothstep)
	s := &baseState{
		nextIntent:   intent.ToIntent,
		shipSprite:   atlas.GetSprite("travel/ship"),
		podSprite:    atlas.GetSprite("travel/pod"),
		planetSprite: atlas.GetSprite(intent.PlanetSpriteName),
		batch:        atlas.NewBatch(),

		group:      group,
		fadeIn:     group.AddKeyFrame(0, 2).AddTransition(5, 7),
		launch:     group.AddKeyFrame(1, 4),
		accelerate: group.AddKeyFrame(3, 7),
	}
	for i := 0; i < 100; i++ {
		s.bgSprites = append(s.bgSprites, newStarBgSprite())
	}
	return s
}

func (b *baseState) ClearColor() pixel.RGBA {
	return colors.FromString("#180a31")
}

func (b *baseState) OnTick(target pixel.ComposeTarget, targetBounds pixel.Rect, timeDelta float64) {
	b.batch.Clear()
	b.group.Update(timeDelta)
	starTimeDelta := timeDelta * (1.0 + b.launch.Progress()*16 + b.accelerate.Progress()*16)
	for _, sprite := range b.bgSprites {
		sprite.Update(timeDelta, starTimeDelta)
		sprite.Render(b.batch, pixel.IM.Moved(sprite.position))
	}

	x := float64(game.GameWidth / 2)
	startY := float64(game.GameHeight/2) - 20
	b.podSprite.Draw(b.batch, pixel.IM.Moved(pixel.V(x, startY-b.launch.Progress()*40+b.accelerate.Progress()*100)))
	b.shipSprite.Draw(b.batch, pixel.IM.Moved(pixel.V(x, startY-b.launch.Progress()*170+b.fadeIn.Progress()*10)))

	fadeColor := colors.Lerp(colors.Black.RGBA, colors.Alpha(0), b.fadeIn.Progress())
	gfx.DrawRect(atlas, b.batch, pixel.IM, gfx.BottomLeft, game.GameWidth, game.GameHeight, fadeColor)

	b.batch.Draw(target)
	if b.group.Progress() >= 1 {
		game.SetActiveStateIntent(b.nextIntent)
	}
}
