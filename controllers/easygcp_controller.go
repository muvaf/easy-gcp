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

package controllers

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types2 "k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/kustomize/api/types"

	"github.com/crossplaneio/easy-gcp/pkg/operations"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gcpv1alpha1 "github.com/crossplaneio/easy-gcp/api/v1alpha1"
)

const (
	shortWait = 30 * time.Second
	longWait  = 1 * time.Minute
)

// EasyGCPReconciler reconciles a EasyGCP object
type EasyGCPReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=gcp.easystacks.crossplane.io,resources=easygcps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gcp.easystacks.crossplane.io,resources=easygcps/status,verbs=get;update;patch

func (r *EasyGCPReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	cr := &gcpv1alpha1.EasyGCP{}
	if err := r.Client.Get(ctx, req.NamespacedName, cr); err != nil {
		return ctrl.Result{RequeueAfter: shortWait}, err
	}
	// todo: use WasDeleted
	if cr.DeletionTimestamp != nil {
		return ctrl.Result{Requeue: false}, nil
	}
	err := operations.ProcessKustomization(func(k *types.Kustomization) {
		k.NamePrefix = fmt.Sprintf("%s-", cr.GetName())
		k.CommonLabels = map[string]string{
			"gcp.configurationstacks.crossplane.io/name": cr.GetName(),
			"gcp.configurationstacks.crossplane.io/uid":  string(cr.GetUID()),
		}
	})
	objects, err := operations.RunKustomize()
	for _, o := range objects {
		if err := ApplyObject(ctx, r.Client, o); err != nil {
			return ctrl.Result{RequeueAfter: shortWait}, err
		}
		fmt.Printf("deployed %s type!", o.GetObjectKind().GroupVersionKind().String())
	}
	return ctrl.Result{RequeueAfter: longWait}, err
}

func ApplyObject(ctx context.Context, kube client.Client, o runtime.Object) error {
	type Object interface {
		runtime.Object
		v1.Object
	}
	resource, ok := o.(Object)
	if !ok {
		return fmt.Errorf("given object does not have metadata")
	}
	existing := resource.DeepCopyObject().(Object)
	err := kube.Get(ctx, types2.NamespacedName{Name: resource.GetName(), Namespace: resource.GetNamespace()}, existing)
	if errors.IsNotFound(err) {
		return kube.Create(ctx, resource)
	}
	if err != nil {
		return err
	}
	resource.SetResourceVersion(existing.GetResourceVersion())
	return kube.Patch(ctx, resource, client.MergeFrom(existing))
}

func (r *EasyGCPReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gcpv1alpha1.EasyGCP{}).
		Complete(r)
}
