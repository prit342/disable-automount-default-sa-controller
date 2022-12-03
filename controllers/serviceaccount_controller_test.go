package controllers

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("ServiceAccountPatch controllers", func() {
	// we first create a new service account with automountService account set to false and
	// check to see if its get patched by our controller

	timeout := time.Second * 30
	interval := time.Millisecond * 250

	Context("When default service account is created with AutomountServiceAccountToken set to false", func() {
		It("should be patched successfully ", func() {
			// name of the service account
			name := "default"
			// default namespace is already present in the envTest control plane
			namespace := "default"

			serviceAccountLookupKey := types.NamespacedName{
				Name:      name,
				Namespace: namespace,
			}

			ctx := context.Background()

			defaultAutomountServiceAccountToken := true
			sa := newServiceAccountObject(
				name,
				namespace,
				&defaultAutomountServiceAccountToken,
			)

			// we expect the service account to be created without any issues
			Expect(k8sClient.Create(ctx, sa)).Should(Succeed())

			createdServiceAccount := &corev1.ServiceAccount{}

			// We now try to grab the service account in a retry loop as it will not be created immediately
			Eventually(func() bool {
				err := k8sClient.Get(ctx, serviceAccountLookupKey, createdServiceAccount)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			// we now check to see if the service account has been patched by the controllers
			// this will also not be immediate, so we use  Eventually()
			By("the default service account patch controller")
			Eventually(func() bool {
				patchedServiceAccount := &corev1.ServiceAccount{}
				err := k8sClient.Get(ctx, serviceAccountLookupKey, patchedServiceAccount)
				if err != nil {
					return false
				}
				// clearer to read than !*patchedServiceAccount.AutomountServiceAccountToken, at least for me :)
				if *patchedServiceAccount.AutomountServiceAccountToken == false {
					return true
				}
				//dLog.Printf("PatchedAutomount=%+v", *patchedServiceAccount.AutomountServiceAccountToken)
				return false
			}, timeout, interval).Should(BeTrue())

		})
	})

	Context("When default service account is created with AutomountServiceAccountToken set to nil in kube-system ns", func() {
		It("should be patched successfully ", func() {

			// name of the service account
			name := "default"
			// namespace where the service account will be created
			namespace := "kube-system"

			serviceAccountLookupKey := types.NamespacedName{
				Name:      name,
				Namespace: namespace,
			}
			ctx := context.Background()

			// create the same service account but this time we set the value of
			saWithNilField := newServiceAccountObject(
				name,
				namespace,
				nil, // we want to set this value to nil
			)
			// we expect the service account to be created without any issues
			Expect(k8sClient.Create(ctx, saWithNilField)).Should(Succeed())

			createdServiceAccount := &corev1.ServiceAccount{}

			// We now try to grab the service account in a retry loop as it will not be created immediately
			// the same lookup key should work as we haven't changed the name and namespace of the object
			Eventually(func() bool {
				err := k8sClient.Get(ctx, serviceAccountLookupKey, createdServiceAccount)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			// finally we check to ensure that the controllers has acted on our new serviceaccount and
			// set the automountServiceAccount to false
			By("the default service account patch controllers")
			Eventually(func() bool {
				patchedServiceAccount := &corev1.ServiceAccount{}
				err := k8sClient.Get(ctx, serviceAccountLookupKey, patchedServiceAccount)
				if err != nil {
					return false
				}
				// clearer to read than !*patchedServiceAccount.AutomountServiceAccountToken, at least for me :)
				if *patchedServiceAccount.AutomountServiceAccountToken == false {
					return true
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})
	})

})
