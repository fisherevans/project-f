package adventure

func init() {
	registerStepConverter("dialogue", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		return []Effect{NewDialogueEffect(resolveString(step.Params, tc))}
	})
	registerStepConverter("self_dialogue", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		return []Effect{NewSelfDialogueEffect(resolveString(step.Params, tc))}
	})
	registerStepConverter("chatter", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		m := resolveMap(step.Params, tc)
		return []Effect{NewChatterEffect(mapStr(m, "entity"), mapFloat(m, "duration", 3), mapStr(m, "message"))}
	})
	registerStepConverter("pick_dialogue", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		return convertPickDialogue(step, tc)
	})
	registerStepConverter("pick_self_dialogue", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		return convertPickSelfDialogue(step, tc)
	})
	registerStepConverter("pick_chatter", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		return convertPickChatter(step, tc)
	})
}

type EffectDialogue struct {
	DialogueId string `auto_generate:"true"`
	Text       string
	Style      *DialogueStyle
}

func NewSelfDialogueEffect(text string) *EffectDialogue {
	s := DialogueStyleSelf
	return &EffectDialogue{
		Text:  text,
		Style: &s,
	}
}

func (e *EffectDialogue) CompletionID() string {
	if e.DialogueId == "" {
		return ""
	}
	return "dialogue:" + e.DialogueId
}

func (e *EffectDialogue) Process(source EntityReader, s *State) bool {
	entity, _ := s.entities.GetEntity(source.GetId())
	cfg := TalkerConfigMetadataKey.Get(entity)
	style := DialogueStyleRegular
	if e.Style != nil {
		style = *e.Style
	}
	dialogue := NewDialogue(e.Text, e.DialogueId, e.CompletionID(), cfg, style)
	s.dialogues.Append(dialogue)
	return true
}

type EffectChatter struct {
	ChatterId       string `auto_generate:"true"`
	EntityId        string
	DurationSeconds float64
	Message         string
}

func (e *EffectChatter) CompletionID() string {
	if e.ChatterId == "" {
		return ""
	}
	return "chatter:" + e.ChatterId
}

func (e *EffectChatter) Process(source EntityReader, s *State) bool {
	s.chatters.Add(newBasicEntityChatter(e.EntityId, e.DurationSeconds, e.Message, e.ChatterId, e.CompletionID()))
	return true
}
