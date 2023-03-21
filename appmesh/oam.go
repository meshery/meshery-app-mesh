package appmesh

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/layer5io/meshery-adapter-library/meshes"
	"github.com/layer5io/meshery-app-mesh/internal/config"
	"github.com/layer5io/meshkit/models/oam/core/v1alpha1"
	"gopkg.in/yaml.v2"
)

// CompHandler type functions handle OAM components
type CompHandler func(*AppMesh, v1alpha1.Component, bool, []string) (string, error)

// HandleComponents handles the processing of OAM components
func (appMesh *AppMesh) HandleComponents(comps []v1alpha1.Component, isDel bool, kubeconfigs []string) (string, error) {
	var errs []error
	var msgs []string
	stat1 := "deploying"
	stat2 := "deployed"
	if isDel {
		stat1 = "removing"
		stat2 = "removed"
	}
	compFuncMap := map[string]CompHandler{
		"AppMesh": handleComponentAppMesh,
	}

	for _, comp := range comps {
		ee := &meshes.EventsResponse{
			OperationId:   uuid.New().String(),
			Component:     config.ServerConfig["type"],
			ComponentName: config.ServerConfig["name"],
		}
		fnc, ok := compFuncMap[comp.Spec.Type]
		if !ok {
			msg, err := handleAppMeshCoreComponent(appMesh, comp, isDel, "", "", kubeconfigs)
			if err != nil {
				ee.Summary = fmt.Sprintf("Error while %s %s", stat1, comp.Spec.Type)
				appMesh.streamErr(ee.Summary, ee, err)
				errs = append(errs, err)
				continue
			}
			ee.Summary = fmt.Sprintf("%s %s successfully", comp.Spec.Type, stat2)
			ee.Details = fmt.Sprintf("The %s is now %s.", comp.Spec.Type, stat2)
			appMesh.StreamInfo(ee)

			msgs = append(msgs, msg)
			continue
		}

		msg, err := fnc(appMesh, comp, isDel, kubeconfigs)
		if err != nil {
			ee.Summary = fmt.Sprintf("Error while %s %s", stat1, comp.Spec.Type)
			appMesh.streamErr(ee.Summary, ee, err)
			errs = append(errs, err)
			continue
		}
		ee.Summary = fmt.Sprintf("%s %s %s successfully", comp.Name, comp.Spec.Type, stat2)
		ee.Details = fmt.Sprintf("The %s %s is now %s.", comp.Name, comp.Spec.Type, stat2)
		appMesh.StreamInfo(ee)

		msgs = append(msgs, msg)
	}

	if err := mergeErrors(errs); err != nil {
		return mergeMsgs(msgs), err
	}

	return mergeMsgs(msgs), nil
}

// HandleApplicationConfiguration handles the processing of OAM application configuration
func (appMesh *AppMesh) HandleApplicationConfiguration(config v1alpha1.Configuration, isDel bool, kubeconfigs []string) (string, error) {
	var errs []error
	var msgs []string
	for _, comp := range config.Spec.Components {
		for _, trait := range comp.Traits {
			if trait.Name == "automaticSidecarInjection.AppMesh" {
				namespaces := castSliceInterfaceToSliceString(trait.Properties["namespaces"].([]interface{}))
				if err := handleNamespaceLabel(appMesh, namespaces, isDel, kubeconfigs); err != nil {
					errs = append(errs, err)
				}
			}

			msgs = append(msgs, fmt.Sprintf("applied trait \"%s\" on service \"%s\"", trait.Name, comp.ComponentName))
		}
	}

	if err := mergeErrors(errs); err != nil {
		return mergeMsgs(msgs), err
	}

	return mergeMsgs(msgs), nil
}

func handleNamespaceLabel(appMesh *AppMesh, namespaces []string, isDel bool, kubeconfigs []string) error {
	var errs []error
	for _, ns := range namespaces {
		if err := appMesh.LoadNamespaceToMesh(ns, isDel, kubeconfigs); err != nil {
			errs = append(errs, err)
		}
	}

	return mergeErrors(errs)
}

func handleComponentAppMesh(appMesh *AppMesh, comp v1alpha1.Component, isDel bool, kubeconfigs []string) (string, error) {
	// Get the kuma version from the settings
	// we are sure that the version of kuma would be present
	// because the configuration is already validated against the schema
	version := comp.Spec.Version
	if version == "" {
		return "", fmt.Errorf("missing valid version inside service for AppMesh installation")
	}
	//TODO: When no version is passed in service, use the latest AppMesh version
	msg, err := appMesh.installAppMesh(isDel, version, comp.Namespace, kubeconfigs)
	if err != nil {
		return fmt.Sprintf("%s: %s", comp.Name, msg), err
	}

	return fmt.Sprintf("%s: %s", comp.Name, msg), nil
}

func handleAppMeshCoreComponent(
	appMesh *AppMesh,
	comp v1alpha1.Component,
	isDel bool,
	apiVersion,
	kind string,
	kubeconfigs []string) (string, error) {
	if apiVersion == "" {
		apiVersion = v1alpha1.GetAPIVersionFromComponent(comp)
		if apiVersion == "" {
			return "", ErrAppMeshCoreComponentFail(fmt.Errorf("failed to get API Version for: %s", comp.Name))
		}
	}

	if kind == "" {
		kind = v1alpha1.GetKindFromComponent(comp)
		if kind == "" {
			return "", ErrAppMeshCoreComponentFail(fmt.Errorf("failed to get kind for: %s", comp.Name))
		}
	}

	component := map[string]interface{}{
		"apiVersion": apiVersion,
		"kind":       kind,
		"metadata": map[string]interface{}{
			"name":        comp.Name,
			"annotations": comp.Annotations,
			"labels":      comp.Labels,
		},
		"spec": comp.Spec.Settings,
	}

	// Convert to yaml
	yamlByt, err := yaml.Marshal(component)
	if err != nil {
		err = ErrParseAppMeshCoreComponent(err)
		appMesh.Log.Error(err)
		return "", err
	}

	msg := fmt.Sprintf("created %s \"%s\" in namespace \"%s\"", kind, comp.Name, comp.Namespace)
	if isDel {
		msg = fmt.Sprintf("deleted %s config \"%s\" in namespace \"%s\"", kind, comp.Name, comp.Namespace)
	}

	return msg, appMesh.applyManifest(yamlByt, isDel, comp.Namespace, kubeconfigs)
}

func castSliceInterfaceToSliceString(in []interface{}) []string {
	var out []string

	for _, v := range in {
		cast, ok := v.(string)
		if ok {
			out = append(out, cast)
		}
	}

	return out
}

func mergeErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}

	var errMsgs []string

	for _, err := range errs {
		errMsgs = append(errMsgs, err.Error())
	}

	return fmt.Errorf(strings.Join(errMsgs, "\n"))
}

func mergeMsgs(strs []string) string {
	return strings.Join(strs, "\n")
}
