package util

type Visibility struct {
	isVisible          bool
	visibleAmount      float64
	transitionDuration float64
}

func NewVisibility(visible bool, transitionDuration float64, skipTransition bool) *Visibility {
	v := &Visibility{
		transitionDuration: transitionDuration,
	}
	v.SetVisibleImmediately(!visible)
	if skipTransition {
		v.SetVisibleImmediately(visible)
	} else {
		v.SetVisible(visible)
	}
	return v
}

func (v *Visibility) Update(timeDelta float64) {
	if !v.isVisible {
		timeDelta *= -1
	}
	v.visibleAmount = max(min(1.0, v.visibleAmount+timeDelta/v.transitionDuration), 0)
}

func (v *Visibility) IsVisible() bool {
	return v.isVisible
}

func (v *Visibility) SetVisible(isVisible bool) {
	v.isVisible = isVisible
}

func (v *Visibility) SetVisibleImmediately(isVisible bool) {
	v.isVisible = isVisible
	if v.isVisible {
		v.visibleAmount = 1.0
	} else {
		v.visibleAmount = 0.0
	}
}

func (v *Visibility) GetVisibleAmount() float64 {
	return v.visibleAmount
}

func (v *Visibility) GetInvisibleAmount() float64 {
	return 1.0 - v.visibleAmount
}
