package appmesh

import (
	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/status"
	kubernetes "github.com/layer5io/meshkit/utils/kubernetes"
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

	return "", nil
}	
