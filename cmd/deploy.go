package cmd

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"strings"

	"github.com/valer-cara/mgo/pkg/deploy"
	"github.com/valer-cara/mgo/pkg/manifest"
	"github.com/valer-cara/mgo/pkg/services"
)

var (
	deployCluster string
	deploySource  string
	deployImage   string
	deployAuthor  string
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Create a new commit in gitops repo recording a new deployment",
	Long:  "",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := doDeploy()
		if err != nil {
			log.Fatal(err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringVar(&deployCluster, "cluster", "", "Cluster to deploy to, as given by 'kubectl config get-contexts'. Eg: minikube")
	deployCmd.MarkFlagRequired("cluster")
	deployCmd.Flags().StringVar(&deploySource, "source", "", "Repo associated with this deploy, as recored in the '__mygitops' section. Eg: github.com/foo/bar")
	deployCmd.MarkFlagRequired("source")
	deployCmd.Flags().StringVar(&deployImage, "image", "", "Docker image for the new deployment. Eg: quay.io/foo/bar")
	deployCmd.MarkFlagRequired("image")
	deployCmd.Flags().StringVar(&deployAuthor, "author", "", "Author recorded for this deployment. Eg: linus@kernel.org")
	deployCmd.MarkFlagRequired("author")
}

func doDeploy() error {
	deploySvc := services.NewDeployService(gitopsRepo, &deploy.DeployOptions{
		TriggerRepo: deploySource,
		Image:       splitImage(deployImage),
		Author:      deployAuthor,
		Cluster:     deployCluster,
	})

	if err := deploySvc.Execute(); err != nil {
		return errors.New(fmt.Sprintf("Failed deployment %v: %v", deploySvc, err))
	}

	return nil
}

func splitImage(image string) manifest.HeaderImage {
	tag := "latest"
	separatedImage := strings.Split(image, ":")

	if len(separatedImage) > 1 {
		tag = separatedImage[1]
	}

	return manifest.HeaderImage{
		Repository: separatedImage[0],
		Tag:        tag,
	}
}
