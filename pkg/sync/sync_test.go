package sync

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mc "github.com/bakito/adguardhome-sync/pkg/mocks/client"
	"github.com/bakito/adguardhome-sync/pkg/types"
	gm "github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

var _ = Describe("Sync", func() {
	var (
		mockCtrl *gm.Controller
		client   *mc.MockClient
		w        *worker
		te       error
	)

	BeforeEach(func() {
		mockCtrl = gm.NewController(GinkgoT())
		client = mc.NewMockClient(mockCtrl)
		w = &worker{}
		te = errors.New(uuid.NewString())
	})
	AfterEach(func() {
		defer mockCtrl.Finish()
	})

	Context("worker", func() {
		Context("syncRewrites", func() {
			var (
				domain string
				answer string
				reO    types.RewriteEntries
				reR    types.RewriteEntries
			)

			BeforeEach(func() {
				domain = uuid.NewString()
				answer = uuid.NewString()
				reO = []types.RewriteEntry{{Domain: domain, Answer: answer}}
				reR = []types.RewriteEntry{{Domain: domain, Answer: answer}}
			})
			It("should have no changes (empty slices)", func() {
				client.EXPECT().RewriteList().Return(&reR, nil)
				client.EXPECT().AddRewriteEntries()
				client.EXPECT().DeleteRewriteEntries()
				err := w.syncRewrites(&reO, client)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should add one rewrite entry", func() {
				reR = []types.RewriteEntry{}
				client.EXPECT().RewriteList().Return(&reR, nil)
				client.EXPECT().AddRewriteEntries(reO[0])
				client.EXPECT().DeleteRewriteEntries()
				err := w.syncRewrites(&reO, client)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should remove one rewrite entry", func() {
				reO = []types.RewriteEntry{}
				client.EXPECT().RewriteList().Return(&reR, nil)
				client.EXPECT().AddRewriteEntries()
				client.EXPECT().DeleteRewriteEntries(reR[0])
				err := w.syncRewrites(&reO, client)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should return error when error on RewriteList()", func() {
				client.EXPECT().RewriteList().Return(nil, te)
				err := w.syncRewrites(&reO, client)
				Ω(err).Should(HaveOccurred())
			})
			It("should return error when error on AddRewriteEntries()", func() {
				client.EXPECT().RewriteList().Return(&reR, nil)
				client.EXPECT().AddRewriteEntries().Return(te)
				err := w.syncRewrites(&reO, client)
				Ω(err).Should(HaveOccurred())
			})
			It("should return error when error on DeleteRewriteEntries()", func() {
				client.EXPECT().RewriteList().Return(&reR, nil)
				client.EXPECT().AddRewriteEntries()
				client.EXPECT().DeleteRewriteEntries().Return(te)
				err := w.syncRewrites(&reO, client)
				Ω(err).Should(HaveOccurred())
			})
		})
		Context("syncClients", func() {
			var (
				clO *types.Clients
				clR *types.Clients
			)
			BeforeEach(func() {
				clO = &types.Clients{}
				clR = &types.Clients{}
			})
			It("should have no changes (empty slices)", func() {
				client.EXPECT().Clients().Return(clR, nil)
				client.EXPECT().AddClients()
				client.EXPECT().UpdateClients()
				client.EXPECT().DeleteClients()
				err := w.syncClients(clO, client)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
	})
})
