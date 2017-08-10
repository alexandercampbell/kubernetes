package resource

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/golang/glog"

	yaml "gopkg.in/yaml.v1"

	"k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const manifestFilename = "Kubernetes-manifest.yaml"

// Manifest: describe some transforms to make to a Kubernetes object
// description.
type Manifest struct {
	ObjectMeta v1.ObjectMeta     "metadata"
	APIVersion string            "apiVersion"
	Kind       string            "kind"
	Labels     map[string]string "labels"
	NamePrefix string            "namePrefix"
}

func loadManifest(filename string) (*Manifest, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var manifest Manifest
	err = yaml.Unmarshal(bytes, &manifest)
	return &manifest, err
}

func applyManifestTransformsToDeployment(manifest Manifest,
	deployment *extensionsv1beta1.Deployment) {

	// Add the name prefix and the custom labels to the deployment.
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
}

func applyManifestToResource(manifest *Manifest, object runtime.Object) error {
	v5 := glog.V(5)
	// Somehow the object that gets passed in here is of concrete type
	// (*unstructured.Unstructured).
	// This is not useful.
	v5.Infof("object type: %T", object)

	// After discussing with Phil yesterday I know the best way to solve
	// this is just to walk the JSON fields of the object as they are right
	// now.
	unstructured, ok := object.(*unstructured.Unstructured)
	if !ok {
		v5.Info("could not cast runtime.Object to unstructured.Unstructured")
		return nil
	}

	// left the type in here for clarity on what is going on.
	var topLevelFields map[string]interface{}
	topLevelFields = unstructured.Object

	// Check that the type is what we expect.
	apiVersion, ok := topLevelFields["apiVersion"].(string)
	if !ok || apiVersion != "extensions/v1beta1" {
		return nil
	}
	kind, ok := topLevelFields["kind"].(string)
	if !ok || kind != "Deployment" {
		return nil
	}

	// Load the metadata and update the metadata.name field to include the
	// manifest's NamePrefix.
	metadata, ok := topLevelFields["metadata"].(map[string]interface{})
	if !ok {
		v5.Info("could not access 'metadata' field as map[string]interface{} in object")
		return nil
	}
	metadataName, ok := metadata["name"].(string)
	if !ok {
		glog.V(5).Info("could not find 'name' field in object metadata")
		return nil
	}
	metadataName = manifest.NamePrefix + metadataName
	v5.Infof("new metadata.name: %q", metadataName)
	unstructured.Object["metadata"].(map[string]interface{})["name"] = metadataName

	// debugging
	bytes, _ := json.MarshalIndent(object, "", "\t")
	v5.Infof("%s", bytes)

	return nil
}
