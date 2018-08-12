package cmd

import (
	log "github.com/sirupsen/logrus"
	"os"
	"path"

	"github.com/valer-cara/mgo/pkg/config"
	"github.com/spf13/cobra"
)

var (
	gitopsRepo   string
	debugLogging bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "mygitops",
	Short: "MyGitops: a tool to help with gitops on kubernetes",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().BoolVar(&debugLogging, "debug", false, "Enable debug logging")
	RootCmd.PersistentFlags().StringVar(&gitopsRepo, "gitops-repo", "", "Location of your gitops repository")
	RootCmd.MarkPersistentFlagRequired("gitops-repo")
}

func initConfig() {
	// cobra/required flags are processed after this
	if len(gitopsRepo) > 0 {
		configFilePath := path.Join(gitopsRepo, "/mygitops.yaml")
		err := config.LoadGlobalConfig(configFilePath)
		if err != nil {
			log.Printf("Cannot load configuration file %s: %v", configFilePath, err)
			os.Exit(1)
		}
	}

	if debugLogging {
		log.SetLevel(log.DebugLevel)
	}
}

func getKubeconfig() string {
	return path.Join(os.Getenv("HOME"), ".kube/config")
}
