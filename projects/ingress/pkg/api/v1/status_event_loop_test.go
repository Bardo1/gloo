// Code generated by solo-kit. DO NOT EDIT.

package v1

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var _ = Describe("StatusEventLoop", func() {
	var (
		namespace string
		emitter   StatusEmitter
		err       error
	)

	BeforeEach(func() {

		kubeServiceClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}
		kubeServiceClient, err := NewKubeServiceClient(kubeServiceClientFactory)
		Expect(err).NotTo(HaveOccurred())

		ingressClientFactory := &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}
		ingressClient, err := NewIngressClient(ingressClientFactory)
		Expect(err).NotTo(HaveOccurred())

		emitter = NewStatusEmitter(kubeServiceClient, ingressClient)
	})
	It("runs sync function on a new snapshot", func() {
		_, err = emitter.KubeService().Write(NewKubeService(namespace, "jerry"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		_, err = emitter.Ingress().Write(NewIngress(namespace, "jerry"), clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		sync := &mockStatusSyncer{}
		el := NewStatusEventLoop(emitter, sync)
		_, err := el.Run([]string{namespace}, clients.WatchOpts{})
		Expect(err).NotTo(HaveOccurred())
		Eventually(sync.Synced, 5*time.Second).Should(BeTrue())
	})
})

type mockStatusSyncer struct {
	synced bool
	mutex  sync.Mutex
}

func (s *mockStatusSyncer) Synced() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.synced
}

func (s *mockStatusSyncer) Sync(ctx context.Context, snap *StatusSnapshot) error {
	s.mutex.Lock()
	s.synced = true
	s.mutex.Unlock()
	return nil
}
