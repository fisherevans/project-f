package adventure

import (
	"math"
	"math/rand"
	"slices"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/astar"
	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

// Small epsilon used for boundary fudge factors
const epsilon = 0.001

type ShadowMobState string

const ShadowMobStateWandering ShadowMobState = "wandering"
const ShadowMobStateChasing ShadowMobState = "chasing"
const ShadowMobStateTriggering ShadowMobState = "triggering"

type ShadowMob struct {
	Location pixel.Vec

	animations map[ShadowMobState]*anim.AnimatedSprite

	// Smooth movement
	vel            pixel.Vec // current velocity in tiles/sec
	accel          float64   // acceleration speed
	friction       float64   // friction applied to velocity
	targetLocation pixel.Vec // where we are heading
	targetReached  float64   // threshold to consider target reached

	// Tunables
	maxSpeed float64 // tiles/sec
	radius   float64 // collision radius in tiles for movementRestrictions

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
	chaseBias       float64 // 0..1: how much to bias heading toward player vs random
	chaseExitFactor float64 // hysteresis: exit when distance > senseCurrent*factor

	// Triggering: fire a hook when close to player (e.g., to start combat)
	triggerRadius   float64                      // tiles; 0=disabled
	triggerCooldown float64                      // seconds between triggers
	triggerTimer    float64                      // counts down
	triggerElapsed  float64                      // how long we've been in triggering state
	OnTrigger       func(s *State, m *ShadowMob) // optional callback

	// Visuals
	spriteRow         int
	colorMask         pixel.RGBA
	footstepDistance  float64 // configured base distance
	footstepCounter   float64
	footstepOffset    float64
	footstepGaitScale float64 // scaling factor for speed
	footstepSide      bool    // false=left, true=right

	// Pathfinding
	currentPath    []MapLocation
	pathTimer      float64
	lastPathTarget MapLocation
}

// ShadowMobConfig exposes the high-level knobs you likely want to set at spawn time.
// Everything else uses sensible internal defaults and can still be tweaked on the
// instance after construction if needed.
type ShadowMobConfig struct {
	LeashRadius       float64
	SenseMin          float64
	SenseMax          float64
	TriggerRadius     float64
	OnTrigger         func(s *State, m *ShadowMob)
	SpriteRow         int
	ColorMask         pixel.RGBA
	FootstepDistance  float64
	FootstepOffset    float64
	FootstepGaitScale float64
}

type ShadowMobParams struct {
	CombatId         string  `yaml:"combat_id"`
	BroadcastId      string  `yaml:"broadcast_id"`
	LeashRadius      float64 `yaml:"leash_radius"`
	RespawnDelay     float64 `yaml:"respawn_delay"`
	RespawnJitter    float64 `yaml:"respawn_jitter"`
	CombatBackground string  `yaml:"combat_background"`

	OpponentPool      string            `yaml:"opponent_pool"`
	Opponent          rpg.PrimortalType `yaml:"opponent"`
	OpponentArchetype string            `yaml:"opponent_archetype"`

	SpriteRow         int     `yaml:"sprite_row"`
	ColorMask         string  `yaml:"color_mask"`
	FootstepDistance  float64 `yaml:"footstep_distance"`
	FootstepOffset    float64 `yaml:"footstep_offset"`
	FootstepGaitScale float64 `yaml:"footstep_gait_scale"`
}

func NewShadowMobParamsFromProperties(props *util.Properties) ShadowMobParams {
	params := DefaultShadowMobParams()
	if props == nil {
		return params
	}
	props.LoadStructFromKey("shadow_mob", &params)
	triggerTypes := 0
	if params.CombatId != "" {
		triggerTypes++
	}
	if params.BroadcastId != "" {
		triggerTypes++
	}
	if params.OpponentPool != "" {
		triggerTypes++
	}
	if triggerTypes > 1 {
		log.Fatal().Msg("Cannot specify more than one trigger type for shadow mob.")
	}
	if triggerTypes == 0 {
		log.Fatal().Msg("Must specify exactly one trigger type for shadow mob.")
	}
	return params
}

func DefaultShadowMobParams() ShadowMobParams {
	return ShadowMobParams{
		LeashRadius:       5,
		RespawnDelay:      15,
		RespawnJitter:     0,
		CombatBackground:  "combat/background_sylvoria",
		SpriteRow:         1,
		ColorMask:         "#4d4d4d", // Greyscale(0.3) approx
		FootstepDistance:  0.5,
		FootstepOffset:    0.1,
		FootstepGaitScale: 0.5,
	}
}

// DefaultShadowMobConfig returns a config pre-populated with balanced defaults.
func DefaultShadowMobConfig(params ShadowMobParams) ShadowMobConfig {
	return ShadowMobConfig{
		LeashRadius:       params.LeashRadius,
		SenseMin:          3.5,
		SenseMax:          4.5,
		TriggerRadius:     0.8,
		SpriteRow:         params.SpriteRow,
		ColorMask:         colors.HexString(params.ColorMask),
		FootstepDistance:  params.FootstepDistance,
		FootstepOffset:    params.FootstepOffset,
		FootstepGaitScale: params.FootstepGaitScale,
		OnTrigger: func(s *State, m *ShadowMob) {
			if game.DebugToggles().F4().ToggleState() {
				game.DebugNotificationf("shadow mob trigger prevented due to f4 toggle")
				return
			}
			var effects []Effect
			if params.BroadcastId != "" {
				effects = append(effects, NewSendBroadcastEffect(params.BroadcastId, nil))
			} else {
				var opponent rpg.PrimortalType
				var opponentArchetype string
				if params.OpponentPool != "" {
					opponent, opponentArchetype = rpg.GetOpponentPool(params.OpponentPool).Random()
				} else {
					opponent = params.Opponent
					opponentArchetype = params.OpponentArchetype
				}
				triggerCombat := NewTriggerCombatEffect(params.CombatBackground).
					WithCombatId(params.CombatId).
					WithOpponent(game.NewCombatOpponent(opponent, opponentArchetype))
				effects = append(effects, triggerCombat)
			}
			effects = append(effects,
				NewFunctionEffect(func(_ EntityReader, _ *State) {
					for i, mob := range s.mobs {
						if mob == m {
							s.mobs[i] = s.mobs[len(s.mobs)-1]
							s.mobs = s.mobs[:len(s.mobs)-1]
							break
						}
					}
				}),
			)
			if params.RespawnDelay+params.RespawnJitter > 0 {
				effects = append(effects,
					NewTimerEffect(params.RespawnDelay+rand.Float64()*params.RespawnDelay),
					NewFunctionEffect(func(_ EntityReader, _ *State) {
						s.mobs = append(s.mobs, m)
					}))
			}
			s.ExecuteSystemEffectsInOrder(effects...)
		},
	}
}

// NewShadowMobWithConfig constructs a ShadowMob focusing on the key knobs; all
// other tunables are set to internal defaults that you can modify later if desired.
func NewShadowMobWithConfig(entityId string, location MapLocation, c *ShadowMobConfig) *ShadowMob {
	// All main tunables are grouped for clarity. Wandering movement controls organic patrolling;
	// leash/aggro/sensing/chase/trigger sections are tuned for fun and variety.
	m := &ShadowMob{
		Location: pixel.V(float64(location.X), float64(location.Y)),
		animations: map[ShadowMobState]*anim.AnimatedSprite{
			ShadowMobStateWandering:  anim.LoadTilesheetAnimation(atlas, "adventure/entities/shadow_mob/shadow_mob", "wandering"),
			ShadowMobStateChasing:    anim.LoadTilesheetAnimation(atlas, "adventure/entities/shadow_mob/shadow_mob", "chasing"),
			ShadowMobStateTriggering: anim.LoadTilesheetAnimation(atlas, "adventure/entities/shadow_mob/shadow_mob", "chasing"),
		},

		// --- Physics defaults ---
		maxSpeed:      1.5,
		accel:         12.0,
		friction:      8.0,
		radius:        0.4,
		targetReached: 0.1,

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
		chaseBias:       0.95, // 0..1; 0 = pure random, 1 = pure toward-player
		chaseExitFactor: 1.25, // hysteresis; drop aggro when dist > senseCurrent*factor

		// --- Trigger defaults (proximity/combat start) ---
		triggerRadius:   c.TriggerRadius, // tiles; 0 disables triggers entirely
		triggerCooldown: 1.5,             // seconds between successive triggers
		OnTrigger:       c.OnTrigger,     // optional callback

		// --- Visuals ---
		spriteRow:         c.SpriteRow,
		colorMask:         c.ColorMask,
		footstepDistance:  c.FootstepDistance,
		footstepOffset:    c.FootstepOffset,
		footstepGaitScale: c.FootstepGaitScale,
	}

	// Timers and sense values are seeded to avoid any pop-in or awkward startup motion.
	m.targetLocation = m.origin
	m.pickNewTarget(nil)

	// Initialize sensing band to a sane starting point (no pop-in on first frame)
	m.senseTarget = randRange(m.senseMin, m.senseMax)
	m.senseCurrent = m.senseTarget
	m.senseChangeTimer = randRange(m.senseChangeMin, m.senseChangeMax)

	// Allow immediate triggering if the player spawns on top, then apply cooldown thereafter
	m.triggerTimer = 0
	return m
}

func NewShadowMob(entityId string, location MapLocation, params ShadowMobParams) *ShadowMob {
	cfg := DefaultShadowMobConfig(params)
	return NewShadowMobWithConfig(entityId, location, &cfg)
}

// --- Internal helpers to keep Update() focused ---

func (m *ShadowMob) getPlayerPos(s *State) pixel.Vec {
	entity, ok := s.entities.GetEntity(s.player)
	if !ok {
		log.Error().Str("player", string(s.player)).Msg("player not found in state")
		return pixel.ZV
	}
	ml := entity.GetPreciseLocation()
	return pixel.V(float64(ml.X), float64(ml.Y))
}

func (m *ShadowMob) updatePath(s *State, target pixel.Vec) {
	targetLoc := MapLocation{X: int(math.Floor(target.X + 0.5)), Y: int(math.Floor(target.Y + 0.5))}
	startLoc := MapLocation{X: int(math.Floor(m.Location.X + 0.5)), Y: int(math.Floor(m.Location.Y + 0.5))}

	// Only recalculate if target has moved to a new tile, UNLESS we currently have no path
	// or if we are stuck (caller should handle stuck timer, but we ensure we recalculate if needed)
	if targetLoc == m.lastPathTarget && len(m.currentPath) > 0 && startLoc != m.lastPathTarget {
		// If we still have a path and the player is in the same place, we can usually stick to it.
		// However, check if our current target node is still valid if we've been blocked.
		return
	}

	m.lastPathTarget = targetLoc

	// ShadowMobs use a simpler pathfinding context than standard Entities.
	// We use the common PathNeighbors/PathHeuristic implementations in pathfinding.go via this context.
	ctx := shadowMobPathfindingContext{s: s}
	path, _, found := astar.Path[MapLocation, any](startLoc, targetLoc, ctx, 1000)
	if found && len(path) > 1 {
		slices.Reverse(path)
		m.currentPath = path[1:] // skip current tile
	} else if found && len(path) == 1 {
		// Already at the tile
		m.currentPath = nil
	} else {
		m.currentPath = nil
	}
}

type shadowMobPathfindingContext struct {
	s *State
}

// resetRetargetTimer seeds the retarget timer based on current state.
func (m *ShadowMob) Update(s *State, timeDelta float64) {
	if timeDelta <= 0 {
		return
	}

	game.DebugBLf("mob state: %s", m.state)

	// Cooldown for trigger hook
	if m.triggerTimer > 0 {
		m.triggerTimer -= timeDelta
		if m.triggerTimer < 0 {
			m.triggerTimer = 0
		}
	}

	playerPos := m.getPlayerPos(s)

	if a, exists := m.animations[m.state]; exists {
		a.Update(timeDelta)
	}

	// Update wandering detection radius drift
	m.updateSenseRadius(timeDelta)

	toPlayer := playerPos.Sub(m.Location)
	distToPlayer := toPlayer.Len()
	game.DebugTRf("dist: %.1f, path: %d, state: %s", distToPlayer, len(m.currentPath), m.state)

	// Proximity trigger (independent of state)
	if m.triggerRadius > 0 && distToPlayer <= m.triggerRadius && m.triggerTimer == 0 && !s.mobTriggering && m.state == ShadowMobStateChasing {
		s.mobTriggering = true
		m.state = ShadowMobStateTriggering
		m.triggerElapsed = 0
	}

	// --- state transitions and logic ---
	switch m.state {
	case ShadowMobStateWandering:
		if distToPlayer <= m.senseCurrent {
			m.state = ShadowMobStateChasing
			m.pathTimer = 0 // force immediate pathfind
		} else {
			// If we reached our target, or somehow got blocked, pick a new one
			distToTarget := m.targetLocation.Sub(m.Location).Len()
			if distToTarget < m.targetReached {
				m.pickNewTarget(s)
			}
		}
	case ShadowMobStateChasing:
		if distToPlayer > m.senseCurrent*m.chaseExitFactor {
			m.state = ShadowMobStateWandering
			m.currentPath = nil
			m.pickNewTarget(s)
		}
	case ShadowMobStateTriggering:
		m.triggerElapsed += timeDelta
		if distToPlayer < 0.05 {
			m.Location = playerPos
			if m.OnTrigger != nil {
				m.OnTrigger(s, m)
			}
			s.mobTriggering = false
			m.triggerTimer = m.triggerCooldown
			m.state = ShadowMobStateWandering
			m.currentPath = nil
			return // Triggered!
		}
	}

	// Pathfinding refresh for chasing/triggering
	if m.state == ShadowMobStateChasing || m.state == ShadowMobStateTriggering {
		m.pathTimer -= timeDelta
		// Refresh path if timer expired OR if we are currently not moving but have a target
		isStuck := m.vel.Len() < 0.2 && m.state != ShadowMobStateTriggering
		if m.pathTimer <= 0 || (isStuck && m.pathTimer < 0.1) {
			m.updatePath(s, playerPos)
			m.pathTimer = 0.2 + rand.Float64()*0.2 // 200-400ms refresh
		}

	}

	// Determine desired direction and speed
	var target pixel.Vec
	speed := m.maxSpeed

	if m.state == ShadowMobStateTriggering {
		speed = m.maxSpeed + m.triggerElapsed*4.0
		target = playerPos // Ignore pathfinding for triggering
	} else if m.state == ShadowMobStateChasing {
		// Dynamic Path Shortcutting: If we have a path, and we reached a node, check if we can see the player
		// directly from this new vantage point. If so, we can discard the rigid path.
		// We only do this when a node is reached to avoid discarding paths prematurely near corners.
		if len(m.currentPath) > 0 {
			target = m.currentPath[0].ToVec()
			// If we're close enough to the first node, pop it and head for the next
			if m.Location.Sub(target).Len() < 0.3 {
				m.currentPath = m.currentPath[1:]
				if len(m.currentPath) > 0 {
					target = m.currentPath[0].ToVec()
					// Since we just cleared a node, check if we have LOS to the player now
					if m.canTraverse(s, m.Location, playerPos, m.radius) {
						m.currentPath = nil
						target = playerPos
					}
				} else {
					target = playerPos
				}
			}
		} else {
			target = playerPos
		}
	} else { // ShadowMobStateWandering
		target = m.targetLocation
	}

	diff := target.Sub(m.Location)
	dist := diff.Len()
	var desiredVel pixel.Vec
	if dist > 0.01 {
		desiredVel = diff.Scaled(1.0 / dist).Scaled(speed)
	}

	// Apply acceleration and friction for slippery movement
	if desiredVel.Len() > 0 {
		// Use a higher acceleration when chasing to feel more responsive and maintain speed
		accel := m.accel
		if m.state != ShadowMobStateWandering {
			accel *= 2.0
		}
		m.vel = m.vel.Add(desiredVel.Sub(m.vel).Scaled(accel * timeDelta))
	} else {
		m.vel = m.vel.Add(m.vel.Scaled(-1).Scaled(m.friction * timeDelta))
	}

	// Clamp to current speed (which might be higher than m.maxSpeed in triggering)
	if m.vel.Len() > speed {
		m.vel = m.vel.Unit().Scaled(speed)
	}

	// Movement and Collision
	proposed := m.Location.Add(m.vel.Scaled(timeDelta))
	prevLocation := m.Location

	if m.state == ShadowMobStateTriggering {
		// Triggering ignores walls, leashes, and pathfinding
		m.Location = proposed
	} else {
		// Leash check
		if m.leashRadius > 0 {
			d := proposed.Sub(m.origin)
			if d.Len() > m.leashRadius {
				proposed = m.origin.Add(d.Unit().Scaled(m.leashRadius - epsilon))
				// Reflect velocity or just stop it? Let's just dampen it for now.
				m.vel = m.vel.Scaled(0.5)
				if m.state == ShadowMobStateWandering {
					m.pickNewTarget(s)
				}
			}
		}

		// Tile collision
		if m.canTraverse(s, m.Location, proposed, m.radius) {
			m.Location = proposed
		} else {
			// Slide along walls: try X and Y separately
			proposedX := m.Location.Add(pixel.V(m.vel.X, 0).Scaled(timeDelta))
			canTraverseX := m.canTraverse(s, m.Location, proposedX, m.radius)
			proposedY := m.Location.Add(pixel.V(0, m.vel.Y).Scaled(timeDelta))
			canTraverseY := m.canTraverse(s, m.Location, proposedY, m.radius)

			if canTraverseX {
				m.Location = proposedX
				m.vel.Y = 0
			} else if canTraverseY {
				m.Location = proposedY
				m.vel.X = 0
			} else {
				// Completely blocked
				m.vel = pixel.ZV
				if m.state == ShadowMobStateWandering {
					// If we're blocked during wandering, it might be because our target is unreachable
					// or we're stuck in a corner. Let's pick a new target immediately.
					m.pickNewTarget(s)
				}
			}
		}
	}

	// Footsteps
	if m.footstepDistance > 0 {
		m.footstepCounter += m.Location.Sub(prevLocation).Len()
		dynamicDistance := m.footstepDistance
		if m.footstepGaitScale > 0 {
			// scale distance based on current speed relative to max speed
			// gait = base * (1 + (currentSpeed / maxSpeed - 1) * scale)
			speedRatio := m.vel.Len() / m.maxSpeed
			dynamicDistance *= 1.0 + (speedRatio-1.0)*m.footstepGaitScale
		}
		if m.footstepCounter >= dynamicDistance {
			m.footstepCounter = 0 // reset rather than subtract to handle dynamic gait smoothly
			m.addFootstepFx(s)
		}
	}
}

func (m *ShadowMob) addFootstepFx(s *State) {
	if m.vel.Len() < 0.1 {
		return
	}
	// 8 cardinal directions
	angle := m.vel.Angle()
	// Angle() returns radians in range [-Pi, Pi], starting from positive X axis (Right)
	// We want to map this to 8 sectors, centered on 0, Pi/4, Pi/2, 3Pi/4, Pi, -3Pi/4, -Pi/2, -Pi/4
	// normalize to [0, 2Pi)
	if angle < 0 {
		angle += 2 * math.Pi
	}
	// Shift by half a sector (Pi/8) so that sectors are centered on the cardinal directions
	sector := int(math.Floor((angle+math.Pi/8)/(math.Pi/4))) % 8

	// Sector mapping (assuming 1-based columns in tilesheet, 1=N, 2=NE, 3=E, 4=SE, 5=S, 6=SW, 7=W, 8=NW)
	// Sector 0 is East (Right)
	// Sector 1 is North-East
	// Sector 2 is North
	// Sector 3 is North-West
	// Sector 4 is West
	// Sector 5 is South-West
	// Sector 6 is South
	// Sector 7 is South-East

	// colMap: sector -> column index
	colMap := []int{3, 2, 1, 8, 7, 6, 5, 4}
	col := colMap[sector]

	sprite := atlas.GetTilesheetSprite("adventure/entities/shadow_mob/footsteps", col, m.spriteRow)

	pos := m.Location
	if m.footstepOffset > 0 {
		// Calculate perpendicular offset
		// Velocity vector: m.vel
		// Perpendicular vector: (-m.vel.Y, m.vel.X) or (m.vel.Y, -m.vel.X)
		perp := pixel.V(-m.vel.Y, m.vel.X).Unit()
		offset := m.footstepOffset
		if m.footstepSide {
			offset = -offset
		}
		pos = pos.Add(perp.Scaled(offset))
		m.footstepSide = !m.footstepSide
	}

	s.addForegroundFx(newFadingSpriteFx(pos, sprite, colors.LayerAlpha(m.colorMask, m.stateBasedAlpha()), 3.0))
}

func (m *ShadowMob) stateBasedAlpha() float64 {
	switch m.state {
	case ShadowMobStateWandering:
		return 0.6
	case ShadowMobStateChasing:
		return 0.8
	case ShadowMobStateTriggering:
		return 1.0
	}
	return 1.0
}

// pickNewTarget chooses a new random target within the leash and legal area.
func (m *ShadowMob) pickNewTarget(s *State) {
	if m.leashRadius <= 0 {
		// Just pick a random direction nearby if no leash
		for i := 0; i < 20; i++ {
			angle := rand.Float64() * 2 * math.Pi
			dist := randRange(2, 5)
			candidate := m.Location.Add(pixel.V(math.Cos(angle), math.Sin(angle)).Scaled(dist))
			if s == nil || (m.canTraverse(s, candidate, candidate, m.radius) && m.canTraverse(s, m.Location, candidate, m.radius)) {
				m.targetLocation = candidate
				return
			}
		}
		// If we couldn't find a reachable target, don't just pick something potentially invalid.
		// We'll try again next update or just stay put.
		return
	}

	// Try to find a valid tile within the leash
	for i := 0; i < 20; i++ {
		angle := rand.Float64() * 2 * math.Pi
		dist := rand.Float64() * m.leashRadius
		candidate := m.origin.Add(pixel.V(math.Cos(angle), math.Sin(angle)).Scaled(dist))

		// Check if the candidate point is traversable AND we can reach it from where we are
		if s == nil || (m.canTraverse(s, candidate, candidate, m.radius) && m.canTraverse(s, m.Location, candidate, m.radius)) {
			m.targetLocation = candidate
			return
		}
	}

	// If we can't find anything reachable in the leash, try picking something very close to us that is reachable
	for i := 0; i < 10; i++ {
		angle := rand.Float64() * 2 * math.Pi
		dist := randRange(0.5, 1.5)
		candidate := m.Location.Add(pixel.V(math.Cos(angle), math.Sin(angle)).Scaled(dist))
		// Still check leash
		if m.origin.Sub(candidate).Len() > m.leashRadius {
			continue
		}
		if s == nil || (m.canTraverse(s, candidate, candidate, m.radius) && m.canTraverse(s, m.Location, candidate, m.radius)) {
			m.targetLocation = candidate
			return
		}
	}

	// Fallback to origin if it's reachable, otherwise just stay put
	if s == nil || m.canTraverse(s, m.Location, m.origin, m.radius) {
		m.targetLocation = m.origin
	} else {
		m.targetLocation = m.Location
	}
}

func (m *ShadowMob) Render(batch *pixel.Batch, renderDelta pixel.Vec) {
	if a, exists := m.animations[m.state]; exists {
		mask := colors.Alpha(m.stateBasedAlpha())
		a.Sprite().DrawColorMask(batch, pixel.IM.Moved(renderDelta), mask)
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

// sampleCircleAgainstTiles checks this center movement against all tiles that could intersect the circle.
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
	p, _ := s.entities.GetEntity(s.player)
	isValid, _ := s.entities.isValidTransition(p, MapLocation{X: x, Y: y}, input.Down, true)
	return isValid
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
