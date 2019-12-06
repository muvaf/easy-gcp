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

	"github.com/crossplaneio/stack-gcp/apis/v1alpha3"

	"k8s.io/apimachinery/pkg/api/errors"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types2 "k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/kustomize/api/types"

	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/easy-gcp/pkg/operations"
	"github.com/crossplaneio/easy-gcp/pkg/resource"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gcpv1alpha1 "github.com/crossplaneio/easy-gcp/api/v1alpha1"
)

const (
	shortWait = 30 * time.Second
	longWait  = 3 * time.Minute

	finalizer = "configurationstacks.crossplane.io"
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
	if meta.WasDeleted(cr) {
		return ctrl.Result{Requeue: false}, nil
	}
	err := operations.ProcessKustomization(func(k *types.Kustomization) {
		k.NamePrefix = fmt.Sprintf("%s-", cr.GetName())
		k.CommonLabels = map[string]string{
			fmt.Sprintf("%s/name", cr.GroupVersionKind().Group): cr.GetName(),
			fmt.Sprintf("%s/uid", cr.GroupVersionKind().Group):  string(cr.GetUID()),
		}
		for key, val := range cr.GetLabels() {
			k.CommonLabels[key] = val
		}
	})
	objects, err := operations.RunKustomize()
	for _, o := range objects {
		// NOTE(muvaf): Provider types get deleted immediately before managed resource
		// controllers get a chance to reconcile the deletion.
		if o.GetObjectKind().GroupVersionKind() != v1alpha3.ProviderGroupVersionKind {
			t := true
			meta.AddOwnerReference(o, v1.OwnerReference{
				APIVersion:         cr.APIVersion,
				Kind:               cr.Kind,
				Name:               cr.GetName(),
				UID:                cr.GetUID(),
				BlockOwnerDeletion: &t,
			})
		}
		if err := ApplyObject(ctx, r.Client, o); err != nil {
			return ctrl.Result{RequeueAfter: shortWait}, err
		}
	}
	if err := r.Client.Update(ctx, cr); err != nil {
		return ctrl.Result{RequeueAfter: shortWait}, err
	}
	return ctrl.Result{RequeueAfter: longWait}, err
}

func ApplyObject(ctx context.Context, kube client.Client, o resource.Resource) error {
	existing := o.DeepCopyObject().(resource.Resource)
	err := kube.Get(ctx, types2.NamespacedName{Name: o.GetName(), Namespace: o.GetNamespace()}, existing)
	if errors.IsNotFound(err) {
		return kube.Create(ctx, o)
	}
	if err != nil {
		return err
	}
	o.SetResourceVersion(existing.GetResourceVersion())
	return kube.Patch(ctx, o, client.MergeFrom(existing))
}

func (r *EasyGCPReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gcpv1alpha1.EasyGCP{}).
		Complete(r)
}
