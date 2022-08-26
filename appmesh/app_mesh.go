// Package appmesh stores common operations
package appmesh

import (
	"context"
	"fmt"

	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/common"
	"github.com/layer5io/meshery-adapter-library/meshes"
	"github.com/layer5io/meshery-adapter-library/status"
	"github.com/layer5io/meshery-app-mesh/appmesh/oam"
	internalconfig "github.com/layer5io/meshery-app-mesh/internal/config"
	meshkitCfg "github.com/layer5io/meshkit/config"
	"github.com/layer5io/meshkit/errors"
	"github.com/layer5io/meshkit/logger"
	"github.com/layer5io/meshkit/models"
	"github.com/layer5io/meshkit/models/oam/core/v1alpha1"
	"gopkg.in/yaml.v2"
)

// AppMesh is the app-mesh adapter. It embeds adapter.Adapter
type AppMesh struct {
	adapter.Adapter
}

// New initializes AppMesh handler.
func New(c meshkitCfg.Handler, l logger.Handler, kc meshkitCfg.Handler) adapter.Handler {
	return &AppMesh{
		Adapter: adapter.Adapter{
			Config:            c,
			Log:               l,
			KubeconfigHandler: kc,
		},
	}
}

// ApplyOperation applies the requested operation on app-mesh
func (appMesh *AppMesh) ApplyOperation(ctx context.Context, opReq adapter.OperationRequest, hchan *chan interface{}) error {
	err := appMesh.CreateKubeconfigs(opReq.K8sConfigs)
	if err != nil {
		return err
	}
	kubeConfigs := opReq.K8sConfigs
	appMesh.SetChannel(hchan);

	operations := make(adapter.Operations)
	err = appMesh.Config.GetObject(adapter.OperationsKey, &operations)
	if err != nil {
		return err
	}

	e := &meshes.EventsResponse{
		OperationId: opReq.OperationID,
		Summary:     status.Deploying,
		Details:     "Operation is not supported",
		Component:   internalconfig.ServerConfig["type"],
		ComponentName: internalconfig.ServerConfig["name"],
	}
	stat := status.Deploying

	switch opReq.OperationName {
	case internalconfig.AppMeshOperation:
		go func(hh *AppMesh, ee *meshes.EventsResponse) {
			version := string(operations[opReq.OperationName].Versions[0])
			if stat, err = hh.installAppMesh(opReq.IsDeleteOperation, version, opReq.Namespace, kubeConfigs); err != nil {
				summary := fmt.Sprintf("Error while %s AWS App mesh", stat)
				hh.streamErr(summary, e, err)
				return
			}
			ee.Summary = fmt.Sprintf("App mesh %s successfully", stat)
			ee.Details = fmt.Sprintf("The App mesh is now %s.", stat)
			hh.StreamInfo(e)
		}(appMesh, e)

	case internalconfig.LabelNamespace:
		go func(hh *AppMesh, ee *meshes.EventsResponse) {
			err := hh.LoadNamespaceToMesh(opReq.Namespace, opReq.IsDeleteOperation, kubeConfigs)
			operation := "enabled"
			if opReq.IsDeleteOperation {
				operation = "removed"
			}
			if err != nil {
				summary := fmt.Sprintf("Error while labelling %s", opReq.Namespace)
				hh.streamErr(summary, e, err)
				return
			}
			ee.Summary = fmt.Sprintf("Label updated on %s namespace", opReq.Namespace)
			ee.Details = fmt.Sprintf("APP-MESH-INJECTION label %s on %s namespace", operation, opReq.Namespace)
			hh.StreamInfo(e)
		}(appMesh, e)
	case internalconfig.PrometheusAddon, internalconfig.GrafanaAddon:
		go func(hh *AppMesh, ee *meshes.EventsResponse) {
			svcname := operations[opReq.OperationName].AdditionalProperties[common.ServiceName]
			helmChartURL := operations[opReq.OperationName].AdditionalProperties[internalconfig.HelmChartURL]
			patches := make([]string, 0)
			patches = append(patches, operations[opReq.OperationName].AdditionalProperties[internalconfig.ServicePatchFile])
			_, err := hh.installAddon(opReq.Namespace, opReq.IsDeleteOperation, svcname, patches, helmChartURL, kubeConfigs)
			operation := "install"
			if opReq.IsDeleteOperation {
				operation = "uninstall"
			}

			if err != nil {
				summary := fmt.Sprintf("Error while %sing %s", operation, opReq.OperationName)
				hh.streamErr(summary, e, err)
				return
			}
			ee.Summary = fmt.Sprintf("Succesfully %sed %s", operation, opReq.OperationName)
			ee.Details = fmt.Sprintf("Succesfully %sed %s from the %s namespace", operation, opReq.OperationName, opReq.Namespace)
			hh.StreamInfo(e)
		}(appMesh, e)
	case common.BookInfoOperation, common.HTTPBinOperation, common.ImageHubOperation, common.EmojiVotoOperation:
		go func(hh *AppMesh, ee *meshes.EventsResponse) {
			appName := operations[opReq.OperationName].AdditionalProperties[common.ServiceName]
			stat, err := hh.installSampleApp(opReq.Namespace, opReq.IsDeleteOperation, operations[opReq.OperationName].Templates, kubeConfigs)
			if err != nil {
				summary := fmt.Sprintf("Error while %s %s application", stat, appName)
				hh.streamErr(summary, e, err)
				return
			}
			ee.Summary = fmt.Sprintf("%s application %s successfully", appName, stat)
			ee.Details = fmt.Sprintf("The %s application is now %s.", appName, stat)
			hh.StreamInfo(e)
		}(appMesh, e)

	case common.CustomOperation:
		go func(hh *AppMesh, ee *meshes.EventsResponse) {
			stat, err := hh.applyCustomOperation(opReq.Namespace, opReq.CustomBody, opReq.IsDeleteOperation, kubeConfigs)
			if err != nil {
				summary := fmt.Sprintf("Error while %s custom operation", stat)
				hh.streamErr(summary, e, err)
				return
			}
			ee.Summary = fmt.Sprintf("Manifest %s successfully", status.Deployed)
			ee.Details = ""
			hh.StreamInfo(e)
		}(appMesh, e)
	default:
		appMesh.streamErr("Invalid operation", e, ErrOpInvalid)

	}
	return nil
}

