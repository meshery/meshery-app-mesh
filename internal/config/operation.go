package config

import (
	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/meshes"
)

var (
	ServiceName = "service_name"
)

func getOperations(dev adapter.Operations) adapter.Operations {

	versions, _ := getLatestReleaseNames(3)

	dev[AppMeshOperation] = &adapter.Operation{
		Type:        int32(meshes.OpCategory_INSTALL),
		Description: "AWS App Mesh",
		Versions:    versions,
	}

	dev[LabelNamespace] = &adapter.Operation{
		Type:        int32(meshes.OpCategory_CONFIGURE),
		Description: "Automatic Sidecar Injection",
	}
	dev[PrometheusAddon] = &adapter.Operation{
		Type:        int32(meshes.OpCategory_CONFIGURE),
		Description: "Add-on: Prometheus",
		HelmConfig: adapter.HelmConfig{
			URL: "https://aws.github.io/eks-charts/appmesh-prometheus-1.0.0.tgz",
		},
		AdditionalProperties: map[string]string{
			ServiceName:      "prometheus",
			ServicePatchFile: "file://templates/patches/service-loadbalancer.json",
		},
	}
	return dev
}
