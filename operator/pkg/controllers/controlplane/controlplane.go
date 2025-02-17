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

package controlplane

import (
	"context"
	"fmt"

	"github.com/awslabs/kit/operator/pkg/apis/controlplane"
	"github.com/awslabs/kit/operator/pkg/apis/controlplane/v1alpha1"
	"github.com/awslabs/kit/operator/pkg/awsprovider"
	"github.com/awslabs/kit/operator/pkg/controllers"
	"github.com/awslabs/kit/operator/pkg/controllers/addons"
	"github.com/awslabs/kit/operator/pkg/controllers/etcd"
	"github.com/awslabs/kit/operator/pkg/controllers/master"
	"github.com/awslabs/kit/operator/pkg/kubeprovider"
	"github.com/awslabs/kit/operator/pkg/results"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type controlPlane struct {
	etcdController   *etcd.Controller
	masterController *master.Controller
	addonsController *addons.Controller
}

// NewController returns a controller for managing controlPlane components of the cluster
func NewController(kubeClient client.Client, account awsprovider.AccountMetadata, iamProvider controlplane.Controller) *controlPlane {
	return &controlPlane{
		etcdController:   etcd.New(kubeprovider.New(kubeClient)),
		masterController: master.New(kubeprovider.New(kubeClient), account, iamProvider),
		addonsController: addons.New(kubeprovider.New(kubeClient)),
	}
}

// Name returns the name of the controller
func (c *controlPlane) Name() string {
	return "control-plane"
}

// For returns the resource this controller is for.
func (c *controlPlane) For() controllers.Object {
	return &v1alpha1.ControlPlane{}
}

// Reconcile will reconcile all the components running on the control plane
func (c *controlPlane) Reconcile(ctx context.Context, object controllers.Object) (res *reconcile.Result, err error) {
	for _, resource := range []controlplane.Controller{
		c.etcdController,
		c.masterController,
		c.addonsController,
	} {
		if err := resource.Reconcile(ctx, object.(*v1alpha1.ControlPlane)); err != nil {
			return nil, fmt.Errorf("control plane reconciling, %w", err)
		}
	}
	return results.Created, nil
}

func (c *controlPlane) Finalize(ctx context.Context, object controllers.Object) (*reconcile.Result, error) {
	if err := c.masterController.Finalize(ctx, object.(*v1alpha1.ControlPlane)); err != nil {
		return results.Failed, err
	}
	return results.Terminated, nil
}
