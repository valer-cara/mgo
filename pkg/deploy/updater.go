package deploy

import (
	"errors"
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	//log "github.com/sirupsen/logrus"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/valer-cara/mgo/pkg/manifest"
)

// Updater groups operations tied to a specific gitops repo structure.
// This is where the opinionated stuff is restrained.
// If you'd like to structure your project differently, hack here!
type Updater interface {
	Update(string, *DeployOptions) error
}

// FakeUpdater
type FakeUpdater struct{}

func (u *FakeUpdater) Update(gitopsRepo string, deployOptions *DeployOptions) error {
	return nil
}

// The updater used with mygitops' conventions
type MyUpdater struct{}

func (u *MyUpdater) Update(gitopsRepo string, deployOptions *DeployOptions) error {
	globPath := path.Join(gitopsRepo, "/installations/", deployOptions.Cluster, "/*-values.yaml")

	valueFiles, err := filepath.Glob(globPath)
	if err != nil {
		return err
	}

	if len(valueFiles) == 0 {
		return errors.New(fmt.Sprintf(
			"No files found for for cluster '%s' in gitops repo '%s'. Searched for '%s'. Aborting...",
			deployOptions.Cluster,
			gitopsRepo,
			globPath,
		))
	}

	didPatchAnything := false
	for _, valueFile := range valueFiles {
		didPatch, err := updateFile(valueFile, deployOptions)
		if err != nil {
			return err
		}
		if didPatch {
			didPatchAnything = true
		}
	}

	if !didPatchAnything {
		// TODO: nicer errors/hints
		return errors.New(fmt.Sprintf(
			"No deployments were patched. Is the repo '%s' referenced in any of the manifests? Aborting...",
			deployOptions.TriggerRepo,
		))
	}

	return nil
}

// Update the values file
//
// Until this pull request is merged (yaml.v3) we're hacking yaml edits so we don't affect comments
// https://github.com/go-yaml/yaml/pull/219
// https://github.com/go-yaml/yaml/issues/132
//
// The hacks means matching the line interval for the whole `__mygitops` block and
// only editing that piece, writing the rest unchanged
//
// Indicates whether it actually patched or not via the bool return
// Returns: didPatch bool, error
func updateFile(valueFilePath string, deployOptions *DeployOptions) (bool, error) {
	var (
		offt         []int
		chartSection *manifest.HelmBasic
	)
	file, err := ioutil.ReadFile(valueFilePath)
	if err != nil {
		return false, err
	}

	offt, chartSection, err = getChartSection(file)
	if err != nil {
		return false, errors.New(fmt.Sprintf("File %s: %v", valueFilePath, err))
	} else if chartSection == nil {
		return false, nil
	}

	if len(chartSection.Chart.Images) == 0 {
		//return errors.New(fmt.Sprintf("File <%s> does not define __mygitops.images.*", valueFilePath))
		return false, nil
	}

	var proceedPatching = false

	for crtTargetRepo, _ := range chartSection.Chart.Images {
		if strings.Compare(deployOptions.TriggerRepo, crtTargetRepo) == 0 {
			proceedPatching = true
			//log.Printf("Will update: %s -> %s:%s", crtTargetRepo, conf.Image, conf.Tag)
			chartSection.Chart.Images[crtTargetRepo] = manifest.HeaderImage{
				Repository: deployOptions.Image.Repository,
				Tag:        deployOptions.Image.Tag,
			}
			//log.Printf("Updated: %s -> %s:%s", crtTargetRepo, chartSection.Chart.Images[crtTargetRepo].Image, chartSection.Chart.Images[crtTargetRepo].Tag)
		}
	}

	if !proceedPatching {
		return false, nil
	}

	return true, patchChartSection(valueFilePath, file, chartSection, offt)
}

// Super hacky: until we get yaml.v3 which should be better with a unmarshal->marshal loop // returns offsets, parsed chart, error
func getChartSection(file []byte) ([]int, *manifest.HelmBasic, error) {
	var (
		start, end    int
		parsedSection manifest.HelmBasic
	)

	// Find section start
	re := regexp.MustCompile(`(?m)^__mygitops:`)
	idxStart := re.FindAllIndex(file, -1)

	if len(idxStart) == 0 {
		return nil, nil, nil
	} else if len(idxStart) > 1 {
		return nil, nil, errors.New(fmt.Sprintf("File contains multiple __mygitops keys"))
	}

	start = idxStart[0][0]

	// Find section end
	re = regexp.MustCompile(`\n\r?\S`) // newline followed by non space immediately
	idxEnd := re.FindIndex(file[start:])

	if idxEnd == nil {
		end = len(file) - 1
	} else {
		end = start + idxEnd[0] + 1
	}

	err := yaml.Unmarshal(file[start:end], &parsedSection)
	if err != nil {
		return nil, nil, err
	}

	return []int{start, end}, &parsedSection, nil
}

func patchChartSection(path string, originalContent []byte, chartSection *manifest.HelmBasic, patchOffests []int) error {
	var (
		snipStart = patchOffests[0]
		snipEnd   = patchOffests[1]
	)

	newChartSection, err := yaml.Marshal(*chartSection)
	if err != nil {
		return err
	}

	newChartSection, err = restoreAnchors(originalContent[snipStart:snipEnd], newChartSection)
	if err != nil {
		return err
	}

	out := []byte{}
	out = append(out, originalContent[0:patchOffests[0]]...)
	out = append(out, newChartSection...)
	out = append(out, originalContent[patchOffests[1]:]...)

	err = ioutil.WriteFile(path, out, 0644)
	if err != nil {
		return err
	}

	return nil
}

func restoreAnchors(oldSection, newSection []byte) ([]byte, error) {
	//re := regexp.MustCompile(`(?m)^\s+(\S): \&([\S])`)
	reOrig := regexp.MustCompile(`(?m)^\s+([\S]+):\s*(&[\S]+)`)
	res := reOrig.FindAllSubmatch(oldSection, -1)

	for _, match := range res {
		searchFor := `(?m)^\s+` + string(match[1]) + `:`
		replaceWith := string(match[0])
		reNew := regexp.MustCompile(searchFor)
		newSection = reNew.ReplaceAll(newSection, []byte(replaceWith))
	}

	return newSection, nil
}
