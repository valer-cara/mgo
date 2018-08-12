package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"

	"github.com/valer-cara/mgo/pkg/server"
)

var (
	serveAddr   string
	kubeconfig  string
	kubecontext string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "API server: handles new deployments and syncs each time",
	Long:  "",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := doServe()
		if err != nil {
			log.Fatal(err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&serveAddr, "listen", "l", "127.0.0.1:8080", "Listen to address")
	serveCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Alternative kubeconfig for helm")
	serveCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Dry Run mode")
}

func doServe() error {
	if kubeconfig == "" {
		kubeconfig = getKubeconfig()
	}

	helmHome := getHelmHome()

	log.Printf("Starting server on %s.", serveAddr)
	log.Printf("  - handling gitops repo at %s", gitopsRepo)
	log.Printf("  - kubeconfig at %s", kubeconfig)
	log.Printf("  - helmHome at %s", helmHome)

	serv := server.NewServer(
		serveAddr,
		gitopsRepo,
		helmHome,
		kubeconfig,
		dryRun,
	)
	err := serv.Serve()
	if err != nil {
		return err
	}

	return nil
}
