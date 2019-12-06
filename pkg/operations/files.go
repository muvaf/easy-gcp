/*
Copyright 2019 The Crossplane Authors.

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

package operations

import (
	"fmt"
	"io/ioutil"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"sigs.k8s.io/kustomize/api/filesys"

	"gopkg.in/yaml.v2"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"

	"github.com/crossplaneio/easy-gcp/pkg/resource"
)

const resourcesDirectory = "resources"

var kustomizationFilePath = fmt.Sprintf("%s/%s", resourcesDirectory, "kustomization.yaml")

func ProcessKustomization(process func(*types.Kustomization)) error {
	k := &types.Kustomization{}
	data, err := ioutil.ReadFile(kustomizationFilePath)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, k); err != nil {
		return err
	}
	process(k)
	yamlData, err := yaml.Marshal(k)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(kustomizationFilePath, yamlData, os.ModePerm)
}

func RunKustomize() ([]resource.Resource, error) {
	kustomizer := krusty.MakeKustomizer(filesys.MakeFsOnDisk(), krusty.MakeDefaultOptions())
	resMap, err := kustomizer.Run(resourcesDirectory)
	if err != nil {
		return nil, err
	}
	var objects []resource.Resource
	for _, res := range resMap.Resources() {
		u := &unstructured.Unstructured{}
		u.SetUnstructuredContent(res.Map())
		objects = append(objects, u)
	}
	return objects, nil
}
