package adventure

import (
	"math"
	"math/rand"

	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
)

// Small epsilon used for boundary fudge factors
const epsilon = 0.001

type ShadowMobState string

const ShadowMobStateWandering ShadowMobState = "wandering"
const ShadowMobStateChasing ShadowMobState = "chasing"

type ShadowMob struct {
	Location pixel.Vec

	animations map[ShadowMobState]*anim.AnimatedSprite

	// Smooth wandering
	vel         pixel.Vec // current velocity in tiles/sec
	targetVel   pixel.Vec // where we're steering toward
	changeTimer float64   // seconds until we pick a new targetVel

	// Tunables
	minSpeed  float64 // tiles/sec
	maxSpeed  float64 // tiles/sec
	steerLerp float64 // 0..1 per second, how quickly vel chases targetVel
	minChange float64 // min seconds between direction changes
	maxChange float64 // max seconds between direction changes
	radius    float64 // collision radius in tiles for movementRestrictions

	// Leash: keep wandering within a circle from spawn
	origin      pixel.Vec
	leashRadius float64 // tiles; 0 = disabled

	// State
	state ShadowMobState

	// Sensing: detection radius that slowly wanders between [senseMin, senseMax]
	senseCurrent     float64
	senseTarget      float64
	senseMin         float64
	senseMax         float64
	senseLerpPerSec  float64 // how fast senseCurrent chases senseTarget
	senseChangeTimer float64
	senseChangeMin   float64
	senseChangeMax   float64

	// Chasing behavior
	chaseBias        float64 // 0..1: how much to bias heading toward player vs random
	chaseSpeedMinMul float64 // scales minSpeed while chasing
	chaseSpeedMaxMul float64 // scales maxSpeed while chasing
	chaseSteerLerp   float64 // override steering while chasing (0..inf)
	chaseExitFactor  float64 // hysteresis: exit when distance > senseCurrent*factor
	chaseChangeMin   float64 // min seconds between chase retargets
	chaseChangeMax   float64 // max seconds between chase retargets

	// Triggering: fire a hook when close to player (e.g., to start combat)
	triggerRadius   float64                      // tiles; 0=disabled
	triggerCooldown float64                      // seconds between triggers
	triggerTimer    float64                      // counts down
	OnTrigger       func(s *State, m *ShadowMob) // optional callback
}

// ShadowMobConfig exposes the high-level knobs you likely want to set at spawn time.
// Everything else uses sensible internal defaults and can still be tweaked on the
// instance after construction if needed.
type ShadowMobConfig struct {
	LeashRadius   float64
	SenseMin      float64
	SenseMax      float64
	TriggerRadius float64
	OnTrigger     func(s *State, m *ShadowMob)
}

// DefaultShadowMobConfig returns a config pre-populated with balanced defaults.
func DefaultShadowMobConfig() ShadowMobConfig {
	return ShadowMobConfig{
		LeashRadius:   5,
		SenseMin:      3.5,
		SenseMax:      4.5,
		TriggerRadius: 0.8,
		OnTrigger: func(s *State, m *ShadowMob) {
			if game.DebugToggles().F5().ToggleState() {
				s.ExecuteSystemEffects(events.Effect{
					Plan: &events.EffectPlan{
						Steps: []events.PlanStep{
							{
								Serial: []events.Effect{
									{
										TriggerCombat: &events.EffectTriggerCombat{
											Opponent:   nil,
											Background: "combat/background_sylvoria",
										},
									},
									{
										Timer: &events.EffectTimer{
											DurationSeconds: 15,
										},
										Function: &events.EffectFunction{
											Fn: func() {
												for i, mob := range s.mobs {
													if mob == m {
														s.mobs[i] = s.mobs[len(s.mobs)-1]
														s.mobs = s.mobs[:len(s.mobs)-1]
														break
													}
												}
											},
										},
									},
									{
										Function: &events.EffectFunction{
											Fn: func() {
												s.mobs = append(s.mobs, m)
											},
										},
									},
								},
							},
						},
					},
				})
			} else {
				game.DebugNotification("Mob caught you! Toggle F5")
			}
		},
	}
}

