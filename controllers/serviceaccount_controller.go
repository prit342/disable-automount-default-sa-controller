package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	serviceAccountPatch       = `{"automountServiceAccountToken": false}` // the patch we will apply
	defaultServiceAccountName = "default"                                 // the name of service account that we will patch with the above patch
)

// ServiceAccountReconciler - reconciles the default serviceAccount(s)
type ServiceAccountReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// ensure that we implement the reconcile.Reconciler interface
// so that it can be passed to the Builder
var _ reconcile.Reconciler = &ServiceAccountReconciler{}

// Reconcile - Reconciles a service account in a namespace
func (r *ServiceAccountReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	// set up a convenient log object, so we don't have to type request over and over again
	// Fetch the service account from our local cache populated by the shared informer
	sa := &corev1.ServiceAccount{}

	l := r.Log.WithValues("serviceAccount", request.NamespacedName)

	err := r.Client.Get(ctx, request.NamespacedName, sa)

	if err != nil && errors.IsNotFound(err) {
		l.Info("Could not find service account")
		return reconcile.Result{}, nil
	}

	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not fetch service account: %+v", err)
	}

	l.Info("Reconciling service account")

	if request.Name != defaultServiceAccountName {
		l.Info("Skip patching, as name does not match " + defaultServiceAccountName)
		return reconcile.Result{}, nil
	}

	if err := r.patchServiceAccount(ctx, sa, []byte(serviceAccountPatch)); err != nil {
		l.Error(err, "failed to patch service account")
		return reconcile.Result{}, err
	}

	l.Info("successfully patched the service account")
	return reconcile.Result{}, nil
}

// patchServiceAccount - patches a kubernetes service account with supplied patchData
func (r *ServiceAccountReconciler) patchServiceAccount(ctx context.Context, saObj *corev1.ServiceAccount, patchData []byte) error {

	l := r.Log.WithValues("name", saObj.Name, "namespace", saObj.Namespace)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in patchServiceAccount()", r)
		}
	}()

	// if automountServiceAccountToken is already set to false then we
	// do not need to patch it
	if saObj.AutomountServiceAccountToken != nil && !*saObj.AutomountServiceAccountToken {
		l.Info("skip patching as automountServiceAccountField is already set to false")
		return nil
	}

	saPatch := client.RawPatch(types.StrategicMergePatchType, patchData)

	if err := r.Client.Patch(ctx, saObj, saPatch); err != nil {
		l.Error(err, "failed to patch service account")
		return err
	}

	return nil
}
