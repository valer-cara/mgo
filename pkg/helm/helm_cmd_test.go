package helm

import (
	"testing"
)

func TestHelmInitWithRepos(t *testing.T) {
	presetRepos := []HelmRepo{
		{Name: "foobar", Url: "https://kubernetes-charts.storage.googleapis.com"},
		{Name: "bazbar", Url: "https://kubernetes-charts.storage.googleapis.com"},
		{Name: "harlquin", Url: "https://kubernetes-charts.storage.googleapis.com"},
	}

	x := NewHelmCmd(&HelmCmdOptions{
		Repositories: presetRepos,
	})

	if err := x.Init(); err != nil {
		t.Fatal(".Init():", err)
	}

	repos, err := x.ListRepos()
	if err != nil {
		t.Fatal(err)
	}

	if len(repos) != 5 {
		t.Log("Shuold have found preset repos, 'stable' and 'local' repos only in", x)
		t.Log("Found instead", repos)
		t.Fail()
	}
}

func TestHelmAddRepo(t *testing.T) {
	x := NewHelmCmd(&HelmCmdOptions{})

	if err := x.Init(); err != nil {
		t.Fatal(err)
	}

	err := x.AddRepo(&HelmRepo{
		Name: "cool-repo",
		Url:  "https://kubernetes-charts.storage.googleapis.com",
	})

	if err != nil {
		t.Fatal(err)
	}

	repos, err := x.ListRepos()
	if err != nil {
		t.Fatal(err)
	}

	if len(repos) != 3 {
		t.Log("Shuold have found 'cool-repo', 'stable' and 'local' repos only in", x)
		t.Log("Found instead", repos)
		t.Fail()
	}
}

//func TestHelmInstall(t *testing.T) {
//	x := NewHelmService()
//	releases, err := x.InstallRelease(&HelmRelease{
//		Name:      "foobar",
//		Namespace: "new-namespace",
//		Repo:      "stable",
//		Chart:     "redis",
//		Version:   "3.0.6",
//	})
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	t.Log(releases)
//}
