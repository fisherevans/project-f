package adventure

import (
	"sort"
	"strconv"
	"strings"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/commands"
	"fisherevans.com/project/f/internal/game/rpg"
	"github.com/spf13/cobra"
)

func (s *State) initCommands() {
	s.commandRoot = commands.New("", "Adventure state commands")
	s.commandRoot.Register(
		s.newTeleportCommand(),
		s.newDetailCommand(),
		s.newGlobalCommand(),
		s.newMapCommand(),
		s.newBroadcastCommand(),
		s.newTooltipCommand(),
		s.newElythiumCommand(),
	)
}

func (s *State) HandleConsoleInput(cmd string) bool {
	return s.commandRoot.HandleInput(cmd)
}

func (s *State) newTeleportCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "tp [entityId]",
		Short:   "Teleport to an entity",
		Aliases: []string{"teleport"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			target := args[0]
			if _, ok := s.entities.GetEntity(target); ok {
				s.ExecuteSystemEffectsInOrder(NewTeleportPlayerEffect().
					WithToEntityId(target))
				game.Console().WriteLines("teleporting to entity")
				return
			}
			if ref, ok := s.teleports[TeleportReference("teleport:"+target)]; ok {
				s.ExecuteSystemEffectsInOrder(NewTeleportPlayerEffect().
					WithToLocation(ref.Location))
				game.Console().WriteLines("teleporting to reference")
			}
			game.Console().WriteLines("invalid target")
			return
		},
	}
}

func (s *State) newDetailCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "detail",
		Short: "Show details about the location in front of the player",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			p := s.Globals().Player()
			loc := p.GetLocation().Moved(p.GetFacingDirection())
			game.Console().Writef("location: %s", loc)
			for _, eId := range s.entities.occupations.OccupyingEntityList(loc) {
				game.Console().Writef("entity: %s", eId)
				if p, ok := s.entities.presences[eId]; ok {
					game.Console().Writef("- presence: %#v", p)
				}
				if r, ok := s.entities.renderers[eId]; ok {
					game.Console().Writef("- renderer: %#v", r)
				}
				for _, b := range s.entities.behaviors[eId] {
					game.Console().Writef("- behavior: %#v", b)
				}
				if db, ok := s.entities.disabledBehaviors[eId]; ok {
					for by := range db {
						game.Console().Writef("- behavior disabled by: %#v", by)
					}
				}
				for _, s := range s.entities.soundProviders[eId] {
					game.Console().Writef("- sound: %#v", s)
				}
				if db, ok := s.entities.disabledSounds[eId]; ok {
					for by := range db {
						game.Console().Writef("- sound disabled by: %#v", by)
					}
				}
				if md, ok := s.entities.metadata[eId]; ok {
					game.Console().Writef("- metadata: %#v", md)
				}
			}
			for _, z := range s.zones.ZonesAt(loc) {
				game.Console().Writef("zone: %s", z)
			}
		},
	}
}

func (s *State) newMapCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "map [name]",
		Short: "Load another map",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			waypoint := ""
			if len(args) >= 2 {
				waypoint = args[1]
			}
			s.ExecuteSystemEffects(NewLoadMapEffect(args[0]).WithWaypoint(waypoint))
			game.Console().Writef("loading map: %s", args[0])
		},
	}
}

func (s *State) newBroadcastCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "broadcast [id]",
		Short: "Broadcast an event ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			s.ExecuteSystemEffects(NewSendBroadcastEffect(args[0], nil))
			game.Console().Writef("broadcasted: %s", args[0])
		},
	}
}

func (s *State) newTooltipCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "tooltip [message]",
		Short: "Trigger a tooltip",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			s.ExecuteSystemEffects(NewPushTooltipEffect(args[0]))
			game.Console().Writef("sent tooltip: %s", args[0])
		},
	}
}

func (s *State) newElythiumCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "elythium [amount]",
		Short: "Set the current elythium count",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			amnt, err := strconv.Atoi(args[0])
			if err != nil {
				game.Console().Writef("invalid amount: %v", err)
				return
			}
			s.globals.Set(rpg.GlobalKeyElythium, amnt)
			game.Console().Writef("elythium set to %d", amnt)
		},
	}
}

func (s *State) newGlobalCommand() *cobra.Command {
	resolveKey := func(key string) string {
		if strings.HasPrefix(key, "#") {
			idx, err := strconv.Atoi(key[1:])
			if err == nil {
				resolved := game.Console().GetLastGlobalKey(idx)
				if resolved != "" {
					return resolved
				}
			}
		}
		return key
	}

	globalCmd := &cobra.Command{
		Use:   "global",
		Short: "Manage global variables",
	}

	getCmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Get a global variable value",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := resolveKey(args[0])
			val := s.globals.Get(id)
			if !val.Exists() {
				game.Console().Writef("global %s not found", id)
				return
			}
			game.Console().Writef("%s = %v", id, val.Value())
		},
	}

	setCmd := &cobra.Command{
		Use:   "set [id] [type] [value]",
		Short: "Set a global variable value (types: string, int, float, bool)",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			id := resolveKey(args[0])
			typ := args[1]
			valStr := args[2]
			var val any
			var err error
			switch typ {
			case "string":
				val = valStr
			case "int":
				var i int
				i, err = strconv.Atoi(valStr)
				val = i
			case "float":
				val, err = strconv.ParseFloat(valStr, 64)
			case "bool":
				val, err = strconv.ParseBool(valStr)
			default:
				game.Console().Writef("unknown type: %s", typ)
				return
			}
			if err != nil {
				game.Console().Writef("invalid value for type %s: %v", typ, err)
				return
			}
			s.globals.Set(id, val)
			game.Console().Writef("%s set to %v (%s)", id, val, typ)
		},
	}

	listCmd := &cobra.Command{
		Use:   "list [prefix]",
		Short: "List global variables with an optional prefix",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			prefix := ""
			if len(args) > 0 {
				prefix = args[0]
			}
			keys := s.globals.KeysWithPrefix(prefix)
			sort.Strings(keys)
			if len(keys) == 0 {
				game.Console().WriteLines("no globals found")
				return
			}
			game.Console().SetLastGlobalKeys(keys)
			for i, k := range keys {
				val := s.globals.Get(k)
				game.Console().Writef("%d. %s = %v", i+1, k, val.Value())
			}
		},
	}

	globalCmd.AddCommand(getCmd, setCmd, listCmd)
	return globalCmd
}
