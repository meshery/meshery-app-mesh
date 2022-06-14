package main

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	//"github.com/aws/aws-sdk-go/service/appmesh"

	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/api/grpc"
	configprovider "github.com/layer5io/meshery-adapter-library/config/provider"
	"github.com/layer5io/meshery-app-mesh/appmesh"
	"github.com/layer5io/meshery-app-mesh/internal/config"
	"github.com/layer5io/meshkit/logger"
	"github.com/layer5io/meshkit/utils"
	"github.com/layer5io/meshkit/utils/manifests"

	// "github.com/layer5io/meshkit/tracing"
	"github.com/layer5io/meshery-app-mesh/appmesh/oam"
	smp "github.com/layer5io/service-mesh-performance/spec"
)

var (
	serviceName = "app-mesh-adapter"
	version     = "edge"
	gitsha      = "none"
)

func isDebug() bool {
	return os.Getenv("DEBUG") == "true"
}

// creates the ~/.meshery directory
func init() {
	err := os.MkdirAll(path.Join(config.RootPath(), "bin"), 0750)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}

// main is the entrypoint of the adaptor
func main() {

	// Initialize Logger instance
	log, err := logger.New(serviceName, logger.Options{
		Format:     logger.SyslogLogFormat,
		DebugLevel: isDebug(),
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = os.Setenv("KUBECONFIG", path.Join(
		config.KubeConfig[configprovider.FilePath],
		fmt.Sprintf("%s.%s", config.KubeConfig[configprovider.FileName], config.KubeConfig[configprovider.FileType])),
	)

	if err != nil {
		// Fail silently
		log.Warn(err)
	}

	// Initialize application specific configs and dependencies
	// App and request config
	cfg, err := config.New(configprovider.ViperKey)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	service := &grpc.Service{}
	err = cfg.GetObject(adapter.ServerKey, service)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	kubeconfigHandler, err := config.NewKubeconfigBuilder(configprovider.ViperKey)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	// // Initialize Tracing instance
	// tracer, err := tracing.New(service.Name, service.TraceURL)
	// if err != nil {
	//      log.Err("Tracing Init Failed", err.Error())
	//      os.Exit(1)
	// }

	// Initialize Handler intance
	handler := appmesh.New(cfg, log, kubeconfigHandler)
	handler = adapter.AddLogger(log, handler)

	service.Handler = handler
	service.Channel = make(chan interface{}, 10)
	service.StartedAt = time.Now()
	service.Version = version
	service.GitSHA = gitsha

	go registerCapabilities(service.Port, log)        //Registering static capabilities
	go registerDynamicCapabilities(service.Port, log) //Registering latest capabilities periodically

	// Server Initialization
	log.Info("Adaptor Listening at port: ", service.Port)
	err = grpc.Start(service, nil)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
func registerCapabilities(port string, log logger.Handler) {
	// Register workloads
	log.Info("Starting static component registration...")
	if err := oam.RegisterWorkloads(mesheryServerAddress(), serviceAddress()+":"+port); err != nil {
		log.Info(err.Error())
	}
	log.Info("Static Component registration completed")
	// // Register traits
	// if err := oam.RegisterTraits(mesheryServerAddress(), serviceAddress()+":"+port); err != nil {
	// 	log.Info(err.Error())
	// }
}

func registerDynamicCapabilities(port string, log logger.Handler) {
	registerWorkloads(port, log)
	//Start the ticker
	const reRegisterAfter = 24
	ticker := time.NewTicker(reRegisterAfter * time.Hour)
	for {
		<-ticker.C
		registerWorkloads(port, log)
	}

}
func registerWorkloads(port string, log logger.Handler) {
	versions, err := utils.GetLatestReleaseTagsSorted("aws", "aws-app-mesh-controller-for-k8s")
	if err != nil {
		log.Info("Could not get latest stable release")
	}
	if len(versions) == 0 {
		log.Info("Could not register dynamic components.Latest version could not found")
		return
	}
	version := versions[len(versions)-1]

	if oam.AvailableVersions[version] {
		log.Info("Latest(", version, ") component already available via static component generation\n")
		log.Info("Skipping dynamic component registeration")
		return
	}
	log.Info("Registering latest workload components for version ", version)
	// Register workloads
	if err := adapter.RegisterWorkLoadsDynamically(mesheryServerAddress(), serviceAddress()+":"+port, &adapter.DynamicComponentsConfig{
		TimeoutInMinutes: 30,
		URL:              "https://raw.githubusercontent.com/aws/eks-charts/master/stable/appmesh-controller/crds/crds.yaml",
		GenerationMethod: adapter.Manifests,
		Config: manifests.Config{
			Name:        smp.ServiceMesh_Type_name[int32(smp.ServiceMesh_NGINX_SERVICE_MESH)],
			MeshVersion: version,
			CrdFilter: manifests.NewCueCrdFilter(manifests.ExtractorPaths{
				NamePath:    "spec.names.kind",
				IdPath:      "spec.names.kind",
				VersionPath: "spec.versions[0].name",
				GroupPath:   "spec.group",
				SpecPath:    "spec.versions[0].schema.openAPIV3Schema.properties.spec"}, false),
			ExtractCrds: func(manifest string) []string {
				crds := strings.Split(manifest, "---")
				// trim the spaces
				for _, crd := range crds {
					crd = strings.TrimSpace(crd)
				}
				return crds
			},
		},
		Operation: config.AppMeshOperation,
	}); err != nil {
		log.Info(err.Error())
		return
	}
	log.Info("Latest workload components successfully registered.")
}

func mesheryServerAddress() string {
	meshReg := os.Getenv("MESHERY_SERVER")

	if meshReg != "" {
		if strings.HasPrefix(meshReg, "http") {
			return meshReg
		}

		return "http://" + meshReg
	}

	return "http://localhost:9081"
}

func serviceAddress() string {
	svcAddr := os.Getenv("SERVICE_ADDR")

	if svcAddr != "" {
		return svcAddr
	}

	return "mesherylocal.layer5.io"
}
