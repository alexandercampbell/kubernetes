/*
Copyright 2016 The Kubernetes Authors.

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
	"fmt"
	"io"

	"github.com/spf13/cobra"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	"k8s.io/kubernetes/pkg/kubectl"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/util/i18n"
)

var (
	deploymentLong = templates.LongDesc(i18n.T(`
    Create a deployment with the specified name.`))

	deploymentExample = templates.Examples(i18n.T(`
    # Create a new deployment named my-dep that runs the busybox image.
    kubectl create deployment my-dep --image=busybox`))
)

////////////////////////////////////////////////////////////////////////////////
// This block is unique because it contains structs and functions common to
// `kubectl run` and `kubectl create deployment`.
// See https://github.com/kubernetes/kubectl/issues/11 for discussion.
////////////////////////////////////////////////////////////////////////////////

// Describe the common subset of options.
type DeploymentOptions struct {

	// This struct contains a good portion of the options we need for a
	// deployment, like DryRun.
	CreateSubcommandOptions

	Image []string
}

// Modify the FlagSet to include flags corresponding to the fields of
// DeploymentOptions.
func addDeploymentOptionsFlags(cmd *cobra.Command) {
	// "inherit" the standard flags that go along with any command that
	// might print an object after completion.
	cmdutil.AddPrinterFlags(cmd)

	// Add the --save-config flag.
	cmdutil.AddApplyAnnotationFlags(cmd)

	fs := cmd.Flags()
	fs.StringSlice("image", []string{}, "Image name to run.")
	cmd.MarkFlagRequired("image")
}

// Read the flags of cmd into a struct describing the standard options for
// deployment.
func readDeploymentOptions(cmd *cobra.Command) DeploymentOptions {
	options := DeploymentOptions{}
	options.OutputFormat = cmdutil.GetFlagString(cmd, "output")
	options.DryRun = cmdutil.GetDryRunFlag(cmd)
	options.Image = cmdutil.GetFlagStringSlice(cmd, "image")
	options.ApplyAnnotations = cmdutil.GetFlagBool(cmd, cmdutil.ApplyAnnotationsFlag)

	return options
}

////////////////////////////////////////////////////////////////////////////////

// NewCmdCreateDeployment is a macro command to create a new deployment.
// This command is better known to users as `kubectl create deployment`.
// Note that this command overlaps significantly with the `kubectl run` command.
func NewCmdCreateDeployment(f cmdutil.Factory, cmdOut, cmdErr io.Writer) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "deployment NAME --image=image [--dry-run]",
		Aliases: []string{"deploy"},
		Short:   i18n.T("Create a deployment with the specified name."),
		Long:    deploymentLong,
		Example: deploymentExample,
		Run: func(cmd *cobra.Command, args []string) {
			options := readDeploymentOptions(cmd)
			err := createDeployment(f, options, cmdOut, cmdErr, cmd, args)
			cmdutil.CheckErr(err)
		},
	}

	addDeploymentOptionsFlags(cmd)

	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddGeneratorFlags(cmd, cmdutil.DeploymentBasicV1Beta1GeneratorName)
	return cmd
}

func createDeployment(f cmdutil.Factory, options DeploymentOptions,
	cmdOut, cmdErr io.Writer, cmd *cobra.Command, args []string) error {

	var err error
	options.Name, err = NameFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}

	clientset, err := f.ClientSet()
	if err != nil {
		return err
	}
	resourcesList, err := clientset.Discovery().ServerResources()
	// ServerResources ignores errors for old servers do not expose discovery
	if err != nil {
		return fmt.Errorf("failed to discover supported resources: %v", err)
	}
	generatorName := cmdutil.GetFlagString(cmd, "generator")
	// fallback to the old generator if server does not support apps/v1beta1 deployments
	if generatorName == cmdutil.DeploymentBasicAppsV1Beta1GeneratorName &&
		!contains(resourcesList, appsv1beta1.SchemeGroupVersion.WithResource("deployments")) {
		fmt.Fprintf(cmdErr, "WARNING: New deployments generator specified (%s), but apps/v1beta1.Deployments are not available, falling back to the old one (%s).\n",
			cmdutil.DeploymentBasicAppsV1Beta1GeneratorName, cmdutil.DeploymentBasicV1Beta1GeneratorName)
		generatorName = cmdutil.DeploymentBasicV1Beta1GeneratorName
	}

	switch generatorName {
	case cmdutil.DeploymentBasicAppsV1Beta1GeneratorName:
		options.StructuredGenerator = &kubectl.DeploymentBasicAppsGeneratorV1{
			Name:   options.Name,
			Images: options.Image,
		}
	case cmdutil.DeploymentBasicV1Beta1GeneratorName:
		options.StructuredGenerator = &kubectl.DeploymentBasicGeneratorV1{
			Name:   options.Name,
			Images: options.Image,
		}
	default:
		return errUnsupportedGenerator(cmd, generatorName)
	}

	return RunCreateSubcommand(f, cmd, cmdOut,
		&options.CreateSubcommandOptions)
}
