package common

import (
	"github.com/buildpacks/lifecycle/cmd"
)

type Platform interface {
	API() string
	CodeFor(errType cmd.LifecycleExitError) int
	SupportsAssetPackages() bool
	Analyzer() cmd.Analyzer // TODO: see about removing this dependency
}

type Analyzer interface {
	RequiresRunImage() bool
}
