package config

import (
	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/meshes"
	"github.com/layer5io/meshkit/utils"
)

var (
	ServiceName = "service_name"
)

func getOperations(dev adapter.Operations) adapter.Operations {
	var adapterersions []adapter.Version
	versions, _ := utils.GetLatestReleaseTagsSorted("aws", "aws-app-mesh-controller-for-k8s")

	for _, v := range versions {
		adapterersions = append(adapterersions, adapter.Version(v))
	}
	dev[AppMeshOperation] = &adapter.Operation{
		Type:        int32(meshes.OpCategory_INSTALL),
		Description: "AWS App Mesh",
		Versions:    adapterersions,
	}

	dev[LabelNamespace] = &adapter.Operation{
		Type:        int32(meshes.OpCategory_CONFIGURE),
		Description: "Automatic Sidecar Injection",
	}
	dev[PrometheusAddon] = &adapter.Operation{
		Type:        int32(meshes.OpCategory_CONFIGURE),
		Description: "Add-on: Prometheus",
		AdditionalProperties: map[string]string{
			ServiceName:      "appmesh-prometheus",
			ServicePatchFile: "file://templates/patches/service-loadbalancer.json",
			HelmChartURL:     "https://aws.github.io/eks-charts/appmesh-prometheus-1.0.0.tgz",
		},
	}
	dev[GrafanaAddon] = &adapter.Operation{
		Type:        int32(meshes.OpCategory_CONFIGURE),
		Description: "Add-on: Grafana",
		AdditionalProperties: map[string]string{
			ServiceName:      "appmesh-grafana",
			ServicePatchFile: "file://templates/patches/service-loadbalancer.json",
			HelmChartURL:     "https://aws.github.io/eks-charts/appmesh-grafana-1.0.4.tgz",
		},
	}
	return dev
}
