/*
Copyright (C) 2022-2025 ApeCloud Co., Ltd

This file is part of KubeBlocks project

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package operations

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	dpv1alpha1 "github.com/apecloud/kubeblocks/apis/dataprotection/v1alpha1"
	opsv1alpha1 "github.com/apecloud/kubeblocks/apis/operations/v1alpha1"
	"github.com/apecloud/kubeblocks/pkg/constant"
	intctrlutil "github.com/apecloud/kubeblocks/pkg/controllerutil"
)

var _ = Describe("cleanupTmpResources", func() {

	var (
		cli         client.Client
		reqCtx      intctrlutil.RequestCtx
		opsRes      *OpsResource
		restoreName = "test-restore"
		podName     = "test-pod"
	)

	createTestResources := func() {
		scheme := runtime.NewScheme()
		Ω(opsv1alpha1.AddToScheme(scheme)).Should(Succeed())
		Ω(dpv1alpha1.AddToScheme(scheme)).Should(Succeed())
		Ω(corev1.AddToScheme(scheme)).Should(Succeed())

		opsRequest := &opsv1alpha1.OpsRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-opsrequest",
				Namespace: "default",
			},
		}

		opsRes = &OpsResource{
			OpsRequest: opsRequest,
		}

		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      podName,
				Namespace: "default",
				Labels: map[string]string{
					constant.OpsRequestNameLabelKey:      opsRequest.Name,
					constant.OpsRequestNamespaceLabelKey: opsRequest.Namespace,
				},
			},
		}

		restore := &dpv1alpha1.Restore{
			ObjectMeta: metav1.ObjectMeta{
				Name:      restoreName,
				Namespace: "default",
				Labels: map[string]string{
					constant.OpsRequestNameLabelKey:      opsRequest.Name,
					constant.OpsRequestNamespaceLabelKey: opsRequest.Namespace,
				},
			},
		}

		objects := []runtime.Object{pod, restore}
		cli = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
		reqCtx = intctrlutil.RequestCtx{Ctx: context.Background()}
	}

	Context("when deleteRestoreCR is false", func() {
		It("should not delete Restore CRs", func() {
			createTestResources()

			handler := rebuildInstanceOpsHandler{}
			err := handler.cleanupTmpResources(reqCtx, cli, opsRes, false)
			Ω(err).ShouldNot(HaveOccurred())

			restoreList := &dpv1alpha1.RestoreList{}
			Ω(cli.List(reqCtx.Ctx, restoreList, client.MatchingLabels{
				constant.OpsRequestNameLabelKey:      opsRes.OpsRequest.Name,
				constant.OpsRequestNamespaceLabelKey: opsRes.OpsRequest.Namespace,
			})).Should(Succeed())
			Ω(restoreList.Items).Should(HaveLen(1))
		})
	})

	Context("when deleteRestoreCR is true", func() {
		It("should delete Restore CRs", func() {
			createTestResources()

			handler := rebuildInstanceOpsHandler{}
			err := handler.cleanupTmpResources(reqCtx, cli, opsRes, true)
			Ω(err).ShouldNot(HaveOccurred())

			restoreList := &dpv1alpha1.RestoreList{}
			Ω(cli.List(reqCtx.Ctx, restoreList, client.MatchingLabels{
				constant.OpsRequestNameLabelKey:      opsRes.OpsRequest.Name,
				constant.OpsRequestNamespaceLabelKey: opsRes.OpsRequest.Namespace,
			})).Should(Succeed())
			Ω(restoreList.Items).Should(BeEmpty())
		})
	})
})

func TestCleanupTmpResources(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cleanupTmpResources Test Suite")
}
