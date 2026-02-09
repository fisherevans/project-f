package adventure

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
