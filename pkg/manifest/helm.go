package manifest

import (
	"errors"
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"

	"github.com/valer-cara/mgo/pkg/helm"
)

// This handles yaml manifests, specifically the files containing helm values
// that also have the __mygitops header
//
// TODO: maybe helmbasic isn't so great. also, __mygitops isn't so great either.

type HelmBasic struct {
	Chart *Header `yaml:"__mygitops"`
}

type Header struct {
	// All fields from manifest.HelmRelease: image/repo/...
	helm.HelmRelease `yaml:",inline"`

	// Images map.
	// XXX: needs documentation
	Images map[string]HeaderImage
}

type HeaderImage struct {
	// Either repo/tag combo
	Repository string `yaml:"repository,omitempty"`
	Tag        string `yaml:"tag,omitempty"`

	// Or image
	Image string `yaml:"image,omitempty"`
}

func (h *Header) Validate() error {
	const pre = "`__mygitops` section"

	if h == nil {
		return errors.New(pre + " is empty or missing.")
	}
	if h.HelmRelease == (helm.HelmRelease{}) {
		return errors.New(pre + " is empty or missing..")
	}
	if h.HelmRelease.Chart == "" {
		return errors.New(pre + ": `chart` is empty/missing. Should be 'repo/chartname' (eg: stable/redis)")
	}
	if h.HelmRelease.Version == "" {
		return errors.New(pre + ": `version` is empty/missing. Should be the chart version")
	}
	if h.HelmRelease.Name == "" {
		return errors.New(pre + ": `name` is empty/missing. Should be the release name as seen in `helm ls`")
	}
	if h.HelmRelease.Namespace == "" {
		return errors.New(pre + ": `namespace` is empty/missing. Should be the release namespace")
	}

	return nil
}

func ParseHeader(path string) (*Header, error) {
	var parsed HelmBasic

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(file, &parsed)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("YAML Unmarshal error: %s: %v", path, err))
	}

	return parsed.Chart, nil
}
