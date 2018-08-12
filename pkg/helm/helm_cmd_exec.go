package helm

import (
	log "github.com/sirupsen/logrus"
	"io"
	"strings"

	"github.com/valer-cara/mgo/pkg/util"
)

type HelmCmdExec struct {
	writer      io.Writer
	env         []string
	kubecontext string
}

func NewHelmCmdExec(writer io.Writer, env []string, kubecontext string) *HelmCmdExec {
	return &HelmCmdExec{
		writer:      writer,
		env:         env,
		kubecontext: kubecontext,
	}
}

// Exec helm with provided arguments and environement set in NewHelmCmdExec()
// Output is also written to the provided io.Writer
//
// Returns command's output and an error
func (h *HelmCmdExec) Exec(args ...string) ([]byte, error) {
	if h.kubecontext != "" {
		args = append([]string{
			"--kube-context=" + h.kubecontext,
		}, args...)
	}

	log.Debugf("  - running: helm %s", strings.Join(args, " "))

	out, err := util.Exec("helm", args, h.env)

	if h.writer != nil {
		h.writer.Write(out)
	}

	return out, err
}
