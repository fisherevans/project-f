package textbox

import (
	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/text"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
)

/*
	NewComplexContent parses a template string into a content object.

Each character is parsed into a character object, which is then used to render the text. Each character has it's own
set of render effects, which are applied to the character when it is rendered.

When parsing a template, "commands" can be used to alter the current "character template". When a character is parsed,
it is given the current character template.

Command blocks are in the format: {+cmd1,-cmd2,...}
- Each command is prefixed with a + to indicate "set a template value", or a - to indicate "unset a template value"
- Each command then has a single character, indicated the command type. Some commands take arguments, which are separated by a : character.

The following commands are supported:
- Underline: {+u}Hello{-u}
  - Optional color argument: {+u:blue}Hello{-u}

- Shadow: {+s}Hello{-s}
  - Optional color argument: {+s:blue}Hello{-s}

- Color: {+c:red}Hello{-c}
  - Required color argument, either by name or hex code:
  - {+c:red}Hello{-c}
  - {+c:#ff0000}Hello{-c}

- Typing Weight: {+w:0.5}He{+w:20}llo{-w}
  - Optional weight argument, which is a number indicating the weight of the character. This is used to determine how long it takes to type the character.

- Rumble {+r}Hello{-r}
  - Optional scale argument, which is a number indicating the frequency of the rumble effect.

- Wildcard: {+u}{+c:red}Hello{-*} - only valid with '-' to reset all template values.
*/
func (tb *Instance) NewComplexContent(template string, opts ...ContentOpt) *Content {
	var paragraphs [][]*character
	var currentParagraph []*character

	var parsingCommand bool
	var controlStart int
	cTemplate := &characterTemplate{
		weightScale: 1,
	}

	for cId, c := range []byte(template) {
		if parsingCommand {
			if c != '}' {
				continue
			}
			commandText := template[controlStart+1 : cId]
			cTemplate.parseCommand(commandText)
			parsingCommand = false
		} else {
			if c == '{' {
				parsingCommand = true
				controlStart = cId
				continue
			}
			if c == '\n' {
				paragraphs = append(paragraphs, currentParagraph)
				currentParagraph = nil
				continue
			}
			currentParagraph = append(currentParagraph, cTemplate.newCharacter(c, tb.text))
		}
	}
	paragraphs = append(paragraphs, currentParagraph)

	return tb.newContent(paragraphs, opts...)
}

type characterTemplate struct {
	rumble      *rumbleRenderEffect
	color       *cColor
	shadow      *cShadow
	underline   *cUnderline
	weightScale float64
}

func (t *characterTemplate) newCharacter(char byte, text *text.Text) *character {
	weight := int(1.0 * t.weightScale) // TODO load base weight from cfg?
	var effects []RenderEffect
	if t.rumble != nil {
		effects = append(effects, t.rumble)
	}
	return newCharacter(char, weight, text, cStyle{
		color:     t.color,
		effects:   effects,
		shadow:    t.shadow,
		underline: t.underline,
	})
}

const (
	cmdUnderline = 'u'
	cmdShadow    = 's'
	cmdColor     = 'c'
	cmdWeight    = 'w'
	cmdRumble    = 'r'
	cmdWildcard  = '*'
)

func (t *characterTemplate) parseCommand(commandText string) {
	commandText = strings.TrimSpace(commandText)
	commandText = strings.TrimPrefix(commandText, "{")
	commandText = strings.TrimSuffix(commandText, "}")
	commands := strings.Split(commandText, ",")
	for _, command := range commands {
		command = strings.TrimSpace(command)
		if strings.HasPrefix(command, "+") {
			commandParts := strings.SplitN(strings.TrimPrefix(command, "+"), ":", 2)
			subCmd := commandParts[0]
			param := ""
			if len(commandParts) == 2 {
				param = commandParts[1]
			}
			if len(subCmd) != 1 {
				log.Fatal().Msgf("invalid command %s from %s", subCmd, commandText)
			}
			switch subCmd[0] {
			case cmdUnderline:
				t.underline = &cUnderline{
					color: requireColor(param, pixel.RGB(0.1, 0.1, 0.1)),
				}
			case cmdShadow:
				t.shadow = &cShadow{
					color: requireColor(param, pixel.RGB(0.1, 0.1, 0.1)),
				}
			case cmdColor:
				if len(commandParts) != 2 {
					log.Fatal().Msgf("invalid command %s", command)
				}
				colorString := commandParts[1]
				var color pixel.RGBA
				if strings.HasPrefix(colorString, "#") {
					color = colors.HexColor(commandParts[1])
				} else {
					color = colors.ColorFromName(colorString)
				}
				t.color = &cColor{color}
			case cmdWeight:
				t.weightScale = requireFloat(param, 10.0)
			case cmdRumble:
				t.rumble = newRumble(requireFloat(param, 0.1))
			}
		} else if strings.HasPrefix(command, "-") {
			for _, subCmd := range strings.TrimPrefix(command, "-") {
				switch subCmd {
				case cmdUnderline:
					t.underline = nil
				case cmdShadow:
					t.shadow = nil
				case cmdColor:
					t.color = nil
				case cmdWeight:
					t.weightScale = 1
				case cmdRumble:
					t.rumble = nil
				case cmdWildcard:
					t.underline = nil
					t.shadow = nil
					t.color = nil
					t.weightScale = 1
					t.rumble = nil
				default:
					log.Fatal().Msgf("unknown command %b from %s", subCmd, commandText)
				}
			}
		}
	}
}

func requireColor(param string, defaultColor pixel.RGBA) pixel.RGBA {
	if param == "" {
		return defaultColor
	}
	if strings.HasPrefix(param, "#") {
		return colors.HexColor(param)
	}
	return colors.ColorFromName(param)
}

func requireFloat(param string, defaultValue float64) float64 {
	if param == "" {
		return defaultValue
	}
	f, err := strconv.ParseFloat(param, 64)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to parse float %s", param)
	}
	return f
}

func requireInt(param string, defaultValue int) int {
	if param == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(param)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to parse int %s", param)
	}
	return i
}
