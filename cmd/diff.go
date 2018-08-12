package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"strings"

	"github.com/valer-cara/mgo/pkg/config"
	"github.com/valer-cara/mgo/pkg/helm"
	"github.com/valer-cara/mgo/pkg/jobs"
	"github.com/valer-cara/mgo/pkg/manifest"
)

var (
	diffCluster     string
	diffHelmService *helm.HelmCmd
)

type diffJob struct {
	Header *helm.HelmRelease
	Files  []string
}

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Diff a running cluster against the state set in a gitops repo",
	Long:  "",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		var (
			diffJobs []interface{} = make([]interface{}, 0)
		)

		diffHelmService, err := diffInitHelmService()
		if err != nil {
			log.Println("Error initializing: ", err)
			os.Exit(1)
		}

		diffHelmService.SetOutput(os.Stdout)

		manifests, err := manifest.FindManifests(gitopsRepo, diffCluster)
		if err != nil {
			log.Errorf("Cannot find manifests for cluster %s in repo %s", diffCluster, gitopsRepo)
			os.Exit(1)
		}

		for _, file := range manifests.Helm {
			//echo "helm diff upgrade $rel $chart --version $version --values ./$rel-values.yaml $extra > $outfile" >> $COMMANDS
			header, err := manifest.ParseHeader(file)
			if err != nil {
				log.Errorf("Cannot parse `__mygitops` section in `%s`", file)
				os.Exit(1)
			}

			if err := header.Validate(); err != nil {
				log.Errorf("File %s has an invalid `__mygitops` section: %v", file, err)
				os.Exit(1)
			}

			files := []string{file}

			potentialSecretsFilePath := strings.Replace(file, "-values.yaml", "-secrets.yaml", -1)
			if _, err := os.Stat(potentialSecretsFilePath); err == nil {
				files = append(files, potentialSecretsFilePath)
			}

			diffJobs = append(diffJobs, diffJob{
				Header: &header.HelmRelease,
				Files:  files,
			})
		}

		jobs.Parallel(func(job interface{}) error {
			j := job.(diffJob)

			// Helm service outputs to stdout as set above
			err = diffHelmService.DiffRelease(j.Header, j.Files)
			if err != nil {
				log.Errorf("Cannot diff release %s: %v", j.Header.Name, err)
				os.Exit(1)
			}

			return nil
		}, diffJobs, &jobs.ParallelOpts{MaxParallel: 15})
	},
}

func init() {
	RootCmd.AddCommand(diffCmd)
	diffCmd.Flags().StringVar(&diffCluster, "cluster", "", "Cluster to deploy to, as given by 'kubectl config get-contexts'. Eg: minikube")
	diffCmd.MarkFlagRequired("cluster")
}

func diffInitHelmService() (*helm.HelmCmd, error) {
	helmService := helm.NewHelmCmd(&helm.HelmCmdOptions{
		HelmHome:     getHelmHome(),
		Kubeconfig:   getKubeconfig(),
		Kubecontext:  diffCluster,
		Repositories: config.Global.Helm.Repositories,
	})
	if err := helmService.Init(); err != nil {
		return nil, err
	}

	return helmService, nil
}
