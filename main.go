package main

import (
	"fmt"
	"os"

	"github.com/valer-cara/mgo/cmd"
	log "github.com/sirupsen/logrus"
)

var (
	buildVersion string
)

func main() {
	cmd.RootCmd.Version = buildVersion

	log.Printf("MyGitops (build '%s')", buildVersion)

	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