// NewShadowMobWithConfig constructs a ShadowMob focusing on the key knobs; all
// other tunables are set to internal defaults that you can modify later if desired.
func NewShadowMobWithConfig(entityId string, location MapLocation, cfg *ShadowMobConfig) *ShadowMob {
	c := DefaultShadowMobConfig()
	if cfg != nil {
		// Override provided fields (zero values are treated as explicit, except OnTrigger)
		c.LeashRadius = cfg.LeashRadius
		if cfg.SenseMin != 0 {
			c.SenseMin = cfg.SenseMin
		}
		if cfg.SenseMax != 0 {
			c.SenseMax = cfg.SenseMax
		}
		if cfg.TriggerRadius != 0 {
			c.TriggerRadius = cfg.TriggerRadius
		}
		if cfg.OnTrigger != nil {
			c.OnTrigger = cfg.OnTrigger
		}
	}

	// All main tunables are grouped for clarity. Wandering movement controls organic patrolling;
	// leash/aggro/sensing/chase/trigger sections are tuned for fun and variety.
	m := &ShadowMob{
		Location: pixel.V(float64(location.X), float64(location.Y)),
		animations: map[ShadowMobState]*anim.AnimatedSprite{
			ShadowMobStateWandering: anim.LoadTilesheetAnimation(atlas, "adventure/entities/shadow_mob/shadow_mob", "wandering"),
			ShadowMobStateChasing:   anim.LoadTilesheetAnimation(atlas, "adventure/entities/shadow_mob/shadow_mob", "chasing"),
		},

		// --- Wandering movement tunables (internal defaults) ---
		minSpeed:  0.4, // tiles/sec; lower bound for drift speed
		maxSpeed:  1.9, // tiles/sec; upper bound for drift speed
		steerLerp: 7.0, // how quickly `vel` eases toward `targetVel` (higher = snappier)
		minChange: 0.7, // seconds; min time before picking a new wander heading
		maxChange: 2.2, // seconds; max time before picking a new wander heading
		radius:    0.5, // collision radius in tiles used by canTraverse

		// Leash / origin (keeps the mob near its spawn)
		origin:      pixel.V(float64(location.X), float64(location.Y)),
		leashRadius: c.LeashRadius, // 0 disables; otherwise soft-clamped in `leashClamp`

		// state
		state: ShadowMobStateWandering,

		// --- Sensing defaults (detection/aggro range slowly drifts for variety) ---
		senseMin:        c.SenseMin, // lower bound of detection radius
		senseMax:        c.SenseMax, // upper bound of detection radius
		senseLerpPerSec: 0.5,        // easing speed when drifting senseCurrent -> senseTarget
		senseChangeMin:  2.0,        // seconds; min time between sense target swaps
		senseChangeMax:  5.0,        // seconds; max time between sense target swaps

		// --- Chase defaults (behavior when aggro'd) ---
		chaseBias:        0.95, // 0..1; 0 = pure random, 1 = pure toward-player
		chaseSpeedMinMul: 1.5,  // scales minSpeed while chasing
		chaseSpeedMaxMul: 2,    // scales maxSpeed while chasing
		chaseSteerLerp:   10.0, // steering rate while chasing (more responsive than wander)
		chaseExitFactor:  1.25, // hysteresis; drop aggro when dist > senseCurrent*factor
		chaseChangeMin:   0.15, // seconds; min time between chase retargets
		chaseChangeMax:   0.35, // seconds; max time between chase retargets

		// --- Trigger defaults (proximity/combat start) ---
		triggerRadius:   c.TriggerRadius, // tiles; 0 disables triggers entirely
		triggerCooldown: 1.5,             // seconds between successive triggers
		OnTrigger:       c.OnTrigger,     // optional callback
	}

	// Timers and sense values are seeded to avoid any pop-in or awkward startup motion.
	// Start with a random target velocity and zero actual velocity for a smooth ease-in.
	m.pickNewTarget()
	m.changeTimer = randRange(m.minChange, m.maxChange)

	// Initialize sensing band to a sane starting point (no pop-in on first frame)
	m.senseTarget = randRange(m.senseMin, m.senseMax)
	m.senseCurrent = m.senseTarget
	m.senseChangeTimer = randRange(m.senseChangeMin, m.senseChangeMax)

	// Allow immediate triggering if the player spawns on top, then apply cooldown thereafter
	m.triggerTimer = 0
	return m
}

