/*
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

package master

import (
	"context"
	"fmt"

	"github.com/awslabs/kit/operator/pkg/apis/controlplane/v1alpha1"
	"github.com/awslabs/kit/operator/pkg/errors"
	"github.com/awslabs/kit/operator/pkg/utils/object"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *Controller) reconcileEndpoint(ctx context.Context, cp *v1alpha1.ControlPlane) (err error) {
	return c.kubeClient.EnsureCreate(ctx, object.WithOwner(cp, &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceNameFor(cp.ClusterName()),
			Namespace: cp.Namespace,
			Annotations: map[string]string{
				"service.beta.kubernetes.io/aws-load-balancer-scheme":                  "internet-facing",
				"service.beta.kubernetes.io/aws-load-balancer-type":                    "nlb-ip",
				"service.beta.kubernetes.io/aws-load-balancer-target-group-attributes": "stickiness.enabled=true,stickiness.type=source_ip",
			},
		},
		Spec: v1.ServiceSpec{
			Type:     v1.ServiceTypeLoadBalancer,
			Selector: APIServerLabels(cp.ClusterName()),
			Ports: []v1.ServicePort{{
				Port:       443,
				Name:       apiserverPortName(cp.ClusterName()),
				TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 443},
				Protocol:   "TCP",
			}},
		},
	}))
}

func (c *Controller) getClusterEndpoint(ctx context.Context, nn types.NamespacedName) (string, error) {
	return GetClusterEndpoint(ctx, c.kubeClient, nn)
}

func GetClusterEndpoint(ctx context.Context, client client.Client, nn types.NamespacedName) (string, error) {
	svc := &v1.Service{}
	if err := client.Get(ctx, types.NamespacedName{nn.Namespace, ServiceNameFor(nn.Name)}, svc); err != nil {
		if errors.IsNotFound(err) {
			return "", fmt.Errorf("getting control plane endpoint, %w", errors.WaitingForSubResources)
		}
		return "", fmt.Errorf("getting cluster endpoint, %w", err)
	}
	if len(svc.Status.LoadBalancer.Ingress) > 0 {
		return svc.Status.LoadBalancer.Ingress[0].Hostname, nil
	}
	return "", fmt.Errorf("endpoint name, %w", errors.WaitingForSubResources)
}

func apiserverPortName(clusterName string) string {
	return fmt.Sprintf("%s-port", ServiceNameFor(clusterName))
}

// service name length needs to be <63 for DNS names
// https://github.com/awslabs/kubernetes-iteration-toolkit/issues/70
func ServiceNameFor(clusterName string) string {
	return fmt.Sprintf("%s-cp", clusterName)
}

func labelsFor(clusterName string) map[string]string {
	return map[string]string{
		"app": ServiceNameFor(clusterName),
	}
}
