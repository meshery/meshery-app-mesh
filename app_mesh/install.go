package app_mesh

import (
	"fmt"

	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/status"

	mesherykube "github.com/layer5io/meshkit/utils/kubernetes"
)

const (
	repo  = "https://aws.github.io/eks-charts"
	chart = "appmesh-controller"
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
	err := kClient.ApplyHelmChart(mesherykube.ApplyHelmChartConfig{
		ChartLocation: mesherykube.HelmChartLocation{
			Repository: repo,
			Chart:      chart,
			AppVersion: version,
		},
		Namespace:       namespace,
		Delete:          del,
		CreateNamespace: true,
	})
	if err != nil {
		return ErrApplyHelmChart(err)
	}

	return nil
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
