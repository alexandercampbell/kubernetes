package resource

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	yaml "gopkg.in/yaml.v1"

	"k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
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
	// Somehow the object that gets passed in here is of concrete type
	// (*unstructured.Unstructured).
	// This is not useful.
	// I need to be able to access the label fields and so on.
	// How can I get the Unstructured object transformed into the
	// appropriate extensionsv1beta1.Deployment type?
	log.Printf("object type: %T", object)

	// well I can try marshalling to JSON and back
	// WEW this is bad code
	var deployment extensionsv1beta1.Deployment
	bytes, _ := json.MarshalIndent(object, "", "     ")
	if err := json.Unmarshal(bytes, &deployment); err != nil {
		panic(err)
	}

	// alright thanks to some shenanigans I've got ahold of a Deployment
	// object. Time to apply the manifest.
	applyManifestTransformsToDeployment(*manifest, &deployment)

	// Since we did some hacks to get the Deployment I can't exactly put it
	// back into the runtime.Object. So for now I will settle for printing
	// the new object.
	log.Printf("Put the name prefix on the deployment:")
	bytes, _ = json.MarshalIndent(deployment, "", "     ")
	log.Printf("%s\n", bytes)

	os.Exit(0)

	// Type switch would be nice here but as noted above the concrete type
	// is wrong.
	/*
		switch v := object.(type) {
		case :
			log.Printf("saw Deployment %v", v)
		}
	*/

	return nil
}