func NewShadowMob(entityId string, location MapLocation) *ShadowMob {
	cfg := DefaultShadowMobConfig()
	return NewShadowMobWithConfig(entityId, location, &cfg)
}

// --- Internal helpers to keep Update() focused ---

// playerPosFromState isolates how we fetch the player's position.
func playerPosFromState(s *State) pixel.Vec {
	ml := s.entities.GetEntity(s.player).PreciseLocation()
	return pixel.V(float64(ml.X), float64(ml.Y))
}

// resetRetargetTimer seeds the retarget timer based on current state.
func (m *ShadowMob) resetRetargetTimer() {
	if m.state == ShadowMobStateChasing {
		m.changeTimer = randRange(m.chaseChangeMin, m.chaseChangeMax)
	} else {
		m.changeTimer = randRange(m.minChange, m.maxChange)
	}
}

// retarget chooses a new targetVel according to state.
func (m *ShadowMob) retarget(toPlayer pixel.Vec) {
	if m.state == ShadowMobStateChasing {
		m.pickNewChaseTarget(toPlayer)
	} else {
		m.pickNewTarget()
	}
	m.resetRetargetTimer()
}

// steer advances velocity toward targetVel using state-appropriate steering.
func (m *ShadowMob) steer(dt float64) {
	steer := m.steerLerp
	if m.state == ShadowMobStateChasing {
		steer = m.chaseSteerLerp
	}
	// Exponential smoothing keeps motion framerate-independent and prevents jerkiness
	t := 1 - math.Exp(-steer*dt)
	m.vel = m.vel.Add(m.targetVel.Sub(m.vel).Scaled(t))
}

// proposeMove integrates position using current velocity.
func (m *ShadowMob) proposeMove(dt float64) pixel.Vec {
	return m.Location.Add(m.vel.Scaled(dt))
}

// leashClamp optionally clamps the proposed position to the leash, and updates targetVel to bias inward (or toward player while chasing).
func (m *ShadowMob) leashClamp(proposed pixel.Vec, toPlayer pixel.Vec) (pixel.Vec, bool) {
	if m.leashRadius <= 0 {
		return proposed, false
	}
	d := proposed.Sub(m.origin)
	dist := d.Len()
	if dist <= m.leashRadius {
		return proposed, false
	}
	// Snap to boundary and steer inward
	if dist > 0 {
		n := d.Scaled(1.0 / dist)
		proposed = m.origin.Add(n.Scaled(m.leashRadius - epsilon))
		inward := n.Scaled(-randRange(m.minSpeed, m.maxSpeed))
		if m.state == ShadowMobStateChasing {
			toward := toPlayer
			if l := toward.Len(); l > 0 {
				toward = toward.Scaled(1.0 / l)
			}
			blended := inward.Scaled(0.4).Add(toward.Scaled(0.6))
			if l := blended.Len(); l > 0 {
				m.targetVel = blended.Scaled(randRange(m.minSpeed*m.chaseSpeedMinMul, m.maxSpeed*m.chaseSpeedMaxMul))
			} else {
				m.targetVel = inward
			}
		} else {
			m.targetVel = inward
		}
		m.resetRetargetTimer()
	} else {
		proposed = m.origin
	}
	return proposed, true
}

