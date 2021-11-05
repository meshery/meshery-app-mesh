package appmesh

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/status"

	mesherykube "github.com/layer5io/meshkit/utils/kubernetes"
)

const (
	repo              = "https://aws.github.io/eks-charts"
	appMeshController = "appmesh-controller"
	appMeshInject     = "appmesh-inject"
	appMeshGateway    = "appmesh-gateway"
)

// Installs APP-MESH service mesh using helm charts.
func (appMesh *AppMesh) installAppMesh(del bool, version, namespace string) (string, error) {
	appMesh.Log.Debug(fmt.Sprintf("Requested install of version: %s", version))
	appMesh.Log.Debug(fmt.Sprintf("Requested action is delete: %v", del))
	appMesh.Log.Debug(fmt.Sprintf("Requested action is in namespace: %s", namespace))

	appMesh.Log.Info(fmt.Sprintf("Requested install of version: %s", version))
	st := status.Installing
	if del {
		st = status.Removing
	}

	err := appMesh.Config.GetObject(adapter.MeshSpecKey, appMesh)
	if err != nil {
		return st, ErrMeshConfig(err)
	}

	err = appMesh.applyHelmChart(del, version, namespace)
	if err != nil {
		appMesh.Log.Error(ErrInstallAppMesh(err))
		return st, ErrInstallAppMesh(err)
	}

	if del {
		return status.Removed, nil
	}
	return status.Installed, nil
}

func (appMesh *AppMesh) applyHelmChart(del bool, version, namespace string) error {
	kClient := appMesh.MesheryKubeclient
	if kClient == nil {
		return ErrNilClient
	}

	appMesh.Log.Info("Installing using helm charts...")
	var act mesherykube.HelmChartAction
	if del {
		act = mesherykube.UNINSTALL
	} else {
		act = mesherykube.INSTALL
	}

	// Install the controller
	err := kClient.ApplyHelmChart(mesherykube.ApplyHelmChartConfig{
		ChartLocation: mesherykube.HelmChartLocation{
			Repository: repo,
			Chart:      appMeshController,
			AppVersion: version,
		},
		Namespace:       namespace,
		Action:          act,
		CreateNamespace: true,
	})
	if err != nil {
		return ErrApplyHelmChart(err)
	}

	// Install appmesh-injector. Only needed for controller versions older
	// than 1.0.0
	if controlPlaneVersion, err := strconv.Atoi(strings.TrimPrefix(version, "v")); controlPlaneVersion < 1 && err != nil {
		err = kClient.ApplyHelmChart(mesherykube.ApplyHelmChartConfig{
			ChartLocation: mesherykube.HelmChartLocation{
				Repository: repo,
				Chart:      appMeshInject,
				AppVersion: version,
			},
			Namespace:       namespace,
			Action:          act,
			CreateNamespace: true,
		})
		if err != nil {
			return ErrApplyHelmChart(err)
		}
	}

	return err
}

func (appMesh *AppMesh) applyManifest(manifest []byte, isDel bool, namespace string) error {
	err := appMesh.MesheryKubeclient.ApplyManifest(manifest, mesherykube.ApplyOptions{
		Namespace: namespace,
		Update:    true,
		Delete:    isDel,
	})
	if err != nil {
		return err
	}

	return nil
}
