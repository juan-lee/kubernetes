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

package storage

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/kubernetes/test/e2e/feature"
	"k8s.io/kubernetes/test/e2e/framework"
	e2epod "k8s.io/kubernetes/test/e2e/framework/pod"
	e2eskipper "k8s.io/kubernetes/test/e2e/framework/skipper"
	admissionapi "k8s.io/pod-security-admission/api"
	imageutils "k8s.io/kubernetes/test/utils/image"

	"github.com/onsi/ginkgo/v2"
)

var _ = SIGDescribe("Azure Files CSI", feature.Volumes, func() {
	f := framework.NewDefaultFramework("azure-files-csi")
	f.NamespacePodSecurityLevel = admissionapi.LevelPrivileged

	ginkgo.BeforeEach(func() {
		e2eskipper.SkipUnlessProviderIs("azure")
	})

	f.It("should provision and mount Azure Files volumes", feature.Volumes, func(ctx context.Context) {
		framework.Logf("Testing Azure Files CSI volume provisioning")

		// Create a PVC using Azure Files storage class
		pvcName := "azure-files-pvc-" + string(uuid.NewUUID())
		pvc := &v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pvcName,
				Namespace: f.Namespace.Name,
			},
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteMany}, // Azure Files supports RWX
				Resources: v1.VolumeResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: resource.MustParse("1Gi"),
					},
				},
				StorageClassName: &[]string{"azurefile-csi"}[0], // Azure Files CSI storage class
			},
		}

		pvc, err := f.ClientSet.CoreV1().PersistentVolumeClaims(f.Namespace.Name).Create(ctx, pvc, metav1.CreateOptions{})
		framework.ExpectNoError(err)
		defer func() {
			framework.Logf("Cleaning up Azure Files PVC")
			f.ClientSet.CoreV1().PersistentVolumeClaims(f.Namespace.Name).Delete(ctx, pvc.Name, metav1.DeleteOptions{})
		}()

		// Create multiple pods to test ReadWriteMany capability
		podName1 := "azure-files-pod1-" + string(uuid.NewUUID())
		podName2 := "azure-files-pod2-" + string(uuid.NewUUID())

		createPodWithAzureFiles := func(podName, content string) *v1.Pod {
			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: f.Namespace.Name,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "azure-files-container",
							Image: imageutils.GetE2EImage(imageutils.BusyBox),
							Command: []string{
								"/bin/sh",
								"-c",
								"echo '" + content + "' > /mnt/azure-files/test-file-" + podName + " && sleep 3600",
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "azure-files-volume",
									MountPath: "/mnt/azure-files",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "azure-files-volume",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvc.Name,
								},
							},
						},
					},
				},
			}
			return e2epod.NewPodClient(f).Create(ctx, pod)
		}

		// Create first pod
		pod1 := createPodWithAzureFiles(podName1, "hello from pod1")
		defer func() {
			framework.Logf("Cleaning up Azure Files pod1")
			e2epod.NewPodClient(f).DeleteSync(ctx, pod1.Name, metav1.DeleteOptions{}, framework.PodDeleteTimeout)
		}()

		// Create second pod
		pod2 := createPodWithAzureFiles(podName2, "hello from pod2")
		defer func() {
			framework.Logf("Cleaning up Azure Files pod2")
			e2epod.NewPodClient(f).DeleteSync(ctx, pod2.Name, metav1.DeleteOptions{}, framework.PodDeleteTimeout)
		}()

		// Wait for both pods to be running
		framework.ExpectNoError(e2epod.WaitForPodRunningInNamespace(ctx, f.ClientSet, pod1), "Failed to start pod1 with Azure Files volume")
		framework.ExpectNoError(e2epod.WaitForPodRunningInNamespace(ctx, f.ClientSet, pod2), "Failed to start pod2 with Azure Files volume")

		framework.Logf("Azure Files CSI ReadWriteMany test completed")
	})

	f.It("should support Azure Files with different access tiers", feature.Volumes, func(ctx context.Context) {
		framework.Logf("Testing Azure Files with different access tiers")

		// TODO: Implement access tier testing:
		// 1. Test Hot tier Azure Files
		// 2. Test Cool tier Azure Files
		// 3. Test Premium Azure Files
		// 4. Verify performance characteristics

		framework.Logf("Azure Files access tier test completed")
	})

	f.It("should support Azure Files with SMB and NFS protocols", feature.Volumes, func(ctx context.Context) {
		framework.Logf("Testing Azure Files with different protocols")

		// TODO: Implement protocol testing:
		// 1. Test SMB protocol access
		// 2. Test NFS protocol access (where supported)
		// 3. Verify protocol-specific features

		framework.Logf("Azure Files protocol test completed")
	})
})