package manifest

import (
	"path"
	"path/filepath"
)

type ManifestFileList struct {
	Helm, Raw []string
}

func FindManifests(gitopsRepoRoot, cluster string) (*ManifestFileList, error) {
	globValues := path.Join(gitopsRepoRoot, "/installations/", cluster, "/*-values.yaml")
	valueFiles, err := filepath.Glob(globValues)
	if err != nil {
		return nil, err
	}

	globRaw := path.Join(gitopsRepoRoot, "/installations/", cluster, "/*-raw.yaml")
	rawFiles, err := filepath.Glob(globRaw)
	if err != nil {
		return nil, err
	}

	return &ManifestFileList{
		Helm: valueFiles,
		Raw:  rawFiles,
	}, nil
}
