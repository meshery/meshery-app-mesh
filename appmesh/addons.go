package appmesh

import (
	"context"
	"net/url"
	"sync"

	"github.com/layer5io/meshery-adapter-library/status"
	"github.com/layer5io/meshkit/utils"
	kubernetes "github.com/layer5io/meshkit/utils/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	mesherykube "github.com/layer5io/meshkit/utils/kubernetes"
)

func (appMesh *AppMesh) installAddon(ns string, del bool, svcName string, patches []string, helmChartURL string, kubeconfigs []string) (string, error) {

	st := status.Installing

	if del {
		st = status.Removing
	}

	var wg sync.WaitGroup
	var errs []error
	var errMx sync.Mutex


	for _, config := range kubeconfigs {
		wg.Add(1);
		go func(config string) {
			defer wg.Done()
			kClient, err := mesherykube.New([]byte(config))
			if err != nil {
				errMx.Lock()
				errs = append(errs, err)
				errMx.Unlock()
				return
			}

			err = kClient.ApplyHelmChart(kubernetes.ApplyHelmChartConfig{
				URL:       helmChartURL,
				Namespace: ns,
			})
			if err != nil {
				errMx.Lock()
				errs = append(errs, err)
				errMx.Unlock()
				return
			}

			for _, patch := range patches {
				if !del {
					_, err := url.ParseRequestURI(patch)
					if err != nil {
						errMx.Lock()
						errs = append(errs, err)
						errMx.Unlock()
						return
					}

					content, err := utils.ReadFileSource(patch)
					if err != nil {
						errMx.Lock()
						errs = append(errs, err)
						errMx.Unlock()
						return
					}

					_, err = kClient.KubeClient.CoreV1().Services(ns).Patch(context.TODO(), svcName, types.MergePatchType, []byte(content), metav1.PatchOptions{})
					if err != nil {
						errMx.Lock()
						errs = append(errs, err)
						errMx.Unlock()
						return
					}
				}
			}
		}(config)
	}

	wg.Wait()
	if len(errs) == 0 {
		return st, nil
	}

	return st, ErrAddonFromHelm(mergeErrors(errs))
}
