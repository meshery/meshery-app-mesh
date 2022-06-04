package config

import (
	"path"
	"strings"

	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/common"
	"github.com/layer5io/meshery-adapter-library/config"
	configprovider "github.com/layer5io/meshkit/config/provider"
	"github.com/layer5io/meshery-adapter-library/status"

	//	"github.com/layer5io/meshery-adapter-library"
	"github.com/layer5io/meshkit/utils"
	smp "github.com/layer5io/service-mesh-performance/spec"
)

const (
	LabelNamespace   = "label-namespace"
	ServicePatchFile = "service-patch-file"
	// Addons that the adapter supports
	PrometheusAddon = "appmesh-prometheus-addon"
	GrafanaAddon    = "appmesh-grafana-addon"
	HelmChartURL    = "helm-chart-url"
	// OAM Metadata constants
	OAMAdapterNameMetadataKey       = "adapter.meshery.io/name"
	OAMComponentCategoryMetadataKey = "ui.meshery.io/category"
)

var (
	AppMeshOperation = strings.ToLower(smp.ServiceMesh_APP_MESH.Enum().String())

	ServerVersion  = status.None
	ServerGitSHA   = status.None
	configRootPath = path.Join(utils.GetHome(), ".meshery")

	Config = configprovider.Options{
		FilePath: configRootPath,
		FileName: "app-mesh",
		FileType: "yaml",
	}

	// ServerConfig is the configuration for the gRPC server
	ServerConfig = map[string]string{
		"name":     smp.ServiceMesh_APP_MESH.Enum().String(),
		"type":     "adapter",
		"port":     "10005",
		"traceurl": status.None,
	}

	// MeshSpec is the spec for the service mesh associated with this adapter
	MeshSpec = map[string]string{
		"name":    smp.ServiceMesh_APP_MESH.Enum().String(),
		"status":  status.NotInstalled,
		"version": status.None,
	}

	// ProviderConfig is the config for the configuration provider
	ProviderConfig = map[string]string{
		configprovider.FilePath: configRootPath,
		configprovider.FileType: "yaml",
		configprovider.FileName: "app-mesh",
	}

	// KubeConfig - Controlling the kubeconfig lifecycle with viper
	KubeConfig = map[string]string{
		configprovider.FilePath: configRootPath,
		configprovider.FileType: "yaml",
		configprovider.FileName: "kubeconfig",
	}

	// Operations represents the set of valid operations that are available
	// to the adapter
	Operations = getOperations(common.Operations)
)

// New creates a new config instance
func New(provider string) (h config.Handler, err error) {
	// Config provider
	switch provider {
	case configprovider.ViperKey:
		h, err = configprovider.NewViper(Config)
		if err != nil {
			return nil, err
		}
	case configprovider.InMemKey:
		h, err = configprovider.NewInMem(Config)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrEmptyConfig
	}

	// Setup server config
	if err := h.SetObject(adapter.ServerKey, ServerConfig); err != nil {
		return nil, err
	}

	// Setup mesh config
	if err := h.SetObject(adapter.MeshSpecKey, MeshSpec); err != nil {
		return nil, err
	}

	// Setup Operations Config
	if err := h.SetObject(adapter.OperationsKey, Operations); err != nil {
		return nil, err
	}

	return h, nil
}

func NewKubeconfigBuilder(provider string) (config.Handler, error) {
	opts := configprovider.Options{
		FilePath: configRootPath,
		FileType: "yaml",
		FileName: "kubeconfig",
	}

	// Config provider
	switch provider {
	case configprovider.ViperKey:
		return configprovider.NewViper(opts)
	case configprovider.InMemKey:
		return configprovider.NewInMem(opts)
	}
	return nil, ErrEmptyConfig
}

// RootPath returns the config root path for the adapter
func RootPath() string {
	return configRootPath
}