// moveOrBounce tries to move; on failure it retargets appropriately.
func (m *ShadowMob) moveOrBounce(s *State, proposed pixel.Vec, toPlayer pixel.Vec) {
	if m.canTraverse(s, m.Location, proposed, m.radius) {
		m.Location = proposed
		return
	}
	// blocked: pick a new heading based on state
	m.retarget(toPlayer)
}

func (m *ShadowMob) Update(s *State, timeDelta float64) {
	if timeDelta <= 0 {
		return
	}

	game.DebugBL("mob state: %s", m.state)

	// Cooldown for trigger hook
	if m.triggerTimer > 0 {
		m.triggerTimer -= timeDelta
		if m.triggerTimer < 0 {
			m.triggerTimer = 0
		}
	}

	playerPos := playerPosFromState(s)

	if a, exists := m.animations[m.state]; exists {
		a.Update(timeDelta)
	}

	// Update wandering detection radius drift
	m.updateSenseRadius(timeDelta)

	// Count down to the next direction change
	m.changeTimer -= timeDelta

	toPlayer := playerPos.Sub(m.Location)
	distToPlayer := toPlayer.Len()

	// Proximity trigger (independent of state)
	if m.triggerRadius > 0 && distToPlayer <= m.triggerRadius && m.triggerTimer == 0 {
		if m.OnTrigger != nil {
			m.OnTrigger(s, m)
		}
		m.triggerTimer = m.triggerCooldown
	}

	// --- state transitions ---
	switch m.state {
	case ShadowMobStateWandering:
		if distToPlayer <= m.senseCurrent {
			m.state = ShadowMobStateChasing
			// seed a chase heading
			m.pickNewChaseTarget(toPlayer)
			// tighten steering while chasing
			m.resetRetargetTimer()
		}
	case ShadowMobStateChasing:
		if distToPlayer > m.senseCurrent*m.chaseExitFactor {
			m.state = ShadowMobStateWandering
			m.pickNewTarget() // resume wandering target
			m.resetRetargetTimer()
		}
	}

	if m.changeTimer <= 0 {
		m.retarget(toPlayer)
	}

	// Smoothly steer current velocity toward target velocity
	m.steer(timeDelta)

	// Integrate position in floating space (not snapped to tiles), then leash clamp
	proposed := m.proposeMove(timeDelta)
	if clamped, hit := m.leashClamp(proposed, toPlayer); hit {
		proposed = clamped
	}
	m.moveOrBounce(s, proposed, toPlayer)
}

// pickNewTarget chooses a new random heading and speed for targetVel.
func (m *ShadowMob) pickNewTarget() {
	angle := randRange(0, 2*math.Pi)
	speed := randRange(m.minSpeed, m.maxSpeed)
	dir := pixel.V(math.Cos(angle), math.Sin(angle))
	m.targetVel = dir.Scaled(speed)
}

func (m *ShadowMob) Render(batch *pixel.Batch, moved pixel.Matrix) {
	alpha := 0.7
	if m.state == ShadowMobStateChasing {
		alpha = 0.9
	}
	if a, exists := m.animations[m.state]; exists {
		a.Sprite().DrawColorMask(batch, moved, pixel.RGBA{alpha, alpha, alpha, alpha})
	}
}

// canTraverse determines whether the ShadowMob can move from `from` to `to` without
// crossing an occupied tile, treating the mob as a circle of `radius` tiles.
// `occTile` should return true if the integer tile [ix,iy] is fully occupied (1x1 square).
// If `occTile` is nil, traversal is always allowed.
func (m *ShadowMob) canTraverse(s *State, from, to pixel.Vec, radius float64) bool {
	// March along the segment in small steps; cheaper than continuous swept-circle, good enough here
	delta := to.Sub(from)
	dist := delta.Len()
	if dist == 0 {
		// Even with no movement, ensure we aren't currently intersecting a wall.
		return m.sampleCircleAgainstTiles(s, from, radius)
	}

	step := math.Min(0.25, radius*0.5)
	steps := int(math.Ceil(dist / step))
	dir := delta.Scaled(1.0 / float64(steps))
	p := from
	for i := 0; i <= steps; i++ {
		if !m.sampleCircleAgainstTiles(s, p, radius) {
			return false
		}
		p = p.Add(dir)
	}
	return true
}