//CreateKubeconfigs creates and writes passed kubeconfig onto the filesystem
func (appMesh *AppMesh) CreateKubeconfigs(kubeconfigs []string) error {
	var errs = make([]error, 0)
	for _, kubeconfig := range kubeconfigs {
		kconfig := models.Kubeconfig{}
		err := yaml.Unmarshal([]byte(kubeconfig), &kconfig)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		// To have control over what exactly to take in on kubeconfig
		appMesh.KubeconfigHandler.SetKey("kind", kconfig.Kind)
		appMesh.KubeconfigHandler.SetKey("apiVersion", kconfig.APIVersion)
		appMesh.KubeconfigHandler.SetKey("current-context", kconfig.CurrentContext)
		err = appMesh.KubeconfigHandler.SetObject("preferences", kconfig.Preferences)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		err = appMesh.KubeconfigHandler.SetObject("clusters", kconfig.Clusters)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		err = appMesh.KubeconfigHandler.SetObject("users", kconfig.Users)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		err = appMesh.KubeconfigHandler.SetObject("contexts", kconfig.Contexts)
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return mergeErrors(errs)
}

// ProcessOAM handles the grpc invocation for handling OAM objects
func (appMesh *AppMesh) ProcessOAM(ctx context.Context, oamReq adapter.OAMRequest, hchan *chan interface{}) (string, error) {
	appMesh.SetChannel(hchan)
	err := appMesh.CreateKubeconfigs(oamReq.K8sConfigs)
	if err != nil {
		return "", err
	}
	kubeConfigs := oamReq.K8sConfigs

	var comps []v1alpha1.Component
	for _, acomp := range oamReq.OamComps {
		comp, err := oam.ParseApplicationComponent(acomp)
		if err != nil {
			appMesh.Log.Error(ErrParseOAMComponent)
			continue
		}

		comps = append(comps, comp)
	}

	config, err := oam.ParseApplicationConfiguration(oamReq.OamConfig)
	if err != nil {
		appMesh.Log.Error(ErrParseOAMConfig)
	}

	// If operation is delete then first HandleConfiguration and then handle the deployment
	if oamReq.DeleteOp {
		// Process configuration
		msg2, err := appMesh.HandleApplicationConfiguration(config, oamReq.DeleteOp, kubeConfigs)
		if err != nil {
			return msg2, ErrProcessOAM(err)
		}

		// Process components
		msg1, err := appMesh.HandleComponents(comps, oamReq.DeleteOp, kubeConfigs)
		if err != nil {
			return msg1 + "\n" + msg2, ErrProcessOAM(err)
		}

		return msg1 + "\n" + msg2, nil
	}

	// Process components
	msg1, err := appMesh.HandleComponents(comps, oamReq.DeleteOp, kubeConfigs)
	if err != nil {
		return msg1, ErrProcessOAM(err)
	}

	// Process configuration
	msg2, err := appMesh.HandleApplicationConfiguration(config, oamReq.DeleteOp, kubeConfigs)
	if err != nil {
		return msg1 + "\n" + msg2, ErrProcessOAM(err)
	}

	return msg1 + "\n" + msg2, nil
}

func(appMesh *AppMesh) streamErr(summary string, e *meshes.EventsResponse, err error) {
	e.Summary = summary
	e.Details = err.Error()
	e.ErrorCode = errors.GetCode(err)
	e.ProbableCause = errors.GetCause(err)
	e.SuggestedRemediation = errors.GetRemedy(err)
	appMesh.StreamErr(e, err)
}