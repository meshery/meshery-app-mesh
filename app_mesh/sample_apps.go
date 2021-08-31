package app_mesh

import (
	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/status"
)

func (appMesh *AppMesh) installSampleApp(namespace string, del bool, templates []adapter.Template) (string, error) {
	st := status.Installing

	if del {
		st = status.Removing
	}

	for _, template := range templates {
		err := appMesh.applyManifest([]byte(template.String()), del, namespace)
		if err != nil {
			return st, ErrSampleApp(err, st)
		}
	}

	return status.Installed, nil
}
