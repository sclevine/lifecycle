package v06

type analyzer struct{}

func (a *analyzer) RequiresRunImage() bool {
	return false
}
