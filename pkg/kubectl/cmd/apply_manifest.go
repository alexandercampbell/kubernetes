package cmd

import (
	"io"

	"github.com/spf13/cobra"

	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

type ApplyManifestOptions struct {
	ApplyOptions
}

// NewCmdApplyManifest is nearly the same thing as NewCmdApply, but there are
// two differences.
//
//  1. We want the searchForManifest parameter to be true in the Builder system.
//  2. Our help message should explain the Manifest system.
//
func NewCmdApplyManifest(baseName string, f cmdutil.Factory,
	out, errOut io.Writer) *cobra.Command {

	cmd := NewCmdApply(baseName, f, out, errOut, true)
	cmd.Use = "apply-manifest [path]"
	cmd.Short = "Use a Manifest file to apply a set of configurations"

	return cmd
}
