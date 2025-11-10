package adventure

import "github.com/rs/zerolog/log"

var (
	TalkerConfigMetadataKey = MetadataKey[TalkerConfig]{
		name: "talker_config",
		defaultFactory: func() TalkerConfig {
			return TalkerConfig{
				VoicePitch: 0.0, // Normal pitch
				Energy:     0.0, // Normal energy
			}
		},
	}
)

type TalkerConfig struct {
	Preset     string  `yaml:"preset"`
	VoicePitch float64 `yaml:"voice_pitch"` // -1.0 (deep bass) to 1.0 (high soprano)
	Energy     float64 `yaml:"energy"`      // -1.0 (slow/quiet) to 1.0 (fast/loud)
}

// ToParams converts the intuitive config to internal pitch/gain/speed adjustments
func (cfg TalkerConfig) ToParams() (speedAdjust, gainAdjust, throttleScale float64) {
	switch cfg.Preset {
	case "", "normal":
	case "slow":
		cfg.VoicePitch = 0.7
		cfg.Energy = 0.8
	case "kid":
		cfg.VoicePitch = -0.5
		cfg.Energy = -0.3
	default:
		log.Warn().Interface("config", cfg).Msg("unknown talker config preset")
	}
	speedAdjust = cfg.VoicePitch*0.4 + cfg.Energy*0.05
	gainAdjust = -cfg.Energy*0.5 + cfg.VoicePitch*0.75
	throttleScale = 1.0 + cfg.Energy/4.0 + cfg.VoicePitch/4.0
	return
}
