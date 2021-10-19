package appmesh

import (
	"github.com/layer5io/meshery-adapter-library/status"
)

func (appMesh *AppMesh) applyCustomOperation(namespace string, manifest string, isDel bool) (string, error) {
	st := status.Starting

	err := appMesh.applyManifest([]byte(manifest), isDel, namespace)
	if err != nil {
		return st, ErrCustomOperation(err)
	}

	return status.Completed, nil
}