// sampleCircleAgainstTiles checks this center position against all tiles that could intersect the circle.
func (m *ShadowMob) sampleCircleAgainstTiles(s *State, center pixel.Vec, radius float64) bool {
	// Tiles are 1x1 squares centered on integer coords: [ix-0.5, ix+0.5] x [iy-0.5, iy+0.5]
	// Cover all candidate tile centers within the circle's AABB.
	minX := int(math.Floor(center.X - radius + 0.5))
	maxX := int(math.Floor(center.X + radius + 0.5))
	minY := int(math.Floor(center.Y - radius + 0.5))
	maxY := int(math.Floor(center.Y + radius + 0.5))

	rr := radius * radius
	for ty := minY; ty <= maxY; ty++ {
		for tx := minX; tx <= maxX; tx++ {
			if s.canMobTraverse(tx, ty) {
				// Open tile; no need to test.
				continue
			}
			if tileCircleIntersects(center, rr, tx, ty) {
				return false
			}
		}
	}
	return true
}

func (s *State) canMobTraverse(x, y int) bool {
	// todo mob entity id? direction?
	return s.entities.isValidTransition("", MapLocation{X: x, Y: y}, input.Down, true)
}

// tileCircleIntersects checks if a circle at `center` with squared radius `rr` intersects a unit tile centered at (tx,ty), extents [tx-0.5,tx+0.5] x [ty-0.5,ty+0.5].
func tileCircleIntersects(center pixel.Vec, rr float64, tx, ty int) bool {
	// Clamp circle center to the AABB and measure distance
	cx := clampFloat(center.X, float64(tx)-0.5, float64(tx)+0.5)
	cy := clampFloat(center.Y, float64(ty)-0.5, float64(ty)+0.5)
	dx := center.X - cx
	dy := center.Y - cy
	return dx*dx+dy*dy <= rr
}

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// randRange returns a uniform random in [min,max]; if max<=min it returns min.
func randRange(min, max float64) float64 {
	if max <= min {
		return min
	}
	return min + rand.Float64()*(max-min)
}

// pickNewChaseTarget biases direction toward `toTarget` while keeping organic noise.
func (m *ShadowMob) pickNewChaseTarget(toTarget pixel.Vec) {
	// Random exploratory direction adds a bit of wobble so it doesn't bee-line perfectly
	randAngle := randRange(0, 2*math.Pi)
	randDir := pixel.V(math.Cos(randAngle), math.Sin(randAngle))

	// Desired toward player
	dir := toTarget
	if dir.Len() > 0 {
		dir = dir.Scaled(1.0 / dir.Len())
	}
	// Blend random with biased toward target
	blended := randDir.Scaled(1.0 - m.chaseBias).Add(dir.Scaled(m.chaseBias))
	if blended.Len() == 0 {
		blended = randDir
	}
	blended = blended.Scaled(1.0 / blended.Len())

	// Choose a slightly higher speed band while chasing
	minS := m.minSpeed * m.chaseSpeedMinMul
	maxS := m.maxSpeed * m.chaseSpeedMaxMul
	speed := randRange(minS, maxS)
	m.targetVel = blended.Scaled(speed)
}

// updateSenseRadius eases senseCurrent toward a drifting target within [senseMin, senseMax]
func (m *ShadowMob) updateSenseRadius(dt float64) {
	m.senseChangeTimer -= dt
	if m.senseChangeTimer <= 0 {
		m.senseTarget = randRange(m.senseMin, m.senseMax)
		m.senseChangeTimer = randRange(m.senseChangeMin, m.senseChangeMax)
	}
	// exponential approach similar to velocity steering
	t := 1 - math.Exp(-m.senseLerpPerSec*dt)
	m.senseCurrent = m.senseCurrent + (m.senseTarget-m.senseCurrent)*t
}
