package model

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/zap"

	"github.com/bakito/adguardhome-sync/internal/log"
)

var _ = Describe("Types", func() {
	Context("DhcpConfigV4", func() {
		DescribeTable("DhcpConfigV4 should not be valid",
			func(v4 DhcpConfigV4) {
				gomega.Ω(v4.isValid()).Should(gomega.BeFalse())
			},
			Entry(`When GatewayIp is nil`, DhcpConfigV4{
				GatewayIp:  nil,
				SubnetMask: new("2.2.2.2"),
				RangeStart: new("3.3.3.3"),
				RangeEnd:   new("4.4.4.4"),
			}),
			Entry(`When GatewayIp is ""`, DhcpConfigV4{
				GatewayIp:  new(""),
				SubnetMask: new("2.2.2.2"),
				RangeStart: new("3.3.3.3"),
				RangeEnd:   new("4.4.4.4"),
			}),
			Entry(`When SubnetMask is nil`, DhcpConfigV4{
				GatewayIp:  new("1.1.1.1"),
				SubnetMask: nil,
				RangeStart: new("3.3.3.3"),
				RangeEnd:   new("4.4.4.4"),
			}),
			Entry(`When SubnetMask is ""`, DhcpConfigV4{
				GatewayIp:  new("1.1.1.1"),
				SubnetMask: new(""),
				RangeStart: new("3.3.3.3"),
				RangeEnd:   new("4.4.4.4"),
			}),
			Entry(`When SubnetMask is nil`, DhcpConfigV4{
				GatewayIp:  new("1.1.1.1"),
				SubnetMask: new("2.2.2.2"),
				RangeStart: nil,
				RangeEnd:   new("4.4.4.4"),
			}),
			Entry(`When SubnetMask is ""`, DhcpConfigV4{
				GatewayIp:  new("1.1.1.1"),
				SubnetMask: new("2.2.2.2"),
				RangeStart: new(""),
				RangeEnd:   new("4.4.4.4"),
			}),
			Entry(`When RangeEnd is nil`, DhcpConfigV4{
				GatewayIp:  new("1.1.1.1"),
				SubnetMask: new("2.2.2.2"),
				RangeStart: new("3.3.3.3"),
				RangeEnd:   nil,
			}),
			Entry(`When RangeEnd is ""`, DhcpConfigV4{
				GatewayIp:  new("1.1.1.1"),
				SubnetMask: new("2.2.2.2"),
				RangeStart: new("3.3.3.3"),
				RangeEnd:   new(""),
			}),
		)
	})
	Context("DhcpConfigV6", func() {
		DescribeTable("DhcpConfigV6 should not be valid",
			func(v6 DhcpConfigV6) {
				gomega.Ω(v6.isValid()).Should(gomega.BeFalse())
			},
			Entry(`When SubnetMask is nil`, DhcpConfigV6{RangeStart: nil}),
			Entry(`When SubnetMask is ""`, DhcpConfigV6{RangeStart: new("")}),
		)
	})
	Context("DNSConfig", func() {
		var (
			cfg *DNSConfig
			l   *zap.SugaredLogger
		)

		BeforeEach(func() {
			cfg = &DNSConfig{
				UsePrivatePtrResolvers: new(true),
			}
			l = log.GetLogger("test")
		})
		Context("Sanitize", func() {
			It("should disable UsePrivatePtrResolvers resolvers is nil ", func() {
				cfg.LocalPtrUpstreams = nil
				cfg.Sanitize(l)
				gomega.Ω(cfg.UsePrivatePtrResolvers).ShouldNot(gomega.BeNil())
				gomega.Ω(*cfg.UsePrivatePtrResolvers).Should(gomega.Equal(false))
			})
			It("should disable UsePrivatePtrResolvers resolvers is empty ", func() {
				cfg.LocalPtrUpstreams = new([]string{})
				cfg.Sanitize(l)
				gomega.Ω(cfg.UsePrivatePtrResolvers).ShouldNot(gomega.BeNil())
				gomega.Ω(*cfg.UsePrivatePtrResolvers).Should(gomega.Equal(false))
			})
		})
	})
})
