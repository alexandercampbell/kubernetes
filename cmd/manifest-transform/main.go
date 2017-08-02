/*
Package main: implemention a manifest-based deployment transformer.

Input: manifest YAML, deployment JSON
Output: JSON deployment object modified as described in the Manifest

Background: https://docs.google.com/document/d/1AG4UsVblUAyIXTUcV0hiq7oRMyxogd9jjIxUtw9jUQk/edit#
*/
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	stdlog "log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v1"

	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
)

// Logger to stderr without the timestamp.
var log = stdlog.New(os.Stderr, "| ", 0)

// exitIf: quick util function to do sanity checks.
func exitIf(condition bool, exitMsg string, exitFormattingArgs ...interface{}) {
	if condition {
		log.Println("error: " + fmt.Sprintf(exitMsg, exitFormattingArgs...))
		os.Exit(1)
	}
}

// SupportedAPIVersion: document what we expect to see in the APIVersion field
// of a Manifest.
const SupportedAPIVersion = "manifest.k8s.io/v1alpha1"

// Manifest: describe some transforms to make to a Kubernetes object
// description.
type Manifest struct {
	ObjectMeta v1.ObjectMeta     "metadata"
	APIVersion string            "apiVersion"
	Kind       string            "kind"
	Labels     map[string]string "labels"
	NamePrefix string            "namePrefix"
}

func loadYAML(filename string, i interface{}) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bytes, i)
}

func loadJSON(filename string, i interface{}) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, i)
}

func recursivelyFindAllJSONFiles(directoryName string) ([]string, error) {
	fileInfos, err := ioutil.ReadDir(directoryName)
	toRead := []string{}

	exitIf(err != nil, "%v", err)
	for _, info := range fileInfos {
		if info.IsDir() {
			pathToDirectory := filepath.Join(directoryName, info.Name())
			nestedFiles, err := recursivelyFindAllJSONFiles(pathToDirectory)
			if err != nil {
				return nil, err
			}
			toRead = append(toRead, nestedFiles...)
		} else if strings.HasSuffix(info.Name(), ".json") {
			toRead = append(toRead, filepath.Join(directoryName, info.Name()))
		}
	}

	return toRead, nil
}

func applyManifestTransformsToDeployment(manifest Manifest,
	deployment *v1beta1.Deployment) {

	// Add the name prefix and the custom labels to the deployment.
	log.Println("Manifest name matches deployment name. Applying changes...")
	deployment.ObjectMeta.Name = manifest.NamePrefix + deployment.ObjectMeta.Name
	log.Printf("\tadding name prefix %q", manifest.NamePrefix)
	for labelName, labelValue := range manifest.Labels {
		log.Printf("\tadding label %v=%v", labelName, labelValue)
		deployment.Labels[labelName] = labelValue
		deployment.ObjectMeta.Labels[labelName] = labelValue
		deployment.Spec.Selector.MatchLabels[labelName] = labelValue
		deployment.Spec.Template.ObjectMeta.Labels[labelName] = labelValue
	}

	log.Println("Manifest transforms complete.")
	return nil
}

func main() {
	cmd := &cobra.Command{
		Use: "manifest-processor resources-directory/",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				cmd.Usage()
				return
			}
			dirName := args[0]

			var manifest Manifest
			manifestPath := filepath.Join(dirName, "manifest.yaml")
			exitIf(loadYAML(manifestPath, &manifest) != nil, "couldn't load manifest file")
			log.Printf("Loaded YAML manifest from %q", manifestPath)
			exitIf(manifest.Kind != "Overlay", "unsupported manifest kind %q", manifest.Kind)
			exitIf(manifest.APIVersion != SupportedAPIVersion,
				"unsupported API version %q", manifest.APIVersion)

			toRead, err := recursivelyFindAllJSONFiles(dirName)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%+v\n", toRead)

			for _, filename := range toRead {
				// Read Deployment
				var deployment v1beta1.Deployment
				err := loadJSON(filename, &deployment)
				exitIf(err != nil, "couldn't load deployment file %q: %v", filename, err)
				log.Printf("Loaded deployment from %q", args[2])
				exitIf(manifest.ObjectMeta.Name != deployment.ObjectMeta.Name,
					"Name mismatch (manifest=%q, deploy=%q)",
					manifest.ObjectMeta.Name, deployment.ObjectMeta.Name)
			}
		},
	}
	var filenameOptions resource.FilenameOptions
	cmdutil.AddFilenameOptionFlags(cmd, &filenameOptions, "")

	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}

	// Print the transformed Deployment.
	bytes, err := json.MarshalIndent(deployment, "", "    ")
	exitIf(err != nil, "unable to marshal deployment")
	fmt.Printf("\n%s\n", bytes)
}
