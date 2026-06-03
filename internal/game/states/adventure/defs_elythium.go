package adventure

import (
	"fisherevans.com/project/f/internal/util/rng"

	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/tiles"
)

type ElythiumParams struct {
	ResetTime        float64 `yaml:"reset_time"`
	ResetTimeJitter  float64 `yaml:"reset_time_jitter"`
	EnabledByKey     string  `yaml:"enabled_by_key"`
	BroadcastOnBreak string  `yaml:"broadcast_on_break"`
}

func init() {
	newRegistrarBuilder().
		byTile(tiles.Elythium).
		registrar(func(params NewEntityParams, system *EntitySystem) (Entity, EventHandler) {
			entity := system.RegisterEntity(params.EntityId, params.Location)
			system.SetDebugType(params.EntityId, "pickup")
			cfg := ElythiumParams{
				ResetTime:       5,
				ResetTimeJitter: 2,
			}
			params.Properties.LoadStructFromKey("elythium", &cfg)
			isEnabled := func(g StateGlobalsReader) bool {
				if cfg.EnabledByKey == "" {
					return true
				}
				return g.Get(cfg.EnabledByKey).AsBool(false)
			}
			initialMode := "mined"
			if isEnabled(system.state.globals) {
				initialMode = "ready"
			}
			ModeMetadataKey.Set(entity, initialMode)
			AttachModeBasedEntityRenderer(entity).
				WithModeRenderer("ready", NewBasicEntityRenderer(entity).WithAnimations(
					anim.LoadTilesheetAnimation(atlas, "adventure/entities/elythium/crystals", "default"),
					anim.LoadTilesheetAnimation(atlas, "adventure/entities/elythium/crystals_sparkle", "default")).
					WithLights(NewLight(colors.HexString("#f06"), 1.5))).
				WithModeRenderer("mined", NewBasicEntityRenderer(entity).WithAnimations(
					anim.LoadTilesheetAnimation(atlas, "adventure/entities/elythium/crystals_rock", "default")))
			AttachBlockIngressPresence(entity, true, NewImpassableImpedance())
			handler := NewBasicHandler(None{})
			handler.WithOnInteract(func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if !isEnabled(globals) || event.TargetId != thisEntity.GetId() || ModeMetadataKey.Get(thisEntity) == "mined" {
					return nil
				}
				effects := []Effect{
					NewPlaySoundEffect("adventure/elythium_breaking"),
					NewYieldElythiumEffect(3),
					NewMutateModeBasedEntityEffect(thisEntity.GetId()).WithMode("mined"),
				}
				if cfg.BroadcastOnBreak != "" {
					effects = append(effects, NewSendBroadcastEffect("intro.training.7.elythium", nil))
				}
				if cfg.ResetTime+cfg.ResetTimeJitter > 0 {
					effects = append(effects,
						NewTimerEffect(cfg.ResetTime+rng.Float64()*cfg.ResetTimeJitter).WithTimerId(thisEntity.GetId()+".reset"))
				}
				return NewOutput().WithEffects(effects...)
			})
			handler.WithTimerComplete(func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventTimerComplete) *HandlerOutput {
				if !isEnabled(globals) || event.CreatedBy != thisEntity.GetId() || event.TimerId != thisEntity.GetId()+".reset" {
					return nil
				}
				return NewOutput().WithEffects(NewMutateModeBasedEntityEffect(thisEntity.GetId()).WithMode("ready"))
			})
			handler.WithGlobalVariableUpdated(func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventGlobalVariableUpdated) *HandlerOutput {
				if event.Key != cfg.EnabledByKey {
					return nil
				}
				if event.NewValue.AsBool(false) && ModeMetadataKey.Get(thisEntity) == "mined" {
					return NewOutput().WithEffects(NewMutateModeBasedEntityEffect(thisEntity.GetId()).WithMode("ready"))
				}
				return nil
			})
			return entity, handler.CreateHandler()
		})
}
