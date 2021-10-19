package appmesh

import (
	"context"
	"fmt"

	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/common"
	"github.com/layer5io/meshery-adapter-library/status"
	internalconfig "github.com/layer5io/meshery-app-mesh/internal/config"
	meshkitCfg "github.com/layer5io/meshkit/config"
	"github.com/layer5io/meshkit/logger"
)

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

func (appMesh *AppMesh) ApplyOperation(ctx context.Context, opReq adapter.OperationRequest) error {

	operations := make(adapter.Operations)
	err := appMesh.Config.GetObject(adapter.OperationsKey, &operations)
	if err != nil {
		return err
	}

	e := &adapter.Event{
		Operationid: opReq.OperationID,
		Summary:     status.Deploying,
		Details:     "Operation is not supported",
	}
	stat := status.Deploying

	switch opReq.OperationName {
	case internalconfig.AppMeshOperation:
		go func(hh *AppMesh, ee *adapter.Event) {
			version := string(operations[opReq.OperationName].Versions[0])
			if stat, err = hh.installAppMesh(opReq.IsDeleteOperation, version, opReq.Namespace); err != nil {
				e.Summary = fmt.Sprintf("Error while %s AWS App mesh", stat)
				e.Details = err.Error()
				hh.StreamErr(e, err)
				return
			}
			ee.Summary = fmt.Sprintf("App mesh %s successfully", stat)
			ee.Details = fmt.Sprintf("The App mesh is now %s.", stat)
			hh.StreamInfo(e)
		}(appMesh, e)

	case common.SmiConformanceOperation:
		go func(hh *AppMesh, ee *adapter.Event) {
			name := operations[opReq.OperationName].Description
			_, err := hh.RunSMITest(adapter.SMITestOptions{
				Ctx:         context.TODO(),
				OperationID: ee.Operationid,
				Manifest:    string(operations[opReq.OperationName].Templates[0]),
				Namespace:   "meshery",
				Labels:      make(map[string]string),
				Annotations: map[string]string{
					"appmesh.k8s.aws/sidecarInjectorWebhook": "enabled",
				},
			})

			if err != nil {
				e.Summary = fmt.Sprintf("Error while %s %s test", status.Running, name)
				e.Details = err.Error()
				hh.StreamErr(e, err)
				return
			}
			ee.Summary = fmt.Sprintf("%s test %s successfully", name, status.Completed)
			ee.Details = ""
			hh.StreamInfo(e)
		}(appMesh, e)
	case internalconfig.LabelNamespace:
		go func(hh *AppMesh, ee *adapter.Event) {
			err := hh.LoadNamespaceToMesh(opReq.Namespace, opReq.IsDeleteOperation)
			operation := "enabled"
			if opReq.IsDeleteOperation {
				operation = "removed"
			}
			if err != nil {
				e.Summary = fmt.Sprintf("Error while labelling %s", opReq.Namespace)
				e.Details = err.Error()
				hh.StreamErr(e, err)
				return
			}
			ee.Summary = fmt.Sprintf("Label updated on %s namespace", opReq.Namespace)
			ee.Details = fmt.Sprintf("APP-MESH-INJECTION label %s on %s namespace", operation, opReq.Namespace)
			hh.StreamInfo(e)
		}(appMesh, e)
	case common.BookInfoOperation, common.HTTPBinOperation, common.ImageHubOperation, common.EmojiVotoOperation:
		go func(hh *AppMesh, ee *adapter.Event) {
			appName := operations[opReq.OperationName].AdditionalProperties[common.ServiceName]
			stat, err := hh.installSampleApp(opReq.Namespace, opReq.IsDeleteOperation, operations[opReq.OperationName].Templates)
			if err != nil {
				e.Summary = fmt.Sprintf("Error while %s %s application", stat, appName)
				e.Details = err.Error()
				hh.StreamErr(e, err)
				return
			}
			ee.Summary = fmt.Sprintf("%s application %s successfully", appName, stat)
			ee.Details = fmt.Sprintf("The %s application is now %s.", appName, stat)
			hh.StreamInfo(e)
		}(appMesh, e)
	case common.CustomOperation:
		go func(hh *AppMesh, ee *adapter.Event) {
			stat, err := hh.applyCustomOperation(opReq.Namespace, opReq.CustomBody, opReq.IsDeleteOperation)
			if err != nil {
				e.Summary = fmt.Sprintf("Error while %s custom operation", stat)
				e.Details = err.Error()
				hh.StreamErr(e, err)
				return
			}
			ee.Summary = fmt.Sprintf("Manifest %s successfully", status.Deployed)
			ee.Details = ""
			hh.StreamInfo(e)
		}(appMesh, e)
	default:
		appMesh.StreamErr(e, ErrOpInvalid)

	}
	return nil
}
