package v07

type analyzer struct{}

func (a *analyzer) RequiresRunImage() bool {
	return true
}
