/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	apimachineryversion "k8s.io/apimachinery/pkg/version"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/util/i18n"
	"k8s.io/kubernetes/pkg/version"
)

type Version struct {
	ClientVersion *apimachineryversion.Info `json:"clientVersion,omitempty" yaml:"clientVersion,omitempty"`
	ServerVersion *apimachineryversion.Info `json:"serverVersion,omitempty" yaml:"serverVersion,omitempty"`
}

var (
	versionExample = templates.Examples(i18n.T(`
		# Print the client and server versions for the current context
		kubectl version`))
)

func NewCmdVersion(f cmdutil.Factory, out io.Writer) *cobra.Command {
	options := new(versionOptions)

	cmd := &cobra.Command{
		Use:     "version",
		Short:   i18n.T("Print the client and server version information"),
		Long:    "Print the client and server version information for the current context",
		Example: versionExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(options.Complete(out))
			cmdutil.CheckErr(options.RunVersion(f))
		},
	}

	fl := cmd.Flags()
	fl.BoolVarP(&options.clientVersionOnly, "client", "c", false, "Print only the kubectl version. Does not attempt to connect to server.")
	fl.BoolVarP(&options.versionNumberOnly, "short", "", false, "Print only the version number. Has no effect when used with --output")
	fl.StringVar(&options.outputFormat, "output", "", "Optionally format the version output as 'yaml' or 'json'")
	fl.MarkShorthandDeprecated("client", "please use --client instead.")

	return cmd
}

type versionOptions struct {
	clientVersionOnly bool
	versionNumberOnly bool
	outputFormat      string

	out io.Writer
}

func (v *versionOptions) Complete(out io.Writer) error {
	v.out = out
	return nil
}

func (v *versionOptions) RunVersion(f cmdutil.Factory) error {
	var serverVersion *apimachineryversion.Info
	var serverErr error
	vo := new(Version)

	clientVersion := version.Get()
	vo.ClientVersion = &clientVersion

	if !v.clientVersionOnly {
		serverVersion, serverErr = retrieveServerVersion(f)
		vo.ServerVersion = serverVersion
	}

	switch v.outputFormat {
	case "":
		fmtVersion := func(i apimachineryversion.Info) string {
			if v.versionNumberOnly {
				return i.GitVersion
			}
			return fmt.Sprintf("%#v", i)
		}

		fmt.Fprintf(v.out, "Client Version: %s\n", fmtVersion(clientVersion))
		if serverVersion != nil {
			fmt.Fprintf(v.out, "Server Version: %s\n", fmtVersion(*serverVersion))
		}

	case "yaml":
		y, err := yaml.Marshal(&vo)
		if err != nil {
			return err
		}

		fmt.Fprintln(v.out, string(y))

	case "json":
		y, err := json.Marshal(&vo)
		if err != nil {
			return err
		}
		fmt.Fprintln(v.out, string(y))

	default:
		return errors.New("invalid output format: " + v.outputFormat)
	}

	return serverErr
}

func retrieveServerVersion(f cmdutil.Factory) (*apimachineryversion.Info, error) {
	discoveryClient, err := f.DiscoveryClient()
	if err != nil {
		return nil, err
	}

	// Always request fresh data from the server
	discoveryClient.Invalidate()

	serverVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		return nil, err
	}

	return serverVersion, nil
}
