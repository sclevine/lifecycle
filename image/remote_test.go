package image_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/buildpack/lifecycle/fs"
	"github.com/buildpack/lifecycle/image"
	h "github.com/buildpack/lifecycle/testhelpers"

	dockerClient "github.com/docker/docker/client"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var registryPort string

func TestRemote(t *testing.T) {
	t.Parallel()
	rand.Seed(time.Now().UTC().UnixNano())

	registryPort = h.RunRegistry(t, false)
	defer h.StopRegistry(t)

	spec.Run(t, "remote", testRemote, spec.Sequential(), spec.Report(report.Terminal{}))
}

func testRemote(t *testing.T, when spec.G, it spec.S) {
	var factory image.Factory
	var buf bytes.Buffer
	var repoName string
	var dockerCli *dockerClient.Client

	it.Before(func() {
		var err error
		dockerCli = h.DockerCli(t)
		h.AssertNil(t, err)
		factory = image.Factory{
			Docker: dockerCli,
			Log:    log.New(&buf, "", log.LstdFlags),
			Stdout: &buf,
			FS:     &fs.FS{},
		}
		repoName = "localhost:" + registryPort + "/pack-image-test-" + h.RandString(10)
	})

	when("#label", func() {
		when("image exists", func() {
			var img image.Image
			it.Before(func() {
				h.CreateImageOnRemote(t, dockerCli, repoName, fmt.Sprintf(`
					FROM scratch
					LABEL repo_name_for_randomisation=%s
					LABEL mykey=myvalue other=data
				`, repoName))

				var err error
				img, err = factory.NewRemote(repoName)
				h.AssertNil(t, err)
			})

			it("returns the label value", func() {
				label, err := img.Label("mykey")
				h.AssertNil(t, err)
				h.AssertEq(t, label, "myvalue")
			})

			it("returns an empty string for a missing label", func() {
				label, err := img.Label("missing-label")
				h.AssertNil(t, err)
				h.AssertEq(t, label, "")
			})
		})

		when("image NOT exists", func() {
			it("returns an error", func() {
				img, err := factory.NewRemote(repoName)
				h.AssertNil(t, err)

				_, err = img.Label("mykey")
				h.AssertError(t, err, fmt.Sprintf("failed to get label, image '%s' does not exist", repoName))
			})
		})
	})

	when("#Name", func() {
		it("always returns the original name", func() {
			img, err := factory.NewRemote(repoName)
			h.AssertNil(t, err)
			h.AssertEq(t, img.Name(), repoName)
		})
	})

	when("#Digest", func() {
		it("returns the image digest", func() {
			//busybox:1.29 has digest sha256:915f390a8912e16d4beb8689720a17348f3f6d1a7b659697df850ab625ea29d5
			img, err := factory.NewRemote("busybox:1.29")
			h.AssertNil(t, err)
			digest, err := img.Digest()
			h.AssertNil(t, err)
			h.AssertEq(t, digest, "sha256:915f390a8912e16d4beb8689720a17348f3f6d1a7b659697df850ab625ea29d5")
		})
	})

	when("#SetLabel", func() {
		var img image.Image
		when("image exists", func() {
			it.Before(func() {
				h.CreateImageOnRemote(t, dockerCli, repoName, fmt.Sprintf(`
					FROM scratch
					LABEL repo_name_for_randomisation=%s
					LABEL mykey=myvalue other=data
				`, repoName))

				var err error
				img, err = factory.NewRemote(repoName)
				h.AssertNil(t, err)
			})

			it("sets label on img object", func() {
				h.AssertNil(t, img.SetLabel("mykey", "new-val"))
				label, err := img.Label("mykey")
				h.AssertNil(t, err)
				h.AssertEq(t, label, "new-val")
			})

			it("saves label", func() {
				h.AssertNil(t, img.SetLabel("mykey", "new-val"))
				_, err := img.Save()
				h.AssertNil(t, err)

				// After Pull
				label := remoteLabel(t, dockerCli, repoName, "mykey")
				h.AssertEq(t, "new-val", label)
			})
		})
	})

	when("#SetEnv", func() {
		var (
			img image.Image
		)
		it.Before(func() {
			var err error
			h.CreateImageOnRemote(t, dockerCli, repoName, fmt.Sprintf(`
					FROM scratch
					LABEL repo_name_for_randomisation=%s
					LABEL some-key=some-value
				`, repoName))
			img, err = factory.NewRemote(repoName)
			h.AssertNil(t, err)
		})

		it("sets the environment", func() {
			err := img.SetEnv("ENV_KEY", "ENV_VAL")
			h.AssertNil(t, err)

			_, err = img.Save()
			h.AssertNil(t, err)

			h.AssertNil(t, h.PullImage(dockerCli, repoName))
			defer h.DockerRmi(dockerCli, repoName)

			inspect, _, err := dockerCli.ImageInspectWithRaw(context.TODO(), repoName)
			h.AssertNil(t, err)

			h.AssertContains(t, inspect.Config.Env, "ENV_KEY=ENV_VAL")
		})
	})

	when("#Rebase", func() {
		when("image exists", func() {
			var oldBase, oldTopLayer, newBase string
			var oldBaseLayers, newBaseLayers, repoTopLayers []string
			it.Before(func() {
				var wg sync.WaitGroup
				wg.Add(1)

				newBase = "localhost:" + registryPort + "/pack-newbase-test-" + h.RandString(10)
				go func() {
					defer wg.Done()
					h.CreateImageOnRemote(t, dockerCli, newBase, fmt.Sprintf(`
						FROM busybox
						LABEL repo_name_for_randomisation=%s
						RUN echo new-base > base.txt
						RUN echo text-new-base > otherfile.txt
					`, repoName))
					newBaseLayers = manifestLayers(t, newBase)
				}()

				oldBase = "localhost:" + registryPort + "/pack-oldbase-test-" + h.RandString(10)
				oldTopLayer = h.CreateImageOnRemote(t, dockerCli, oldBase, fmt.Sprintf(`
					FROM busybox
					LABEL repo_name_for_randomisation=%s
					RUN echo old-base > base.txt
					RUN echo text-old-base > otherfile.txt
				`, oldBase))
				oldBaseLayers = manifestLayers(t, oldBase)

				h.CreateImageOnRemote(t, dockerCli, repoName, fmt.Sprintf(`
					FROM %s
					LABEL repo_name_for_randomisation=%s
					RUN echo text-from-image-1 > myimage.txt
					RUN echo text-from-image-2 > myimage2.txt
				`, oldBase, repoName))
				repoTopLayers = manifestLayers(t, repoName)[len(oldBaseLayers):]

				wg.Wait()
			})

			it.After(func() {
				h.AssertNil(t, h.DockerRmi(dockerCli, oldBase))
			})

			it("switches the base", func() {
				// Before
				h.AssertEq(t,
					manifestLayers(t, repoName),
					append(oldBaseLayers, repoTopLayers...),
				)

				// Run rebase
				img, err := factory.NewRemote(repoName)
				h.AssertNil(t, err)
				newBaseImg, err := factory.NewRemote(newBase)
				h.AssertNil(t, err)
				err = img.Rebase(oldTopLayer, newBaseImg)
				h.AssertNil(t, err)
				_, err = img.Save()
				h.AssertNil(t, err)

				// After
				h.AssertEq(t,
					manifestLayers(t, repoName),
					append(newBaseLayers, repoTopLayers...),
				)
			})
		})
	})

	when("#TopLayer", func() {
		when("image exists", func() {
			it("returns the digest for the top layer (useful for rebasing)", func() {
				expectedTopLayer := h.CreateImageOnRemote(t, dockerCli, repoName, fmt.Sprintf(`
					FROM busybox
					LABEL repo_name_for_randomisation=%s
					RUN echo old-base > base.txt
					RUN echo text-old-base > otherfile.txt
				`, repoName))

				img, err := factory.NewRemote(repoName)
				h.AssertNil(t, err)

				actualTopLayer, err := img.TopLayer()
				h.AssertNil(t, err)

				h.AssertEq(t, actualTopLayer, expectedTopLayer)
			})
		})
	})

	when("#AddLayer", func() {
		var (
			tarPath string
			img     image.Image
		)
		it.Before(func() {
			h.CreateImageOnRemote(t, dockerCli, repoName, fmt.Sprintf(`
					FROM busybox
					LABEL repo_name_for_randomisation=%s
					RUN echo -n old-layer > old-layer.txt
				`, repoName))
			tr, err := (&fs.FS{}).CreateSingleFileTar("/new-layer.txt", "new-layer")
			h.AssertNil(t, err)
			tarFile, err := ioutil.TempFile("", "add-layer-test")
			h.AssertNil(t, err)
			defer tarFile.Close()
			_, err = io.Copy(tarFile, tr)
			h.AssertNil(t, err)
			tarPath = tarFile.Name()

			img, err = factory.NewRemote(repoName)
			h.AssertNil(t, err)
		})

		it.After(func() {
			h.AssertNil(t, os.Remove(tarPath))
			h.AssertNil(t, h.DockerRmi(dockerCli, repoName))
		})

		it("appends a layer", func() {
			err := img.AddLayer(tarPath)
			h.AssertNil(t, err)

			_, err = img.Save()
			h.AssertNil(t, err)

			// After Pull
			h.AssertNil(t, h.PullImage(dockerCli, repoName))

			output, err := h.CopySingleFileFromImage(dockerCli, repoName, "old-layer.txt")
			h.AssertNil(t, err)
			h.AssertEq(t, output, "old-layer")

			output, err = h.CopySingleFileFromImage(dockerCli, repoName, "new-layer.txt")
			h.AssertNil(t, err)
			h.AssertEq(t, output, "new-layer")
		})
	})

	when("#ReuseLayer", func() {
		when("previous image", func() {
			var (
				layer2SHA string
				img       image.Image
			)

			it.Before(func() {
				var err error

				h.CreateImageOnRemote(t, dockerCli, repoName, fmt.Sprintf(`
					FROM busybox
					LABEL repo_name_for_randomisation=%s
					RUN echo -n old-layer-1 > layer-1.txt
					RUN echo -n old-layer-2 > layer-2.txt
				`, repoName))

				h.AssertNil(t, h.PullImage(dockerCli, repoName))
				defer func() {
					h.AssertNil(t, h.DockerRmi(dockerCli, repoName))
				}()
				inspect, _, err := dockerCli.ImageInspectWithRaw(context.TODO(), repoName)
				h.AssertNil(t, err)

				layer2SHA = inspect.RootFS.Layers[2]

				img, err = factory.NewRemote("busybox")
				h.AssertNil(t, err)
			})

			it("reuses a layer", func() {
				img.Rename(repoName)
				err := img.ReuseLayer(layer2SHA)
				h.AssertNil(t, err)

				_, err = img.Save()
				h.AssertNil(t, err)

				h.AssertNil(t, h.PullImage(dockerCli, repoName))
				defer h.DockerRmi(dockerCli, repoName)
				output, err := h.CopySingleFileFromImage(dockerCli, repoName, "layer-2.txt")
				h.AssertNil(t, err)
				h.AssertEq(t, output, "old-layer-2")

				// Confirm layer-1.txt does not exist
				_, err = h.CopySingleFileFromImage(dockerCli, repoName, "layer-1.txt")
				h.AssertMatch(t, err.Error(), regexp.MustCompile(`Error: No such container:path: .*:layer-1.txt`))
			})

			it("returns error on nonexistent layer", func() {
				img.Rename(repoName)
				err := img.ReuseLayer("some-bad-sha")

				h.AssertError(t, err, "previous image did not have layer with sha 'some-bad-sha'")
			})
		})

		it("returns errors on nonexistent prev image", func() {
			img, err := factory.NewRemote("busybox")
			h.AssertNil(t, err)
			img.Rename("some-bad-repo-name")

			err = img.ReuseLayer("some-bad-sha")

			h.AssertError(t, err, "failed to get layers for previous image with repo name 'some-bad-repo-name'")
		})
	})

	when("#Save", func() {
		when("image exists", func() {
			it("returns the image digest", func() {
				h.CreateImageOnRemote(t, dockerCli, repoName, fmt.Sprintf(`
					FROM busybox
					LABEL repo_name_for_randomisation=%s
					LABEL mykey=oldValue
				`, repoName))

				img, err := factory.NewRemote(repoName)
				h.AssertNil(t, err)

				err = img.SetLabel("mykey", "newValue")
				h.AssertNil(t, err)

				imgDigest, err := img.Save()
				h.AssertNil(t, err)

				// After Pull
				label := remoteLabel(t, dockerCli, repoName+"@"+imgDigest, "mykey")
				h.AssertEq(t, "newValue", label)

			})
		})
	})

	when("#Found", func() {
		when("it exists", func() {
			it.Before(func() {
				h.CreateImageOnRemote(t, dockerCli, repoName, fmt.Sprintf(`
					FROM scratch
					LABEL repo_name_for_randomisation=%s
				`, repoName))
			})

			it("returns true, nil", func() {
				image, err := factory.NewRemote(repoName)
				h.AssertNil(t, err)
				exists, err := image.Found()

				h.AssertNil(t, err)
				h.AssertEq(t, exists, true)
			})
		})

		when("it does not exist", func() {
			it("returns false, nil", func() {
				image, err := factory.NewRemote(repoName)
				h.AssertNil(t, err)
				exists, err := image.Found()

				h.AssertNil(t, err)
				h.AssertEq(t, exists, false)
			})
		})
	})
}

