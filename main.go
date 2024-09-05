package main

import (
	"context"
	"fmt"
	"os"

	"github.com/prit342/disable-automount-default-sa-controller/controllers"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	scheme                    = runtime.NewScheme()
	setupLog                  = ctrl.Log.WithName("setup")
	defaultServiceAccountName = `default`
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

func main() {

	ctrl.SetLogger(zap.New(zap.UseDevMode(false)))

	electionNamesapce := os.Getenv("CONTROLLER_NAMESPACE")

	if electionNamesapce == "" {
		setupLog.Error(fmt.Errorf("CONTROLLER_NAMESPACE environment variable is not set"), "initilisation failed")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), manager.Options{
		Metrics: metricsserver.Options{
			SecureServing: false,
			BindAddress:   "0.0.0.0:8080",
		},
		Scheme:                        scheme,
		LeaderElection:                true,
		LeaderElectionID:              "a0ea523r0e-default-sa-controller",
		LeaderElectionNamespace:       electionNamesapce,
		LeaderElectionReleaseOnCancel: true,
	})

	// if we failed to setup the manager
	if err != nil {
		setupLog.Error(err, "unable to set up the controllers manager")
		os.Exit(1)
	}

	r := &controllers.ServiceAccountReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("saPatchController"),
		Scheme: mgr.GetScheme(),
	}

	// isDefaultServiceAccount is a predicate function that
	// returns true if the object is a service account and is named "default"
	isDefaultServiceAccount := func(obj client.Object) bool {
		sa, ok := obj.(*corev1.ServiceAccount)
		if !ok {
			return false
		}
		return sa.Name == defaultServiceAccountName
	}

	// When a Namespace event occurs, this function returns a reconcile request for the default
	// ServiceAccount in that namespace.
	findDefaultServiceAccount := func(ctx context.Context, obj client.Object) []reconcile.Request {
		namespace, ok := obj.(*corev1.Namespace)
		if !ok {
			r.Log.Error(nil, "Expected a Namespace but got something else")
			return nil
		}

		return []reconcile.Request{
			{
				NamespacedName: types.NamespacedName{
					Name:      defaultServiceAccountName,
					Namespace: namespace.GetName(),
				},
			},
		}
	}
	// set up a new Controller that watches serviceAccount and reconciles them
	err = ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ServiceAccount{}, // watch serviceaccount resources
			// add a predicate to filter only default service accounts
			builder.WithPredicates(predicate.NewPredicateFuncs(isDefaultServiceAccount))).
		Watches(
			&corev1.Namespace{},
			handler.EnqueueRequestsFromMapFunc(findDefaultServiceAccount),
		).
		Complete(r)

	if err != nil {
		setupLog.Error(err, "failed to build a controller")
		os.Exit(1)
	}

	setupLog.Info("starting up the manager")

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "unable to start the manager ")
		os.Exit(1)
	}

}
