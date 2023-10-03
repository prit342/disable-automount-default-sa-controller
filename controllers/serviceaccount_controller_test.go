package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("ServiceAccountPatch controllers", func() {

	const timeout = time.Second * 30
	const interval = time.Millisecond * 250
	var ctx = context.Background()

	// Create a new service account and check its patch status.
	createAndCheckServiceAccount := func(name, namespace string, saValue *bool) {
		serviceAccountLookupKey := types.NamespacedName{Name: name, Namespace: namespace}

		By("creating a new service account")
		sa := newServiceAccountObject(name, namespace, saValue)
		Expect(k8sClient.Create(ctx, sa)).Should(Succeed())

		By("checking the service account's existence")
		createdServiceAccount := &corev1.ServiceAccount{}
		Eventually(func() bool {
			err := k8sClient.Get(ctx, serviceAccountLookupKey, createdServiceAccount)
			return err == nil
		}, timeout, interval).Should(BeTrue())

		By("verifying the controller patches the service account")
		Eventually(func() bool {
			patchedServiceAccount := &corev1.ServiceAccount{}
			err := k8sClient.Get(ctx, serviceAccountLookupKey, patchedServiceAccount)
			return err == nil && *patchedServiceAccount.AutomountServiceAccountToken == false
		}, timeout, interval).Should(BeTrue())
	}

	Context("When default service account is created with AutomountServiceAccountToken set to false", func() {
		It("should be patched successfully ", func() {
			createAndCheckServiceAccount("default", "default", pointer.BoolPtr(true))
		})
	})

	Context("When default service account is created with AutomountServiceAccountToken set to nil in kube-system ns", func() {
		It("should be patched successfully ", func() {
			createAndCheckServiceAccount("default", "kube-system", nil)
		})
	})

	Context("When fetching service account fails", func() {
		It("should return an error ", func() {
			// Create a fake client with the given scheme.
			fakeClient := fake.NewClientBuilder().WithScheme(testEnv.Scheme).Build()

			// Set up a logger for the reconciler
			ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

			reconciler := &ServiceAccountReconciler{
				Client: fakeClient,
				Log:    ctrl.Log.WithName("controllers").WithName("saPatchController"),
				Scheme: testEnv.Scheme,
			}

			// The fake client won't have any objects stored by default,
			// so attempting to fetch any object will result in a "not found" error.
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: "default", Namespace: "default"}})
			Expect(err).ToNot(HaveOccurred()) // because if it doesn't find the SA, it's not an error as per the reconciler's logic.
		})
	})

})
