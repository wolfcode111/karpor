// Copyright The Karbour Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package insight

import (
	"context"
	"strings"

	"github.com/KusionStack/karbour/pkg/core/entity"
	"github.com/KusionStack/karbour/pkg/infra/multicluster"
	topologyutil "github.com/KusionStack/karbour/pkg/util/topology"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "sigs.k8s.io/yaml"
)

// GetResource returns the unstructured cluster object for a given cluster.
func (i *InsightManager) GetResource(
	ctx context.Context, client *multicluster.MultiClusterClient, resourceGroup *entity.ResourceGroup,
) (*unstructured.Unstructured, error) {
	resourceGVR, err := topologyutil.GetGVRFromGVK(resourceGroup.APIVersion, resourceGroup.Kind)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(resourceGroup.Kind, "Secret") {
		secret, err := client.DynamicClient.Resource(resourceGVR).Namespace(resourceGroup.Namespace).Get(ctx, resourceGroup.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		return i.SanitizeSecret(secret)
	}
	return client.DynamicClient.
		Resource(resourceGVR).
		Namespace(resourceGroup.Namespace).
		Get(ctx, resourceGroup.Name, metav1.GetOptions{})
}

// GetYAMLForResource returns the yaml byte array for a given cluster
func (i *InsightManager) GetYAMLForResource(
	ctx context.Context, client *multicluster.MultiClusterClient, resourceGroup *entity.ResourceGroup,
) ([]byte, error) {
	obj, err := i.GetResource(ctx, client, resourceGroup)
	if err != nil {
		return nil, err
	}
	return k8syaml.Marshal(obj.Object)
}

// SanitizeSecret redact the data field in the secret object
func (i *InsightManager) SanitizeSecret(original *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	sanitized := original
	if _, ok := sanitized.Object["data"]; ok {
		sanitized.Object["data"] = "[redacted]"
	}
	return original, nil
}
