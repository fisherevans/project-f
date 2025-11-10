package adventure

type EffectDialogue struct {
	DialogueId string `auto_generate:"true"`
	Text       string
}

func (e *EffectDialogue) CompletionID() string {
	if e.DialogueId == "" {
		return ""
	}
	return "dialogue:" + e.DialogueId
}

func (e *EffectDialogue) Process(source EntityContext, s *State) bool {
	entity, _ := s.entities.GetEntity(source.EntityId())
	cfg, _ := TalkerConfigMetadataKey.Get(entity)
	s.dialogues.Append(NewBasicDialogue(e.Text, e.DialogueId, e.CompletionID(), cfg))
	logEffectInfof(source, e, "dialogue added")
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

func (e *EffectChatter) Process(source EntityContext, s *State) bool {
	s.chatters.Add(newBasicEntityChatter(e.EntityId, e.DurationSeconds, e.Message, e.ChatterId, e.CompletionID()))
	logEffectInfof(source, e, "chatter added")
	return true
}
