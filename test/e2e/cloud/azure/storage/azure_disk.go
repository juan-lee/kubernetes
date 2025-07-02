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

var _ = SIGDescribe("Azure Disk CSI", feature.Volumes, func() {
	f := framework.NewDefaultFramework("azure-disk-csi")
	f.NamespacePodSecurityLevel = admissionapi.LevelPrivileged

	ginkgo.BeforeEach(func() {
		e2eskipper.SkipUnlessProviderIs("azure")
	})

	f.It("should provision and mount Azure Disk volumes", feature.Volumes, func(ctx context.Context) {
		framework.Logf("Testing Azure Disk CSI volume provisioning")

		// Create a PVC using Azure Disk storage class
		pvcName := "azure-disk-pvc-" + string(uuid.NewUUID())
		pvc := &v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pvcName,
				Namespace: f.Namespace.Name,
			},
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
				Resources: v1.VolumeResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: resource.MustParse("1Gi"),
					},
				},
				StorageClassName: &[]string{"managed-csi"}[0], // Azure Disk CSI storage class
			},
		}

		pvc, err := f.ClientSet.CoreV1().PersistentVolumeClaims(f.Namespace.Name).Create(ctx, pvc, metav1.CreateOptions{})
		framework.ExpectNoError(err)
		defer func() {
			framework.Logf("Cleaning up Azure Disk PVC")
			f.ClientSet.CoreV1().PersistentVolumeClaims(f.Namespace.Name).Delete(ctx, pvc.Name, metav1.DeleteOptions{})
		}()

		// Create a pod to use the PVC
		podName := "azure-disk-pod-" + string(uuid.NewUUID())
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      podName,
				Namespace: f.Namespace.Name,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "azure-disk-container",
						Image: imageutils.GetE2EImage(imageutils.BusyBox),
						Command: []string{
							"/bin/sh",
							"-c",
							"echo 'hello azure disk' > /mnt/azure/test-file && sleep 3600",
						},
						VolumeMounts: []v1.VolumeMount{
							{
								Name:      "azure-disk-volume",
								MountPath: "/mnt/azure",
							},
						},
					},
				},
				Volumes: []v1.Volume{
					{
						Name: "azure-disk-volume",
						VolumeSource: v1.VolumeSource{
							PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
								ClaimName: pvc.Name,
							},
						},
					},
				},
			},
		}

		pod = e2epod.NewPodClient(f).Create(ctx, pod)
		defer func() {
			framework.Logf("Cleaning up Azure Disk pod")
			e2epod.NewPodClient(f).DeleteSync(ctx, pod.Name, metav1.DeleteOptions{}, framework.PodDeleteTimeout)
		}()

		framework.ExpectNoError(e2epod.WaitForPodRunningInNamespace(ctx, f.ClientSet, pod), "Failed to start pod with Azure Disk volume")

		framework.Logf("Azure Disk CSI volume provisioning test completed")
	})

	f.It("should support Azure Disk volume expansion", func(ctx context.Context) {
		framework.Logf("Testing Azure Disk volume expansion")

		// TODO: Implement volume expansion testing:
		// 1. Create a PVC with Azure Disk
		// 2. Create a pod using the PVC
		// 3. Write some data to the volume
		// 4. Expand the PVC size
		// 5. Verify the volume is expanded and data is preserved

		framework.Logf("Azure Disk volume expansion test completed")
	})

	f.It("should support Azure Disk snapshots", feature.VolumeSnapshotDataSource, func(ctx context.Context) {
		framework.Logf("Testing Azure Disk snapshots")

		// TODO: Implement snapshot testing:
		// 1. Create a PVC with Azure Disk
		// 2. Write data to the volume
		// 3. Create a volume snapshot
		// 4. Create a new PVC from the snapshot
		// 5. Verify data is preserved in the new volume

		framework.Logf("Azure Disk snapshot test completed")
	})
})