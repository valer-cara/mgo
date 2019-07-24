package helm

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"

	yaml "gopkg.in/yaml.v2"
)

// TODO: maybe using helm/rudder as an API instead of os.Exec()ing
// the system helm is cleaner

// Implements HelmService interface based on the actual helm executable command
type HelmCmd struct {
	helmHome    string
	kubeconfig  string
	kubecontext string
	dryRun      bool
	execer      *HelmCmdExec
	options     *HelmCmdOptions

	cache struct {
		repoList []byte
	}
}

type HelmCmdOptions struct {
	Repositories []HelmRepo
	HelmHome     string
	Kubeconfig   string
	Kubecontext  string
	DryRun       bool
}

func NewHelmCmd(options *HelmCmdOptions) *HelmCmd {
	return &HelmCmd{
		// helmHome can be left empty, case in which we create a temp helm home,
		// install diff plugin, add repos
		// if populated it uses preprovisioned diff and repos
		helmHome:    options.HelmHome,
		kubeconfig:  options.Kubeconfig,
		kubecontext: options.Kubecontext,
		dryRun:      options.DryRun,
		options:     options,
	}
}

func (h *HelmCmd) String() string {
	return fmt.Sprintf("HelmCmd(helm_home: %s)", h.helmHome)
}

// TODO: break this down, it's hard to read
func (h *HelmCmd) Init() error {
	// Determine helm environement
	// If helmHome was not specified in NewHelmCmd() as an option, generate a fresh tmp helm home
	if h.helmHome == "" {
		dir, err := ioutil.TempDir("/tmp", "mygitops-helmhome-")
		if err != nil {
			return errors.New("Cannot create helmHome dir. " + err.Error())
		}

		log.Printf("No HELM_HOME specified. Using temp dir '%s' as helm home. Initializing helm...", dir)
		h.helmHome = dir

		defaultEnv, err := h.getDefaultEnv()
		if err != nil {
			return errors.New(fmt.Sprintf("Cannot initialize helm client (%s). Cannot set default env. %v", h, err.Error()))
		}

		h.execer = NewHelmCmdExec(nil, defaultEnv, h.options.Kubecontext)

		// Init helm
		_, err = h.execer.Exec("init", "--client-only")
		if err != nil {
			return errors.New(fmt.Sprintf("helm init: Cannot initialize helm client (%s). %v", h, err.Error()))
		}
	} else {
		log.Debugf("HELM_HOME set to '%s'. Using pre-installed helm...", h.helmHome)
		h.execer = NewHelmCmdExec(nil, os.Environ(), h.options.Kubecontext)
	}

	h.installPlugin("diff", "https://github.com/databus23/helm-diff")
	h.installPlugin("s3", "https://github.com/hypnoglow/helm-s3.git")

	// Add predefined repos

	for _, repo := range h.options.Repositories {
		exists, err := h.repoExists(&repo, false)
		if err != nil {
			return err
		}
		if !exists {
			if err := h.AddRepo(&repo); err != nil {
				return errors.New(fmt.Sprintf("Cannot add repo (%s). %s", h, err.Error()))
			}
		}
	}

	return nil
}

func (h *HelmCmd) Teardown() error {
	//os.RemoveAll(h.helmHome)
	return nil
}

func (h *HelmCmd) AddRepo(repo *HelmRepo) error {
	_, err := h.execer.Exec("repo", "add", repo.Name, repo.Url)
	if err != nil {
		return err
	}

	return nil
}

func (h *HelmCmd) repoExists(repo *HelmRepo, forceRefresh bool) (bool, error) {
	if forceRefresh || len(h.cache.repoList) == 0 {
		repoListOutput, err := h.execer.Exec("repo", "list")
		if err != nil {
			return false, err
		}
		h.cache.repoList = repoListOutput
	}

	re := regexp.MustCompile(`(?m)^` + repo.Name + `\s*` + repo.Url + `\s*$`)
	if re.Match(h.cache.repoList) {
		return true, nil
	}

	return false, nil
}

func (h *HelmCmd) pluginExists(pluginName string) (bool, error) {
	repoListOutput, err := h.execer.Exec("plugin", "list")
	if err != nil {
		return false, err
	}

	re := regexp.MustCompile(`(?m)^` + pluginName + `\s`)
	if re.Match(repoListOutput) {
		return true, nil
	}

	return false, nil

}

type yamlRepoMap struct {
	Repositories []HelmRepo
}

// XXX: Implementation specific
func (h *HelmCmd) ListRepos() ([]HelmRepo, error) {
	var repos yamlRepoMap

	filePath := path.Join(h.helmHome, "repository/repositories.yaml")
	reposFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(reposFile, &repos); err != nil {
		return nil, errors.New(fmt.Sprintf("YAML Unmarshal error: %s: %v", filePath, err))
	}

	return repos.Repositories, nil
}

// Pass a HelmRelease and a set of files where to load values from
func (h *HelmCmd) SyncRelease(release *HelmRelease, valueFiles []string) error {
	cmd := []string{
		"upgrade",
		"--install", release.Name, release.Chart,
		"--version", release.Version,
		"--namespace", release.Namespace,
		//"--recreate-pods", // Nasty - recreates all pods :/
	}

	for _, valueFile := range valueFiles {
		cmd = append(cmd, "--values="+valueFile)
	}

	if h.dryRun {
		cmd = append(cmd, "--dry-run")
	}

	output, err := h.execer.Exec(cmd...)
	if err != nil {
		log.Errorf("Helm exec error: %v", string(output))
		return err
	}

	return nil
}

func (h *HelmCmd) DiffRelease(release *HelmRelease, valueFiles []string) error {
	cmd := []string{
		"diff", "upgrade", release.Name, release.Chart,
		"--allow-unreleased",
		"--version", release.Version,
	}

	for _, valueFile := range valueFiles {
		cmd = append(cmd, "--values="+valueFile)
	}

	_, err := h.execer.Exec(cmd...)
	if err != nil {
		return err
	}

	return nil
}

func (h *HelmCmd) UpdateRepos() error {
	_, err := h.execer.Exec("repo", "update")

	if err != nil {
		return err
	}

	return nil
}

func (h *HelmCmd) SetOutput(w io.Writer) {
	h.execer.writer = w
}

func (h *HelmCmd) installPlugin(name, url string) error {
	exists, err := h.pluginExists(name)
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot determine whether helm %s plugin is installed (%s): %s", name, h, err.Error()))
	}
	if !exists {
		_, err = h.execer.Exec("plugin", "install", url)
		if err != nil {
			return errors.New(fmt.Sprintf("Cannot install %s plugin (%s). %v", name, h, err.Error()))
		}
	}

	return nil
}

func (h *HelmCmd) getDefaultEnv() ([]string, error) {
	// XXX: this is a way to make sure that helm is found by the exec'ed command, even when
	// it's installed in nonstandard locations, as long as it's accessible in the parent shell
	helmBinPath, err := exec.LookPath("helm")
	if err != nil {
		return nil, err
	}

	return []string{
		"HELM_HOME=" + h.helmHome,
		"KUBECONFIG=" + h.options.Kubeconfig,
		"PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:" + filepath.Dir(helmBinPath),
	}, nil

}
