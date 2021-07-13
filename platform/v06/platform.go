package v06

import (
	"github.com/buildpacks/lifecycle/api"
	"github.com/buildpacks/lifecycle/cmd"
	"github.com/buildpacks/lifecycle/platform/common"
)

type platform struct {
	api              *api.Version
	analyzer         cmd.Analyzer
	previousPlatform common.Platform
}

func NewPlatform(previousPlatform common.Platform) common.Platform {
	return &platform{
		api:              api.MustParse("0.6"),
		analyzer:         &analyzer{},
		previousPlatform: previousPlatform,
	}
}

func (p *platform) API() string {
	return p.api.String()
}

func (p *platform) Analyzer() cmd.Analyzer {
	return p.analyzer
}

func (p *platform) SupportsAssetPackages() bool {
	return false
}
