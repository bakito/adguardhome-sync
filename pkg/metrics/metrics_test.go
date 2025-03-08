package metrics

import (
	"github.com/bakito/adguardhome-sync/pkg/client/model"
	"github.com/go-faker/faker/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("Metrics", func() {
	Context("Update / getStats", func() {
		It("generate correct stats", func() {
			Update(InstanceMetricsList{[]InstanceMetrics{
				{HostName: "foo", Status: &model.ServerStatus{}, Stats: &model.Stats{
					NumDnsQueries: ptr.To(100),
					DnsQueries: ptr.To(
						[]int{10, 20, 30, 40, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					),
				}},
				{HostName: "bar", Status: &model.ServerStatus{}, Stats: &model.Stats{
					NumDnsQueries: ptr.To(200),
					DnsQueries: ptr.To(
						[]int{20, 40, 60, 80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					),
				}},
			}})
			Ω(stats).Should(HaveKey("foo"))
			Ω(stats["foo"].NumDnsQueries).Should(Equal(ptr.To(100)))
			Ω(stats).Should(HaveKey("bar"))
			Ω(stats["bar"].NumDnsQueries).Should(Equal(ptr.To(200)))

			os := getStats()
			tot := os.Total()
			Ω(*tot.NumDnsQueries).Should(Equal(300))
			Ω(
				*tot.DnsQueries,
			).Should(Equal([]int{30, 60, 90, 120, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))

			foo := os["foo"]
			bar := os["bar"]

			Ω(*foo.NumDnsQueries).Should(Equal(100))
			Ω(*bar.NumDnsQueries).Should(Equal(200))
			Ω(
				*foo.DnsQueries,
			).Should(Equal([]int{10, 20, 30, 40, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
			Ω(
				*bar.DnsQueries,
			).Should(Equal([]int{20, 40, 60, 80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
		})
	})
	Context("StatsGraph", func() {
		var metrics InstanceMetricsList
		BeforeEach(func() {
			err := faker.FakeData(&metrics)
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("should provide correct results with faked values", func() {
			Update(metrics)

			total, dns, blocked, malware, adult := StatsGraph()
			Ω(total).ShouldNot(BeNil())
			Ω(dns).ShouldNot(BeNil())
			Ω(blocked).ShouldNot(BeNil())
			Ω(malware).ShouldNot(BeNil())
			Ω(adult).ShouldNot(BeNil())
		})
	})
})
