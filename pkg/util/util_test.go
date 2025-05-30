package util

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openshiftconfigv1 "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	logger = zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)).WithName("eventEmitter_test")
)

var _ = Describe("Test general utilities", func() {

	Context("test GetOperatorNamespaceFromEnv", func() {
		var origVal string
		BeforeEach(func() {
			origVal = os.Getenv(OperatorNamespaceEnv)
		})

		AfterEach(func() {
			_ = os.Setenv(OperatorNamespaceEnv, origVal)
		})

		It("should return the namespace from the OPERATOR_NAMESPACE env var", func() {

			const expectedNs = "mynamespace"
			_ = os.Setenv(OperatorNamespaceEnv, expectedNs)

			ns := GetOperatorNamespaceFromEnv()
			Expect(ns).To(Equal(expectedNs))
		})
	})

	Context("test EnsureDeleted", func() {

		const appName = "appName"

		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testPod",
				Namespace: "test-namespace",
				Labels: map[string]string{
					AppLabel: appName,
				},
			},
		}

		testScheme := scheme.Scheme
		Expect(openshiftconfigv1.Install(testScheme)).To(Succeed())

		ctx := context.Background()

		It("should delete an existing resource", func() {
			cl := fake.NewClientBuilder().
				WithScheme(testScheme).
				WithObjects(pod).
				WithStatusSubresource(pod).
				Build()

			deleted, err := EnsureDeleted(ctx, cl, pod, appName, logger, false, true, true)
			Expect(err).ToNot(HaveOccurred())
			Expect(deleted).To(BeTrue())

			podToSearch := &corev1.Pod{}
			Expect(cl.Get(ctx, client.ObjectKeyFromObject(pod), podToSearch)).ToNot(Succeed())
		})

		It("should not return error if the resource does not exist", func() {
			cl := fake.NewClientBuilder().
				WithScheme(testScheme).
				Build()

			deleted, err := EnsureDeleted(ctx, cl, pod, appName, logger, false, true, true)
			Expect(err).ToNot(HaveOccurred())
			Expect(deleted).To(BeFalse())

			podToSearch := &corev1.Pod{}
			Expect(cl.Get(ctx, client.ObjectKeyFromObject(pod), podToSearch)).ToNot(Succeed())
		})
	})
})
