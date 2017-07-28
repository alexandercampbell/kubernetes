
# manifest-transform

[alexandercampbell](github.com/alexandercampbell)

manifest-transform is a prototype for editing a Deployment in an automated way
according to a Manifest file.

### Example

manifest.yaml:

```yaml
apiVersion: manifest.k8s.io/v1alpha1
kind: Overlay
metadata:
  name: hello-world
description: Mungebot config for test-infra repo
namePrefix: test-infra-
labels:
  org: kubernetes
  repo: test-infra
annotations:
  note: This is my first try
bases:
- ../../package
#These are strategic merge patch overlays in the form of API resources
resources:
- deployment.yaml
#There could also be configmaps in Base, which would make these overlays
configmaps:
- type: env
  name: app-env
  file: app.env
- type: file
  name: app-config
  file: app-init.ini
#There could be secrets in Base, if just using a fork/rebase workflow
secrets:
- type: tls
  certFile: tls.cert
  keyFile: tls.key
recursive: false
```

deployment.json:

```json
{
    "apiVersion": "extensions/v1beta1",
    "kind": "Deployment",
    "metadata": {
        "annotations": {
            "deployment.kubernetes.io/revision": "1",
            "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"extensions/v1beta1\",\"kind\":\"Deployment\",\"metadata\":{\"annotations\":{},\"creationTimestamp\":null,\"labels\":{\"app\":\"hello-world\"},\"name\":\"hello-world\",\"namespace\":\"default\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"hello-world\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"hello-world\"}},\"spec\":{\"containers\":[{\"image\":\"nginx\",\"name\":\"nginx\",\"resources\":{}}]}}},\"status\":{}}\n"
        },
        "creationTimestamp": "2017-07-27T17:51:27Z",
        "generation": 1,
        "labels": {
            "app": "hello-world"
        },
        "name": "hello-world",
        "namespace": "default",
        "resourceVersion": "3074118",
        "selfLink": "/apis/extensions/v1beta1/namespaces/default/deployments/hello-world",
        "uid": "3687ddc4-72f4-11e7-a781-080027a4747b"
    },
    "spec": {
        "replicas": 1,
        "selector": {
            "matchLabels": {
                "app": "hello-world"
            }
        },
        "strategy": {
            "rollingUpdate": {
                "maxSurge": 1,
                "maxUnavailable": 1
            },
            "type": "RollingUpdate"
        },
        "template": {
            "metadata": {
                "creationTimestamp": null,
                "labels": {
                    "app": "hello-world"
                }
            },
            "spec": {
                "containers": [
                    {
                        "image": "nginx",
                        "imagePullPolicy": "Always",
                        "name": "nginx",
                        "resources": {},
                        "terminationMessagePath": "/dev/termination-log",
                        "terminationMessagePolicy": "File"
                    }
                ],
                "dnsPolicy": "ClusterFirst",
                "restartPolicy": "Always",
                "schedulerName": "default-scheduler",
                "securityContext": {},
                "terminationGracePeriodSeconds": 30
            }
        }
    },
    "status": {
        "availableReplicas": 1,
        "conditions": [
            {
                "lastTransitionTime": "2017-07-27T17:51:27Z",
                "lastUpdateTime": "2017-07-27T17:51:27Z",
                "message": "Deployment has minimum availability.",
                "reason": "MinimumReplicasAvailable",
                "status": "True",
                "type": "Available"
            }
        ],
        "observedGeneration": 1,
        "readyReplicas": 1,
        "replicas": 1,
        "updatedReplicas": 1
    }
}
```

Executed command:

```bash
$ manifest-transform manifest.yaml deployment.json > new-deployment.json
| Loaded YAML manifest from "manifest.yaml"
| Loaded deployment from "deployment.json"
| Manifest name matches deployment name. Applying changes...
|       adding name prefix "test-infra-"
|       adding label org=kubernetes
|       adding label repo=test-infra
| Manifest transforms complete.
$ 
```

new-deployment.json:

```json
{
    "kind": "Deployment",
    "apiVersion": "extensions/v1beta1",
    "metadata": {
        "name": "test-infra-hello-world",
        "namespace": "default",
        "selfLink": "/apis/extensions/v1beta1/namespaces/default/deployments/hello-world",
        "uid": "3687ddc4-72f4-11e7-a781-080027a4747b",
        "resourceVersion": "3074118",
        "generation": 1,
        "creationTimestamp": "2017-07-27T17:51:27Z",
        "labels": {
            "app": "hello-world",
            "org": "kubernetes",
            "repo": "test-infra"
        },
        "annotations": {
            "deployment.kubernetes.io/revision": "1",
            "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"extensions/v1beta1\",\"kind\":\"Deployment\",\"metadata\":{\"annotations\":{},\"creationTimestamp\":null,\"labels\":{\"app\":\"hello-world\"},\"name\":\"hello-world\",\"namespace\":\"default\"},\"spec\":{\"replicas\":1,\"selector\":{\"matchLabels\":{\"app\":\"hello-world\"}},\"strategy\":{},\"template\":{\"metadata\":{\"creationTimestamp\":null,\"labels\":{\"app\":\"hello-world\"}},\"spec\":{\"containers\":[{\"image\":\"nginx\",\"name\":\"nginx\",\"resources\":{}}]}}},\"status\":{}}\n"
        }
    },
    "spec": {
        "replicas": 1,
        "selector": {
            "matchLabels": {
                "app": "hello-world",
                "org": "kubernetes",
                "repo": "test-infra"
            }
        },
        "template": {
            "metadata": {
                "creationTimestamp": null,
                "labels": {
                    "app": "hello-world",
                    "org": "kubernetes",
                    "repo": "test-infra"
                }
            },
            "spec": {
                "containers": [
                    {
                        "name": "nginx",
                        "image": "nginx",
                        "resources": {},
                        "terminationMessagePath": "/dev/termination-log",
                        "terminationMessagePolicy": "File",
                        "imagePullPolicy": "Always"
                    }
                ],
                "restartPolicy": "Always",
                "terminationGracePeriodSeconds": 30,
                "dnsPolicy": "ClusterFirst",
                "securityContext": {},
                "schedulerName": "default-scheduler"
            }
        },
        "strategy": {
            "type": "RollingUpdate",
            "rollingUpdate": {
                "maxUnavailable": 1,
                "maxSurge": 1
            }
        }
    },
    "status": {
        "observedGeneration": 1,
        "replicas": 1,
        "updatedReplicas": 1,
        "readyReplicas": 1,
        "availableReplicas": 1,
        "conditions": [
            {
                "type": "Available",
                "status": "True",
                "lastUpdateTime": "2017-07-27T17:51:27Z",
                "lastTransitionTime": "2017-07-27T17:51:27Z",
                "reason": "MinimumReplicasAvailable",
                "message": "Deployment has minimum availability."
            }
        ]
    }
}
```

