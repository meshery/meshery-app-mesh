// Copyright 2020 Layer5, Inc.
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

package appmesh

import (
	"context"

	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (appMesh *AppMesh) installSampleApp(namespace string, del bool, templates []adapter.Template) (string, error) {
	st := status.Installing

	if del {
		st = status.Removing
	}

	for _, template := range templates {
		err := appMesh.applyManifest([]byte(template.String()), del, namespace)
		if err != nil {
			return st, ErrSampleApp(err, st)
		}
	}

	return status.Installed, nil
}

/* func (appMesh *AppMesh) LoadNamespaceToMesh(namespace string, del bool) error {
	kclient := appMesh.KubeClient
	if kclient == nil {
		return ErrNilClient
	}

	// updating the label on the namespace
	ns, err := kclient.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// updating the annotations on the namespace
	if ns.ObjectMeta.Annotations == nil {
		ns.ObjectMeta.Annotations = map[string]string{}
	}
	ns.ObjectMeta.Annotations["appmesh.k8s.aws/sidecarInjectorWebhook"] = "enabled"

	if del {
		delete(ns.ObjectMeta.Annotations, "appmesh.k8s.aws/sidecarInjectorWebhook")
	}

	_, err = kclient.CoreV1().Namespaces().Update(context.TODO(), ns, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}
*/

// LoadNamespaceToMesh enables sidecar injection on by labelling requested
// namespace
func (appMesh *AppMesh) LoadNamespaceToMesh(namespace string, remove bool) error {
	ns, err := appMesh.KubeClient.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err != nil {
		return ErrLoadNamespaceToMesh(err)
	}

	if ns.ObjectMeta.Labels == nil {
		ns.ObjectMeta.Labels = map[string]string{}
	}
	//appmesh.k8s.aws/sidecarInjectorWebhook
	ns.ObjectMeta.Labels["appmesh.k8s.aws/sidecarInjectorWebhook"] = "enabled"

	if remove {
		ns.ObjectMeta.Labels["appmesh.k8s.aws/sidecarInjectorWebhook"] = "disabled"
	}

	_, err = appMesh.KubeClient.CoreV1().Namespaces().Update(context.TODO(), ns, metav1.UpdateOptions{})
	if err != nil {
		return ErrLoadNamespaceToMesh(err)
	}
	return nil
}
