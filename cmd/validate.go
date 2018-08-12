package cmd

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"

	"github.com/valer-cara/mgo/pkg/manifest"
)

var (
	validateCluster string
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate manifests",
	Long:  "",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := doValidate()
		if err != nil {
			log.Fatal(err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVar(&validateCluster, "cluster", "", "Cluster to validate, as given by 'kubectl config get-contexts'. Eg: minikube")
	validateCmd.MarkFlagRequired("cluster")
}

func doValidate() error {
	var allGood = true

	manifests, err := manifest.FindManifests(gitopsRepo, validateCluster)
	if err != nil {
		log.Printf("Cannot locate manifests in repo %s for cluster %s", gitopsRepo, validateCluster)
		return err
	}

	for _, file := range manifests.Helm {
		header, err := manifest.ParseHeader(file)
		if err != nil {
			log.Printf("Cannot parse file %s: %v", file, err)
			allGood = false
		}

		err = header.Validate()
		if err != nil {
			log.Printf("File %s invalid: %v", file, err)
			allGood = false
		}
	}

	if !allGood {
		return errors.New("Validating manifests failed.")
	}

	log.Println("Manifests are valid!")
	return nil
}
