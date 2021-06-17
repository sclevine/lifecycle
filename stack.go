package lifecycle

import (
	"strings"

	"github.com/buildpacks/lifecycle/platform"

	"github.com/pkg/errors"

	"github.com/buildpacks/lifecycle/buildpack"
)

type StackValidator struct{}

func (v *StackValidator) ValidateMixins(descriptor buildpack.Descriptor, analyzed platform.AnalyzedMetadata) error {
	if len(descriptor.Stacks) == 0 {
		return nil // nothing to validate
	}

	var currentStack buildpack.Stack
	for _, stack := range descriptor.Stacks {
		if stack.ID == analyzed.BuildImageStackID() {
			currentStack = stack
			break
		}
	}
	if currentStack.ID == "" {
		return errors.New("failed to find current stack") // shouldn't get here if analyzer validated the stack id
	}

	for _, mixin := range currentStack.Mixins {
		if !satisfied(mixin, analyzed) {
			return errors.Errorf("buildpack %s missing required mixin %s", descriptor.String(), mixin)
		}
	}

	return nil
}

func satisfied(mixin string, analyzed platform.AnalyzedMetadata) bool {
	if strings.HasPrefix(mixin, "build") {
		return hasMixin(analyzed.BuildImageMixins(), mixin)
	}
	if strings.HasPrefix(mixin, "run") {
		return hasMixin(analyzed.RunImageMixins(), mixin)
	}
	return hasMixin(analyzed.BuildImageMixins(), mixin) && hasMixin(analyzed.RunImageMixins(), mixin)
}

func hasMixin(installedMixins []string, required string) bool {
	for _, installed := range installedMixins {
		if removeStagePrefix(installed) == removeStagePrefix(required) {
			return true
		}
	}
	return false
}