func manifestLayers(t *testing.T, repoName string) []string {
	t.Helper()

	arr := strings.SplitN(repoName, "/", 2)
	if len(arr) != 2 {
		t.Fatalf("expected repoName to have 1 slash (remote test registry): '%s'", repoName)
	}

	url := "http://" + arr[0] + "/v2/" + arr[1] + "/manifests/latest"
	req, err := http.NewRequest("GET", url, nil)
	h.AssertNil(t, err)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	resp, err := http.DefaultClient.Do(req)
	h.AssertNil(t, err)
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		t.Fatalf("HTTP Status was bad: %s => %d", url, resp.StatusCode)
	}

	var manifest struct {
		Layers []struct {
			Digest string `json:"digest"`
		} `json:"layers"`
	}
	json.NewDecoder(resp.Body).Decode(&manifest)
	h.AssertNil(t, err)

	outSlice := make([]string, 0, len(manifest.Layers))
	for _, layer := range manifest.Layers {
		outSlice = append(outSlice, layer.Digest)
	}

	return outSlice
}

func remoteLabel(t *testing.T, dockerCli *dockerClient.Client, repoName, label string) string {
	t.Helper()

	h.AssertNil(t, h.PullImage(dockerCli, repoName))
	defer func() { h.AssertNil(t, h.DockerRmi(dockerCli, repoName)) }()
	inspect, _, err := dockerCli.ImageInspectWithRaw(context.TODO(), repoName)
	h.AssertNil(t, err)
	return inspect.Config.Labels[label]
}