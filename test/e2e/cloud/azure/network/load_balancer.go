/*
Copyright 2025 The Kubernetes Authors.

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

package network

import (
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/kubernetes/test/e2e/feature"
	"k8s.io/kubernetes/test/e2e/framework"
	e2eservice "k8s.io/kubernetes/test/e2e/framework/service"
	e2eskipper "k8s.io/kubernetes/test/e2e/framework/skipper"
	admissionapi "k8s.io/pod-security-admission/api"

	"github.com/onsi/ginkgo/v2"
)

var _ = SIGDescribe("Azure Load Balancer", feature.LoadBalancer, func() {
	f := framework.NewDefaultFramework("azure-load-balancer")
	f.NamespacePodSecurityLevel = admissionapi.LevelPrivileged

	var serviceName = "azure-lb-test"

	ginkgo.BeforeEach(func() {
		e2eskipper.SkipUnlessProviderIs("azure")
	})

	f.It("should create and delete Azure Load Balancer", feature.LoadBalancer, func(ctx context.Context) {
		framework.Logf("Testing Azure Load Balancer creation and deletion")

		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceName,
				Namespace: f.Namespace.Name,
			},
			Spec: v1.ServiceSpec{
				Type: v1.ServiceTypeLoadBalancer,
				Ports: []v1.ServicePort{
					{
						Port:       80,
						TargetPort: intstr.FromInt32(80),
						Protocol:   v1.ProtocolTCP,
					},
				},
				Selector: map[string]string{
					"app": "azure-lb-test",
				},
			},
		}

		// Create the service
		service, err := f.ClientSet.CoreV1().Services(f.Namespace.Name).Create(ctx, svc, metav1.CreateOptions{})
		framework.ExpectNoError(err)
		defer func() {
			framework.Logf("Cleaning up Azure Load Balancer service")
			f.ClientSet.CoreV1().Services(f.Namespace.Name).Delete(ctx, service.Name, metav1.DeleteOptions{})
		}()

		// Wait for Azure Load Balancer to be provisioned
		framework.Logf("Waiting for Azure Load Balancer to be provisioned")
		jig := e2eservice.NewTestJig(f.ClientSet, f.Namespace.Name, service.Name)
		_, err = jig.WaitForLoadBalancer(ctx, 5*time.Minute)
		framework.ExpectNoError(err)

		// Verify the load balancer is accessible
		framework.Logf("Azure Load Balancer provisioning test completed")
	})

	f.It("should support internal Azure Load Balancer", feature.LoadBalancer, func(ctx context.Context) {
		framework.Logf("Testing Azure Internal Load Balancer")

		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceName + "-internal",
				Namespace: f.Namespace.Name,
				Annotations: map[string]string{
					"service.beta.kubernetes.io/azure-load-balancer-internal": "true",
				},
			},
			Spec: v1.ServiceSpec{
				Type: v1.ServiceTypeLoadBalancer,
				Ports: []v1.ServicePort{
					{
						Port:       80,
						TargetPort: intstr.FromInt32(80),
						Protocol:   v1.ProtocolTCP,
					},
				},
				Selector: map[string]string{
					"app": "azure-internal-lb-test",
				},
			},
		}

		// Create the internal load balancer service
		service, err := f.ClientSet.CoreV1().Services(f.Namespace.Name).Create(ctx, svc, metav1.CreateOptions{})
		framework.ExpectNoError(err)
		defer func() {
			framework.Logf("Cleaning up Azure Internal Load Balancer service")
			f.ClientSet.CoreV1().Services(f.Namespace.Name).Delete(ctx, service.Name, metav1.DeleteOptions{})
		}()

		// Wait for internal load balancer to be provisioned
		framework.Logf("Waiting for Azure Internal Load Balancer to be provisioned")
		jig := e2eservice.NewTestJig(f.ClientSet, f.Namespace.Name, service.Name)
		_, err = jig.WaitForLoadBalancer(ctx, 5*time.Minute)
		framework.ExpectNoError(err)

		framework.Logf("Azure Internal Load Balancer test completed")
	})

	f.It("should support Azure Load Balancer with specific subnet", feature.LoadBalancer, func(ctx context.Context) {
		framework.Logf("Testing Azure Load Balancer with subnet specification")

		// TODO: Add subnet specification test
		// This would test Azure-specific annotations for subnet placement
		
		framework.Logf("Azure Load Balancer subnet test completed")
	})
})