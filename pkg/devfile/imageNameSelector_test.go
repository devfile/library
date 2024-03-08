//
// Copyright Red Hat
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package devfile

import (
	"fmt"
	"strings"
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Test_replaceImageNames(t *testing.T) {
	const (
		devfileName    = "my-component-app"
		targetRegistry = "localhost:5000/my-org/my-user"
		targetImageTag = "1.2.3-some-build-id"
	)

	const (
		absoluteImageNoTag  = "quay.io/some-org/absolute-image-no-tag"
		absoluteImageTag    = "ghcr.io/some-user/absolute-image-tag:11.22.33-dev"
		absoluteImageDigest = "docker.io/library/absolute-image-digest@sha256:26c68657ccce2cb0a31b330cb0be2b5e108d467f641c62e13ab40cbec258c68d"

		relativeImageNoTag  = "relative-image-no-tag:tag"
		relativeImageTag    = "relative-image-tag:latest"
		relativeImageTag2   = "org/user/relative-image-tag:2"
		relativeImageDigest = "library/relative-image-digest@sha256:36c68657ccce2cb0a31b330cb0be2b5e108d467f641c62e13ab40cbec258c68d"

		notMatchingRelativeImageNoTag  = "relative-image-no-tag2"
		notMatchingRelativeImageTag    = "relative-image-tag2:latest"
		notMatchingRelativeImageTag2   = "relative-image-tag2:2"
		notMatchingRelativeImageDigest = "relative-image-digest2@sha256:36c68657ccce2cb0a31b330cb0be2b5e108d467f641c62e13ab40cbec258c68d"
	)

	getImageNameReplacement := func(img string) string {
		return fmt.Sprintf("%s/%s-%s:%s", targetRegistry, devfileName, img, targetImageTag)
	}

	type args struct {
		devfileObjProvider func() (*parser.DevfileObj, error)
	}
	tests := []struct {
		name                    string
		args                    args
		wantErr                 bool
		wantImageComponents     []v1.Component
		wantContainerComponents []v1.Component
		wantK8sOcpComponents    []v1.Component
	}{
		{
			name: "relative image names not matching any container or K8s/Openshift components",
			args: args{
				devfileObjProvider: func() (*parser.DevfileObj, error) {
					dData, err := data.NewDevfileData(string(data.APISchemaVersion220))
					if err != nil {
						return nil, err
					}
					err = dData.AddComponents([]v1.Component{
						buildImageComponent("i-absolute-image-no-tag", absoluteImageNoTag),
						buildImageComponent("i-absolute-image-tag", absoluteImageTag),
						buildImageComponent("i-absolute-image-digest", absoluteImageDigest),
						buildImageComponent("i-relative-image-no-tag", relativeImageNoTag),
						buildImageComponent("i-relative-image-tag", relativeImageTag),
						buildImageComponent("i-relative-image-tag-2", relativeImageTag2),
						buildImageComponent("i-relative-image-digest", relativeImageDigest),

						buildContainerComponent("c-absolute-image-no-tag", absoluteImageNoTag),
						buildContainerComponent("c-absolute-image-tag", absoluteImageTag),
						buildContainerComponent("c-absolute-image-digest", absoluteImageDigest),
						buildContainerComponent("c-not-matching-relative-image-no-tag", notMatchingRelativeImageNoTag),
						buildContainerComponent("c-not-matching-relative-image-tag", notMatchingRelativeImageTag),
						buildContainerComponent("c-not-matching-relative-image-tag-2", notMatchingRelativeImageTag2),
						buildContainerComponent("c-not-matching-relative-image-digest", notMatchingRelativeImageDigest),

						buildInlinedKubernetesComponent("k-absolute-image-no-tag", absoluteImageNoTag, absoluteImageNoTag),
						buildInlinedKubernetesComponent("k-absolute-image-tag", absoluteImageNoTag, absoluteImageTag),
						buildInlinedKubernetesComponent("k-absolute-image-digest", absoluteImageNoTag, absoluteImageDigest),
						buildInlinedKubernetesComponent("k-not-matching-relative-image-no-tag", notMatchingRelativeImageNoTag, notMatchingRelativeImageNoTag),
						buildInlinedKubernetesComponent("k-not-matching-relative-image-tag", notMatchingRelativeImageTag, notMatchingRelativeImageTag),
						buildInlinedKubernetesComponent("k-not-matching-relative-image-tag-2", notMatchingRelativeImageTag2, notMatchingRelativeImageTag2),
						buildInlinedKubernetesComponent("k-not-matching-relative-image-digest", notMatchingRelativeImageDigest, notMatchingRelativeImageDigest),

						buildInlinedOpenshiftComponent("o-absolute-image-no-tag", absoluteImageNoTag, absoluteImageNoTag),
						buildInlinedOpenshiftComponent("o-absolute-image-tag", absoluteImageNoTag, absoluteImageTag),
						buildInlinedOpenshiftComponent("o-absolute-image-digest", absoluteImageNoTag, absoluteImageDigest),
						buildInlinedOpenshiftComponent("o-not-matching-relative-image-no-tag", notMatchingRelativeImageNoTag, notMatchingRelativeImageNoTag),
						buildInlinedOpenshiftComponent("o-not-matching-relative-image-tag", notMatchingRelativeImageTag, notMatchingRelativeImageTag),
						buildInlinedOpenshiftComponent("o-not-matching-relative-image-tag-2", notMatchingRelativeImageTag2, notMatchingRelativeImageTag2),
						buildInlinedOpenshiftComponent("o-not-matching-relative-image-digest", notMatchingRelativeImageDigest, notMatchingRelativeImageDigest),
					})
					if err != nil {
						return nil, err
					}
					metadata := dData.GetMetadata()
					metadata.Name = devfileName
					dData.SetMetadata(metadata)
					d := parser.DevfileObj{Data: dData}
					if err != nil {
						return nil, err
					}
					return &d, nil
				},
			},
			wantImageComponents: []v1.Component{
				//Relative image names should still be replaced
				buildImageComponent("i-absolute-image-no-tag", absoluteImageNoTag),
				buildImageComponent("i-absolute-image-tag", absoluteImageTag),
				buildImageComponent("i-absolute-image-digest", absoluteImageDigest),

				buildImageComponent("i-relative-image-no-tag", getImageNameReplacement("relative-image-no-tag")),
				buildImageComponent("i-relative-image-tag", getImageNameReplacement("relative-image-tag")),
				buildImageComponent("i-relative-image-tag-2", getImageNameReplacement("relative-image-tag")),
				buildImageComponent("i-relative-image-digest", getImageNameReplacement("relative-image-digest")),
			},
			wantContainerComponents: []v1.Component{
				// Should not change as none was matching
				buildContainerComponent("c-absolute-image-no-tag", absoluteImageNoTag),
				buildContainerComponent("c-absolute-image-tag", absoluteImageTag),
				buildContainerComponent("c-absolute-image-digest", absoluteImageDigest),
				buildContainerComponent("c-not-matching-relative-image-no-tag", notMatchingRelativeImageNoTag),
				buildContainerComponent("c-not-matching-relative-image-tag", notMatchingRelativeImageTag),
				buildContainerComponent("c-not-matching-relative-image-tag-2", notMatchingRelativeImageTag2),
				buildContainerComponent("c-not-matching-relative-image-digest", notMatchingRelativeImageDigest),
			},
			wantK8sOcpComponents: []v1.Component{
				// Should not change as none was matching
				buildInlinedKubernetesComponent("k-absolute-image-no-tag", absoluteImageNoTag, absoluteImageNoTag),
				buildInlinedKubernetesComponent("k-absolute-image-tag", absoluteImageNoTag, absoluteImageTag),
				buildInlinedKubernetesComponent("k-absolute-image-digest", absoluteImageNoTag, absoluteImageDigest),
				buildInlinedKubernetesComponent("k-not-matching-relative-image-no-tag", notMatchingRelativeImageNoTag, notMatchingRelativeImageNoTag),
				buildInlinedKubernetesComponent("k-not-matching-relative-image-tag", notMatchingRelativeImageTag, notMatchingRelativeImageTag),
				buildInlinedKubernetesComponent("k-not-matching-relative-image-tag-2", notMatchingRelativeImageTag2, notMatchingRelativeImageTag2),
				buildInlinedKubernetesComponent("k-not-matching-relative-image-digest", notMatchingRelativeImageDigest, notMatchingRelativeImageDigest),

				buildInlinedOpenshiftComponent("o-absolute-image-no-tag", absoluteImageNoTag, absoluteImageNoTag),
				buildInlinedOpenshiftComponent("o-absolute-image-tag", absoluteImageNoTag, absoluteImageTag),
				buildInlinedOpenshiftComponent("o-absolute-image-digest", absoluteImageNoTag, absoluteImageDigest),
				buildInlinedOpenshiftComponent("o-not-matching-relative-image-no-tag", notMatchingRelativeImageNoTag, notMatchingRelativeImageNoTag),
				buildInlinedOpenshiftComponent("o-not-matching-relative-image-tag", notMatchingRelativeImageTag, notMatchingRelativeImageTag),
				buildInlinedOpenshiftComponent("o-not-matching-relative-image-tag-2", notMatchingRelativeImageTag2, notMatchingRelativeImageTag2),
				buildInlinedOpenshiftComponent("o-not-matching-relative-image-digest", notMatchingRelativeImageDigest, notMatchingRelativeImageDigest),
			},
		},

		{
			name: "should update images for matching container or K8s/Openshift components",
			args: args{
				devfileObjProvider: func() (*parser.DevfileObj, error) {
					dData, err := data.NewDevfileData(string(data.APISchemaVersion220))
					if err != nil {
						return nil, err
					}
					err = dData.AddComponents([]v1.Component{
						buildImageComponent("i-absolute-image-no-tag", absoluteImageNoTag),
						buildImageComponent("i-absolute-image-tag", absoluteImageTag),
						buildImageComponent("i-absolute-image-digest", absoluteImageDigest),
						buildImageComponent("i-relative-image-no-tag", relativeImageNoTag),
						buildImageComponent("i-relative-image-tag", relativeImageTag),
						buildImageComponent("i-relative-image-tag-2", relativeImageTag2),
						buildImageComponent("i-relative-image-digest", relativeImageDigest),

						buildContainerComponent("c-absolute-image-no-tag", absoluteImageNoTag),
						buildContainerComponent("c-absolute-image-tag", absoluteImageTag),
						buildContainerComponent("c-absolute-image-digest", absoluteImageDigest),
						buildContainerComponent("c-matching-relative-image-no-tag", relativeImageNoTag),
						buildContainerComponent("c-matching-relative-image-tag", relativeImageTag),
						buildContainerComponent("c-matching-relative-image-tag-2", relativeImageTag2),
						buildContainerComponent("c-matching-relative-image-digest", relativeImageDigest),

						buildInlinedKubernetesComponent("k-absolute-image-no-tag", absoluteImageNoTag, absoluteImageNoTag),
						buildInlinedKubernetesComponent("k-absolute-image-tag", absoluteImageNoTag, absoluteImageTag),
						buildInlinedKubernetesComponent("k-absolute-image-digest", absoluteImageNoTag, absoluteImageDigest),
						buildInlinedKubernetesComponent("k-not-matching-relative-image-no-tag", notMatchingRelativeImageNoTag, notMatchingRelativeImageNoTag),
						buildInlinedKubernetesComponent("k-matching-relative-image-no-tag", relativeImageNoTag, relativeImageNoTag),
						buildInlinedKubernetesComponent("k-matching-relative-image-tag", relativeImageTag, relativeImageTag),
						buildInlinedKubernetesComponent("k-matching-relative-image-tag-2", relativeImageTag2, relativeImageTag2),
						buildInlinedKubernetesComponent("k-matching-relative-image-digest", relativeImageDigest, relativeImageDigest),

						buildInlinedOpenshiftComponent("o-absolute-image-no-tag", absoluteImageNoTag, absoluteImageNoTag),
						buildInlinedOpenshiftComponent("o-absolute-image-tag", absoluteImageNoTag, absoluteImageTag),
						buildInlinedOpenshiftComponent("o-absolute-image-digest", absoluteImageNoTag, absoluteImageDigest),
						buildInlinedOpenshiftComponent("o-matching-relative-image-no-tag", relativeImageNoTag, relativeImageNoTag),
						buildInlinedOpenshiftComponent("o-matching-relative-image-tag", relativeImageTag, relativeImageTag),
						buildInlinedOpenshiftComponent("o-matching-relative-image-tag-2", relativeImageTag2, relativeImageTag2),
						buildInlinedOpenshiftComponent("o-matching-relative-image-digest", relativeImageDigest, relativeImageDigest),
						buildInlinedOpenshiftComponent("o-not-matching-relative-image-digest", notMatchingRelativeImageDigest, notMatchingRelativeImageDigest),
					})
					if err != nil {
						return nil, err
					}
					metadata := dData.GetMetadata()
					metadata.Name = devfileName
					dData.SetMetadata(metadata)
					d := parser.DevfileObj{Data: dData}
					if err != nil {
						return nil, err
					}
					return &d, nil
				},
			},
			wantImageComponents: []v1.Component{
				buildImageComponent("i-absolute-image-no-tag", absoluteImageNoTag),
				buildImageComponent("i-absolute-image-tag", absoluteImageTag),
				buildImageComponent("i-absolute-image-digest", absoluteImageDigest),

				buildImageComponent("i-relative-image-no-tag", getImageNameReplacement("relative-image-no-tag")),
				buildImageComponent("i-relative-image-tag", getImageNameReplacement("relative-image-tag")),
				buildImageComponent("i-relative-image-tag-2", getImageNameReplacement("relative-image-tag")),
				buildImageComponent("i-relative-image-digest", getImageNameReplacement("relative-image-digest")),
			},
			wantContainerComponents: []v1.Component{
				buildContainerComponent("c-absolute-image-no-tag", absoluteImageNoTag),
				buildContainerComponent("c-absolute-image-tag", absoluteImageTag),
				buildContainerComponent("c-absolute-image-digest", absoluteImageDigest),
				buildContainerComponent("c-matching-relative-image-no-tag", getImageNameReplacement("relative-image-no-tag")),
				buildContainerComponent("c-matching-relative-image-tag", getImageNameReplacement("relative-image-tag")),
				buildContainerComponent("c-matching-relative-image-tag-2", getImageNameReplacement("relative-image-tag")),
				buildContainerComponent("c-matching-relative-image-digest", getImageNameReplacement("relative-image-digest")),
			},
			wantK8sOcpComponents: []v1.Component{
				buildInlinedKubernetesComponent("k-absolute-image-no-tag", absoluteImageNoTag, absoluteImageNoTag),
				buildInlinedKubernetesComponent("k-absolute-image-tag", absoluteImageNoTag, absoluteImageTag),
				buildInlinedKubernetesComponent("k-absolute-image-digest", absoluteImageNoTag, absoluteImageDigest),
				buildInlinedKubernetesComponent("k-not-matching-relative-image-no-tag", notMatchingRelativeImageNoTag, notMatchingRelativeImageNoTag),
				buildInlinedKubernetesComponent("k-matching-relative-image-no-tag", getImageNameReplacement("relative-image-no-tag"), relativeImageNoTag),
				buildInlinedKubernetesComponent("k-matching-relative-image-tag", getImageNameReplacement("relative-image-tag"), relativeImageTag),
				buildInlinedKubernetesComponent("k-matching-relative-image-tag-2", getImageNameReplacement("relative-image-tag"), relativeImageTag2),
				buildInlinedKubernetesComponent("k-matching-relative-image-digest", getImageNameReplacement("relative-image-digest"), relativeImageDigest),

				buildInlinedOpenshiftComponent("o-absolute-image-no-tag", absoluteImageNoTag, absoluteImageNoTag),
				buildInlinedOpenshiftComponent("o-absolute-image-tag", absoluteImageNoTag, absoluteImageTag),
				buildInlinedOpenshiftComponent("o-absolute-image-digest", absoluteImageNoTag, absoluteImageDigest),
				buildInlinedOpenshiftComponent("o-matching-relative-image-no-tag", getImageNameReplacement("relative-image-no-tag"), relativeImageNoTag),
				buildInlinedOpenshiftComponent("o-matching-relative-image-tag", getImageNameReplacement("relative-image-tag"), relativeImageTag),
				buildInlinedOpenshiftComponent("o-matching-relative-image-tag-2", getImageNameReplacement("relative-image-tag"), relativeImageTag2),
				buildInlinedOpenshiftComponent("o-matching-relative-image-digest", getImageNameReplacement("relative-image-digest"), relativeImageDigest),
				buildInlinedOpenshiftComponent("o-not-matching-relative-image-digest", notMatchingRelativeImageDigest, notMatchingRelativeImageDigest),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devfileObj, err := tt.args.devfileObjProvider()
			if err != nil {
				t.Errorf("unexpected error while building DevfileObj: %v", err)
				return
			}
			err = replaceImageNames(devfileObj, targetRegistry, targetImageTag)
			if tt.wantErr != (err != nil) {
				t.Errorf("replaceImageNames() unexpected error: %v, wantErr: %v", err, tt.wantErr)
			}

			lessFn := cmpopts.SortSlices(func(c1, c2 v1.Component) bool {
				return c1.Name < c2.Name
			})

			var imageComponents []v1.Component
			imageComponents, err = devfileObj.Data.GetComponents(common.DevfileOptions{
				ComponentOptions: common.ComponentOptions{ComponentType: v1.ImageComponentType},
			})
			if err != nil {
				t.Errorf("replaceImageNames() unexpected error while getting image components: %v", err)
				return
			}
			if diff := cmp.Diff(tt.wantImageComponents, imageComponents, cmpopts.EquateEmpty(), lessFn); diff != "" {
				t.Errorf("replaceImageNames() mismatch with image components (-want +got):\n%s", diff)
			}

			var containerComponents []v1.Component
			containerComponents, err = devfileObj.Data.GetComponents(common.DevfileOptions{
				ComponentOptions: common.ComponentOptions{ComponentType: v1.ContainerComponentType},
			})
			if err != nil {
				t.Errorf("replaceImageNames() unexpected error while getting container components: %v", err)
				return
			}
			if diff := cmp.Diff(tt.wantContainerComponents, containerComponents, cmpopts.EquateEmpty(), lessFn); diff != "" {
				t.Errorf("replaceImageNames() mismatch with container components (-want +got):\n%s", diff)
			}

			var allk8sOcComponents []v1.Component
			var k8sComponents []v1.Component
			k8sComponents, err = devfileObj.Data.GetComponents(common.DevfileOptions{
				ComponentOptions: common.ComponentOptions{ComponentType: v1.KubernetesComponentType},
			})
			if err != nil {
				t.Errorf("replaceImageNames() unexpected error while getting Kubernetes components: %v", err)
				return
			}
			allk8sOcComponents = append(allk8sOcComponents, k8sComponents...)
			var ocComponents []v1.Component
			ocComponents, err = devfileObj.Data.GetComponents(common.DevfileOptions{
				ComponentOptions: common.ComponentOptions{ComponentType: v1.OpenshiftComponentType},
			})
			if err != nil {
				t.Errorf("replaceImageNames() unexpected error while getting Openshift components: %v", err)
				return
			}
			allk8sOcComponents = append(allk8sOcComponents, ocComponents...)
			if diff := cmp.Diff(tt.wantK8sOcpComponents, allk8sOcComponents, cmpopts.EquateEmpty(), lessFn); diff != "" {
				t.Errorf("replaceImageNames() mismatch with Kubernetes and OpenShift components (-want +got):\n%s", diff)
			}
		})
	}
}

func buildImageComponent(cmpName, imageName string) v1.Component {
	return v1.Component{
		Name: cmpName,
		ComponentUnion: v1.ComponentUnion{
			Image: &v1.ImageComponent{
				Image: v1.Image{
					ImageName: imageName,
				},
			},
		},
	}
}

func buildContainerComponent(cmpName, image string) v1.Component {
	return v1.Component{
		Name: cmpName,
		ComponentUnion: v1.ComponentUnion{
			Container: &v1.ContainerComponent{
				Container: v1.Container{
					Image: image,
				},
			},
		},
	}
}

func buildInlinedKubernetesComponent(cmpName, imageName, crImageName string) v1.Component {
	return v1.Component{
		Name: cmpName,
		ComponentUnion: v1.ComponentUnion{
			Kubernetes: &v1.KubernetesComponent{
				K8sLikeComponent: v1.K8sLikeComponent{
					K8sLikeComponentLocation: v1.K8sLikeComponentLocation{
						Inlined: buildInlinedK8sResource(imageName, crImageName),
					},
				},
			},
		},
	}
}

func buildInlinedOpenshiftComponent(cmpName, imageName, crImageName string) v1.Component {
	return v1.Component{
		Name: cmpName,
		ComponentUnion: v1.ComponentUnion{
			Openshift: &v1.OpenshiftComponent{
				K8sLikeComponent: v1.K8sLikeComponent{
					K8sLikeComponentLocation: v1.K8sLikeComponentLocation{
						Inlined: buildInlinedK8sResource(imageName, crImageName),
					},
				},
			},
		},
	}
}

func buildInlinedK8sResource(imageName, crImageName string) string {
	return strings.Join([]string{
		buildCustomResource(),
		buildCustomResourceWithImageName(crImageName),
		buildPodManifest(imageName),
		buildDaemonSetManifest(imageName),
		buildDeploymentManifest(imageName),
		buildJobManifest(imageName),
		buildCronJobManifest(imageName),
		buildReplicaSetManifest(imageName),
		buildReplicationControllerManifest(imageName),
		buildStatefulSetManifest(imageName),
	}, "\n---\n")
}

func buildPodManifest(image string) string {
	return strings.TrimSpace(fmt.Sprintf(`
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: null
  name: my-pod
spec:
  containers:
  - image: %[1]s
    name: my-main-cont1
    resources: {}
  - image: my-other-image
    name: my-other-cont
    resources: {}
  ephemeralContainers:
  - image: %[1]s
    name: my-ephemeral-cont1
    resources: {}
  - image: my-ephemeral-container-image
    name: my-ephemeral-cont2
    resources: {}
  initContainers:
  - image: %[1]s
    name: my-init-cont1
    resources: {}
  - image: my-init-container-image
    name: my-init-cont2
    resources: {}
status: {}
`, image))
}

func buildDaemonSetManifest(image string) string {
	return strings.TrimSpace(fmt.Sprintf(`
apiVersion: apps/v1
kind: DaemonSet
metadata:
  creationTimestamp: null
  labels:
    app: my-app
  name: my-daemonset
spec:
  selector:
    matchLabels:
      name: my-app
  template:
    metadata:
      creationTimestamp: null
      labels:
        creationTimestamp: ""
        name: my-app
    spec:
      containers:
      - image: %[1]s
        name: my-main-cont1
        resources: {}
      - image: my-other-image
        name: my-main-cont1
        resources: {}
      initContainers:
      - image: %[1]s
        name: my-init-cont1
        resources: {}
      - image: my-init-container-image
        name: my-init-cont2
        resources: {}
  updateStrategy: {}
status:
  currentNumberScheduled: 0
  desiredNumberScheduled: 0
  numberMisscheduled: 0
  numberReady: 0
`, image))
}

func buildDeploymentManifest(image string) string {
	return strings.TrimSpace(fmt.Sprintf(`
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: my-app
  name: my-deployment
spec:
  replicas: 10
  selector:
    matchLabels:
      app: my-app
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: my-app
    spec:
      containers:
      - image: %[1]s
        name: my-main-cont1
        resources: {}
      - image: my-other-image
        name: my-main-cont1
        resources: {}
      initContainers:
      - image: %[1]s
        name: my-init-cont1
        resources: {}
      - image: my-init-container-image
        name: my-init-cont2
        resources: {}
status: {}
`, image))
}

func buildJobManifest(image string) string {
	return strings.TrimSpace(fmt.Sprintf(`
apiVersion: batch/v1
kind: Job
metadata:
  creationTimestamp: null
  name: my-job
spec:
  template:
    metadata:
      creationTimestamp: null
      name: my-app
    spec:
      containers:
      - image: %[1]s
        name: my-main-cont1
        resources: {}
      - image: my-other-image
        name: my-main-cont1
        resources: {}
      initContainers:
      - image: %[1]s
        name: my-init-cont1
        resources: {}
      - image: my-init-container-image
        name: my-init-cont2
        resources: {}
status: {}
`, image))
}

func buildCronJobManifest(image string) string {
	return strings.TrimSpace(fmt.Sprintf(`
apiVersion: batch/v1
kind: CronJob
metadata:
  creationTimestamp: null
  name: my-cron-job
spec:
  jobTemplate:
    metadata:
      creationTimestamp: null
    spec:
      template:
        metadata:
          creationTimestamp: null
          name: my-app
        spec:
          containers:
          - image: %[1]s
            name: my-main-cont1
            resources: {}
          - image: my-other-image
            name: my-main-cont1
            resources: {}
          initContainers:
          - image: %[1]s
            name: my-init-cont1
            resources: {}
          - image: my-init-container-image
            name: my-init-cont2
            resources: {}
  schedule: '*/1 * * * *'
status: {}
`, image))
}

func buildReplicaSetManifest(image string) string {
	return strings.TrimSpace(fmt.Sprintf(`
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  creationTimestamp: null
  labels:
    app: my-app
  name: my-replicaset
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: my-app
    spec:
      containers:
      - image: %[1]s
        name: my-main-cont1
        resources: {}
      - image: my-other-image
        name: my-main-cont1
        resources: {}
      initContainers:
      - image: %[1]s
        name: my-init-cont1
        resources: {}
      - image: my-init-container-image
        name: my-init-cont2
        resources: {}
status:
  replicas: 0
`, image))
}

func buildReplicationControllerManifest(image string) string {
	return strings.TrimSpace(fmt.Sprintf(`
apiVersion: v1
kind: ReplicationController
metadata:
  creationTimestamp: null
  name: my-replicationcontroller
spec:
  replicas: 3
  selector:
    app: my-app
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: my-app
      name: my-app
    spec:
      containers:
      - image: %[1]s
        name: my-main-cont1
        resources: {}
      - image: my-other-image
        name: my-main-cont1
        resources: {}
      initContainers:
      - image: %[1]s
        name: my-init-cont1
        resources: {}
      - image: my-init-container-image
        name: my-init-cont2
        resources: {}
status:
  replicas: 0
`, image))
}

func buildStatefulSetManifest(image string) string {
	return strings.TrimSpace(fmt.Sprintf(`
apiVersion: apps/v1
kind: StatefulSet
metadata:
  creationTimestamp: null
  name: my-statefulset
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
  serviceName: my-app
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: my-app
    spec:
      containers:
      - image: %[1]s
        name: my-main-cont1
        resources: {}
      - image: my-other-image
        name: my-main-cont1
        resources: {}
      initContainers:
      - image: %[1]s
        name: my-init-cont1
        resources: {}
      - image: my-init-container-image
        name: my-init-cont2
        resources: {}
  updateStrategy: {}
status:
  availableReplicas: 0
  replicas: 0
`, image))
}

func buildCustomResourceWithImageName(imageName string) string {
	return strings.TrimSpace(fmt.Sprintf(`
apiVersion: "stable.example.com/v1"
kind: CronTab
metadata:
  name: my-v1-crontab-cr
spec:
  cronSpec: "* * * * */5"
  image: %s
  someRandomField: 42
`, imageName))
}

func buildCustomResource() string {
	return strings.TrimSpace(`
apiVersion: "stable.example.com/v2"
kind: NewCronTab
metadata:
  name: my-v2-crontab-cr
spec:
  at: "every 1 hour"
  image: my-awesome-cron-image
  randomField: 77
`)
}
