module github.com/layer5io/meshery-app-mesh

go 1.16

replace github.com/kudobuilder/kuttl => github.com/layer5io/kuttl v0.4.1-0.20200723152044-916f10574334

require (
	github.com/layer5io/meshery-adapter-library v0.1.25
	github.com/layer5io/meshkit v0.2.36
	github.com/layer5io/service-mesh-performance v0.3.3
	golang.org/x/sys v0.0.0-20210903071746-97244b99971b // indirect
	google.golang.org/genproto v0.0.0-20210903162649-d08c68adba83 // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/apimachinery v0.23.4
)
