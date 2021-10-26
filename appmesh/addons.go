package appmesh

import (
	"context"
	"net/url"

	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/status"
	"github.com/layer5io/meshkit/utils"
	kubernetes "github.com/layer5io/meshkit/utils/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (appMesh *AppMesh) installAddon(ns string, del bool, svcName string, patches []string, hc adapter.HelmConfig) (string, error) {
	st := status.Installing

	if del {
		st = status.Removing
	}
	err := appMesh.MesheryKubeclient.ApplyHelmChart(kubernetes.ApplyHelmChartConfig{
		URL:       hc.URL,
		Namespace: ns,
	})
	if err != nil {
		return st, err //CHANGE TO MESHKIT ERRROR
	}

	for _, patch := range patches {
		if !del {
			_, err := url.ParseRequestURI(patch)
			if err != nil {
				return st, ErrAddonFromTemplate(err)
			}

			content, err := utils.ReadFileSource(patch)
			if err != nil {
				return st, ErrAddonFromTemplate(err)
			}

			_, err = appMesh.KubeClient.CoreV1().Services(ns).Patch(context.TODO(), svcName, types.MergePatchType, []byte(content), metav1.PatchOptions{})
			if err != nil {
				return st, ErrAddonFromTemplate(err)
			}
		}
	}
	return st, nil
}
