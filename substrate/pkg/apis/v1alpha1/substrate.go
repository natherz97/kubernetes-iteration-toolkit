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

// +k8s:deepcopy-gen=package,register
// +groupName=kit.sh
package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

type SubstrateSpec struct {
	// +optional
	VPC     *VPCSpec      `json:"vpc,omitempty"`
	Subnets []*SubnetSpec `json:"subnets,omitempty"`
	// +optional
	InstanceType *string `json:"instanceType,omitempty"`
}

// Substrate is the Schema for the Substrates API
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=substrates
// +kubebuilder:subresource:status
type Substrate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubstrateSpec   `json:"spec,omitempty"`
	Status SubstrateStatus `json:"status,omitempty"`
}

type VPCSpec struct {
	// TODO accept a slice of CIDR for megaXL we need to create multiple CIDRs
	CIDR string `json:"cidr,omitempty"`
}

type SubnetSpec struct {
	Zone   string
	CIDR   string
	Public bool
}

var (
	substrateConditionSet = apis.NewLivingConditionSet()
)

func (s *Substrate) IsReady() bool {
	return substrateConditionSet.Manage(&s.Status).GetCondition(apis.ConditionReady).IsTrue()
}

func (s *Substrate) Ready() {
	s.Status.SetConditions([]apis.Condition{{Type: apis.ConditionReady, Status: v1.ConditionTrue}})
}
