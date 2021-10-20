package config

import (
	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/meshes"
)

const (
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
		Templates: []adapter.Template{
			"appmesh-prometheus",
		},
		AdditionalProperties: map[string]string{
			ServiceName:      "prometheus",
			"ServicePatchFile": "file://templates/patches/service-loadbalancer.json",
		},
	}

	dev[GrafanaAddon] = &adapter.Operation{
		Type:        int32(meshes.OpCategory_CONFIGURE),
		Description: "Add-on: Grafana",
		Templates: []adapter.Template{
			"appmesh-grafana",
		},
		AdditionalProperties: map[string]string{
			ServiceName:      "grafana",
			"ServicePatchFile": "file://templates/patches/service-loadbalancer.json",
		},
	}

	dev[JaegerAddon] = &adapter.Operation{
		Type:        int32(meshes.OpCategory_CONFIGURE),
		Description: "Add-on: Jaeger",
		Templates: []adapter.Template{
			"appmesh-jaeger",
		},
		AdditionalProperties: map[string]string{
			ServiceName:      "jaeger-collector",
			"ServicePatchFile": "file://templates/patches/service-loadbalancer.json",
		},
	}

	dev[SpireAgent] = &adapter.Operation{
		Type:        int32(meshes.OpCategory_CONFIGURE),
		Description: "Add-on: Spire-agent",
		Templates: []adapter.Template{
			"appmesh-spire-agent",
		},
	}

	dev[SpireServer] = &adapter.Operation{
		Type:        int32(meshes.OpCategory_CONFIGURE),
		Description: "Add-on: Spire-agent",
		Templates: []adapter.Template{
			"appmesh-spire-server",
		},
	}

	return dev
}
