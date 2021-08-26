// Copyright 2019 Layer5.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//Package app-mesh
package app

type supportedOperation struct {
	// a friendly name
	name string
	// the template file name
	templateName string
}

const (
	customOpCommand         = "custom"
	runVet                  = "app-mesh_vet"
	installapp-meshCommand     = "app-mesh_install"
	installmTLSapp-meshCommand = "app-mesh_mtls_install"
	installBookInfoCommand  = "install_book_info"
	cbCommand               = "cb1"
	installSMI              = "install_smi"
	installHTTPBin          = "install_http_bin"
)

var supportedOps = map[string]supportedOperation{
	installapp-meshCommand: {
		name: "Install the latest version of app-mesh",
		// templateName: "install_app-mesh.tmpl",
	},
	installmTLSapp-meshCommand: {
		name: "Install the latest version of app-mesh with mTLS",
	},
	installBookInfoCommand: {
		name: "Install the canonical Book Info Application",
		// templateName: "install_app-mesh.tmpl",
	},
	runVet: {
		name: "Run app-mesh-vet",
		// templateName: "app-mesh_vet.tmpl",
		// appLabel:     "app-mesh-vet",
		// returnLogs:   true,
	},
	cbCommand: {
		name:         "Limit circuit breaker config to one connection",
		templateName: "circuit_breaking.tmpl",
	},
	installSMI: {
		name: "Install Service Mesh Interface (SMI) app-mesh Adapter",
	},
	customOpCommand: {
		name: "Custom YAML",
	},
	installHTTPBin: {
		name:         "Install HTTP Bin app",
		templateName: "httpbin.yaml",
	},
}
