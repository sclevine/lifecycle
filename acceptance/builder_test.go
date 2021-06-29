package acceptance

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	"github.com/buildpacks/lifecycle/api"
	h "github.com/buildpacks/lifecycle/testhelpers"
)

var (
	builderDockerContext = filepath.Join("testdata", "builder")
	builderBinaryDir     = filepath.Join("testdata", "builder", "container", "cnb", "lifecycle")
	builderImage         = "lifecycle/acceptance/builder"
	builderUserID        = "1234"
)

func TestBuilder(t *testing.T) {
	h.SkipIf(t, runtime.GOOS == "windows", "builder acceptance tests are not yet supported on Windows")

	rand.Seed(time.Now().UTC().UnixNano())

	h.MakeAndCopyLifecycle(t, "linux", builderBinaryDir)
	h.DockerBuild(t,
		builderImage,
		builderDockerContext,
		h.WithArgs("--build-arg", fmt.Sprintf("cnb_platform_api=%s", api.Platform.Latest())),
	)
	//defer h.DockerImageRemove(t, builderImage)

	spec.Run(t, "acceptance-builder", testBuilder, spec.Parallel(), spec.Report(report.Terminal{}))
}

func testBuilder(t *testing.T, when spec.G, it spec.S) {
}
