package appmesh

import (
	"context"
	"fmt"
	"net/url"

	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/status"
	"github.com/layer5io/meshkit/utils"
	mesherykube "github.com/layer5io/meshkit/utils/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// installAddon installs/uninstalls an addon in the given namespace
//
// the template is the addon's helm chart's name which needs to be used to
// install the addon
func (appMesh *AppMesh) installAddon(namespace string, del bool, service string, patches []string, templates []adapter.Template) (string, error) {
	st := status.Installing

	if del {
		st = status.Removing
	}

	var act mesherykube.HelmChartAction
	if del {
		act = mesherykube.UNINSTALL
	} else {
		act = mesherykube.INSTALL
	}

	appMesh.Log.Debug(fmt.Sprintf("Overidden namespace: %s", namespace))
	namespace = "appmesh-system"

	for _, chartName := range templates {
		kClient := appMesh.MesheryKubeclient
		if kClient == nil {
			return st, ErrNilClient
		}

		err := kClient.ApplyHelmChart(mesherykube.ApplyHelmChartConfig{
			ChartLocation: mesherykube.HelmChartLocation{
				Repository: repo,
				Chart:      chartName.String(),
			},
			Namespace:       namespace,
			Action:          act,
			CreateNamespace: true,
		})
		if err != nil {
			return st, ErrApplyHelmChart(err)
		}
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

			_, err = appMesh.KubeClient.CoreV1().Services(namespace).Patch(context.TODO(), service, types.MergePatchType, []byte(content), metav1.PatchOptions{})
			if err != nil {
				return st, ErrAddonFromTemplate(err)
			}
		}
	}

	return status.Installed, nil
}
