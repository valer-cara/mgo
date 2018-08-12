package cmd

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"

	"github.com/valer-cara/mgo/pkg/services"
)

var (
	syncCluster string
	dryRun      bool
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync given cluster with the state in the gitops repo",
	Long:  "",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := doSync()
		if err != nil {
			log.Fatal(err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(syncCmd)
	syncCmd.Flags().StringVar(&syncCluster, "cluster", "", "Cluster to sync, as given by 'kubectl config get-contexts'. Eg: minikube")
	syncCmd.MarkFlagRequired("cluster")
	syncCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Don't do any actual changes")
}

func doSync() error {
	helmHome := getHelmHome()
	syncSvc := services.NewSyncService(
		gitopsRepo,
		helmHome,
		getKubeconfig(),
		syncCluster,
		dryRun,
	)

	if err := syncSvc.Init(); err != nil {
		return errors.New(fmt.Sprintf("Failed init sync %v: %v", syncSvc, err))
	}

	if err := syncSvc.Execute(); err != nil {
		return errors.New(fmt.Sprintf("Failed sync %v: %v", syncSvc, err))
	}

	return nil
}

func getHelmHome() string {
	// TODO: figure out what helm home is
	//return ""

	if helmHome := os.Getenv("HELM_HOME"); helmHome != "" {
		return helmHome
	}

	// TODO: if we're to use ~/.helm, we should drop provisioning stuff automatically
	// That's too clever
	//defaultPath := path.Join(os.Getenv("HOME"), ".helm")
	//if _, err := os.Stat(defaultPath); err == nil {
	//	return defaultPath
	//}

	// Empty helm home -> new helm home in temporary dir created by service
	return ""
}
