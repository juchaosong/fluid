/*
Copyright 2022 The Fluid Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package jindofsx

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubectl"
	"gopkg.in/yaml.v2"
)

func (e *JindoFSxEngine) setupMasterInernal() (err error) {
	var (
		chartName = utils.GetChartsDirectory() + "/jindofsx"
	)
	valuefileName, err := e.generateJindoValueFile()
	if err != nil {
		return
	}
	found, err := helm.CheckRelease(e.name, e.namespace)
	if err != nil {
		return
	}
	if found {
		e.Log.Info("The release is already installed", "name", e.name, "namespace", e.namespace)
		return
	}

	return helm.InstallRelease(e.name, e.namespace, valuefileName, chartName)
}

func (e *JindoFSxEngine) generateJindoValueFile() (valueFileName string, err error) {
	// why need to delete configmap e.name+"-jindofs-config" ? Or it should be
	// err = kubeclient.DeleteConfigMap(e.Client, e.name+"-jindofs-config", e.namespace)
	err = kubeclient.DeleteConfigMap(e.Client, e.getConfigmapName(), e.namespace)
	if err != nil {
		e.Log.Error(err, "Failed to clean value files")
	}
	value, err := e.transform(e.runtime)
	if err != nil {
		return
	}
	data, err := yaml.Marshal(value)
	if err != nil {
		return
	}
	valueFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("%s-%s-values.yaml", e.name, e.runtimeType))
	if err != nil {
		e.Log.Error(err, "failed to create value file", "valueFile", valueFile.Name())
		return valueFileName, err
	}
	valueFileName = valueFile.Name()
	e.Log.V(1).Info("Save the values file", "valueFile", valueFileName)

	err = ioutil.WriteFile(valueFileName, data, 0400)
	if err != nil {
		return
	}

	err = kubectl.CreateConfigMapFromFile(e.getConfigmapName(), "data", valueFileName, e.namespace)
	if err != nil {
		return
	}
	return valueFileName, err
}

func (e *JindoFSxEngine) getConfigmapName() string {
	return e.name + "-" + e.runtimeType + "-values"
}