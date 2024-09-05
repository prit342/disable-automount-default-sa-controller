package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	defaultServiceAccountName = "default" // the name of service account that we will patch
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
	//
	l := r.Log.WithValues("name", request.Name, "namespace", request.Namespace)
	l.Info("Reconciling service account")

	sa := &corev1.ServiceAccount{}

	if request.Name != defaultServiceAccountName {
		l.Info("Skip patching the service account, as name does not match to " + defaultServiceAccountName)
		return reconcile.Result{}, nil
	}

	// Fetch the service account from our local cache populated by the shared informer
	err := r.Client.Get(ctx, request.NamespacedName, sa)
	if err != nil {
		if apierrors.IsNotFound(err) {
			l.Info("ServiceAccount not found, it may not have been created yet or it was deleted")
			return reconcile.Result{}, nil // Don't requeue
		}
		l.Error(err, "Failed to get ServiceAccount")
		return reconcile.Result{}, err
	}

	if err := r.applyServiceAccountPatch(ctx, sa.Name, sa.Namespace); err != nil {
		l.Error(err, "failed to patch service account")
		return reconcile.Result{}, err
	}

	l.Info("successfully patched the service account")
	return reconcile.Result{}, nil
}

// applyServiceAccountPatch applies the desired state to the ServiceAccount using server-side apply
func (r *ServiceAccountReconciler) applyServiceAccountPatch(ctx context.Context, name, namespace string) error {

	l := log.FromContext(ctx)

	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		AutomountServiceAccountToken: pointer.Bool(false),
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in patchServiceAccount()", r)
		}
	}()

	// Apply the ServiceAccount using server-side apply
	err := r.Patch(ctx, sa, client.Apply, &client.PatchOptions{
		FieldManager: "disable-automount-default-sa-controller ",
		// force option allows the apply operation to overwrite fields that are managed by
		// other controllers or processes.
		Force: pointer.Bool(true),
	})

	if err != nil {
		l.Error(err, "Failed to apply service account patch")
		return err
	}

	l.Info("Successfully applied ServiceAccount patch")
	return nil
}
