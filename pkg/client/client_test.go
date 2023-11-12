package client_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/bakito/adguardhome-sync/pkg/client"
	"github.com/bakito/adguardhome-sync/pkg/client/model"
	"github.com/bakito/adguardhome-sync/pkg/types"
	"github.com/bakito/adguardhome-sync/pkg/utils"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var (
	username = uuid.NewString()
	password = uuid.NewString()
)

var _ = Describe("Client", func() {
	var (
		cl client.Client
		ts *httptest.Server
	)
	AfterEach(func() {
		if ts != nil {
			ts.Close()
		}
	})

	Context("Host", func() {
		It("should read the current host", func() {
			cl, _ := client.New(types.AdGuardInstance{URL: "https://foo.bar:3000"})
			host := cl.Host()
			Ω(host).Should(Equal("foo.bar:3000"))
		})
	})

	Context("Filter", func() {
		It("should read filter status", func() {
			ts, cl = ClientGet("filtering-status.json", "/filtering/status")
			fs, err := cl.Filtering()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(*fs.Enabled).Should(BeTrue())
			Ω(*fs.Filters).Should(HaveLen(2))
		})
		It("should enable protection", func() {
			ts, cl = ClientPost("/filtering/config", `{"enabled":true,"interval":123}`)
			err := cl.ToggleFiltering(true, 123)
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("should disable protection", func() {
			ts, cl = ClientPost("/filtering/config", `{"enabled":false,"interval":123}`)
			err := cl.ToggleFiltering(false, 123)
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("should call RefreshFilters", func() {
			ts, cl = ClientPost("/filtering/refresh", `{"whitelist":true}`)
			err := cl.RefreshFilters(true)
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("should add Filters", func() {
			ts, cl = ClientPost("/filtering/add_url",
				`{"name":"","url":"foo","whitelist":true}`,
				`{"name":"","url":"bar","whitelist":true}`,
			)
			err := cl.AddFilters(true, model.Filter{Url: "foo"}, model.Filter{Url: "bar"})
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("should update Filters", func() {
			ts, cl = ClientPost("/filtering/set_url",
				`{"data":{"enabled":false,"name":"","url":"foo"},"url":"foo","whitelist":true}`,
				`{"data":{"enabled":false,"name":"","url":"bar"},"url":"bar","whitelist":true}`,
			)
			err := cl.UpdateFilters(true, model.Filter{Url: "foo"}, model.Filter{Url: "bar"})
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("should delete Filters", func() {
			ts, cl = ClientPost("/filtering/remove_url",
				`{"url":"foo","whitelist":true}`,
				`{"url":"bar","whitelist":true}`,
			)
			err := cl.DeleteFilters(true, model.Filter{Url: "foo"}, model.Filter{Url: "bar"})
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Context("Status", func() {
		It("should read status", func() {
			ts, cl = ClientGet("status.json", "/status")
			fs, err := cl.Status()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(fs.DnsAddresses).Should(HaveLen(1))
			Ω(fs.DnsAddresses[0]).Should(Equal("192.168.1.2"))
			Ω(fs.Version).Should(Equal("v0.105.2"))
		})
		It("should return ErrSetupNeeded", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Location", "/install.html")
				w.WriteHeader(http.StatusFound)
			}))
			cl, err := client.New(types.AdGuardInstance{URL: ts.URL})
			Ω(err).ShouldNot(HaveOccurred())
			_, err = cl.Status()
			Ω(err).Should(HaveOccurred())
			Ω(err).Should(Equal(client.ErrSetupNeeded))
		})
	})

	Context("Setup", func() {
		It("should add setup the instance", func() {
			ts, cl = ClientPost("/install/configure", fmt.Sprintf(`{"web":{"ip":"0.0.0.0","port":3000,"status":"","can_autofix":false},"dns":{"ip":"0.0.0.0","port":53,"status":"","can_autofix":false},"username":"%s","password":"%s"}`, username, password))
			err := cl.Setup()
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Context("RewriteList", func() {
		It("should read RewriteList", func() {
			ts, cl = ClientGet("rewrite-list.json", "/rewrite/list")
			rwl, err := cl.RewriteList()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(*rwl).Should(HaveLen(2))
		})
		It("should add RewriteList", func() {
			ts, cl = ClientPost("/rewrite/add", `{"answer":"foo","domain":"foo"}`, `{"answer":"bar","domain":"bar"}`)
			err := cl.AddRewriteEntries(
				model.RewriteEntry{Answer: utils.Ptr("foo"), Domain: utils.Ptr("foo")},
				model.RewriteEntry{Answer: utils.Ptr("bar"), Domain: utils.Ptr("bar")},
			)
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("should delete RewriteList", func() {
			ts, cl = ClientPost("/rewrite/delete", `{"answer":"foo","domain":"foo"}`, `{"answer":"bar","domain":"bar"}`)
			err := cl.DeleteRewriteEntries(
				model.RewriteEntry{Answer: utils.Ptr("foo"), Domain: utils.Ptr("foo")},
				model.RewriteEntry{Answer: utils.Ptr("bar"), Domain: utils.Ptr("bar")},
			)
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Context("SafeBrowsing", func() {
		It("should read safebrowsing status", func() {
			ts, cl = ClientGet("safebrowsing-status.json", "/safebrowsing/status")
			sb, err := cl.SafeBrowsing()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(sb).Should(BeTrue())
		})
		It("should enable safebrowsing", func() {
			ts, cl = ClientPost("/safebrowsing/enable", "")
			err := cl.ToggleSafeBrowsing(true)
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("should disable safebrowsing", func() {
			ts, cl = ClientPost("/safebrowsing/disable", "")
			err := cl.ToggleSafeBrowsing(false)
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Context("SafeSearch", func() {
		It("should read safesearch status", func() {
			ts, cl = ClientGet("safesearch-status.json", "/safesearch/status")
			ss, err := cl.SafeSearch()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(ss).Should(BeTrue())
		})
		It("should enable safesearch", func() {
			ts, cl = ClientPost("/safesearch/enable", "")
			err := cl.ToggleSafeSearch(true)
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("should disable safesearch", func() {
			ts, cl = ClientPost("/safesearch/disable", "")
			err := cl.ToggleSafeSearch(false)
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Context("Parental", func() {
		It("should read parental status", func() {
			ts, cl = ClientGet("parental-status.json", "/parental/status")
			p, err := cl.Parental()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(p).Should(BeTrue())
		})
		It("should enable parental", func() {
			ts, cl = ClientPost("/parental/enable", "")
			err := cl.ToggleParental(true)
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("should disable parental", func() {
			ts, cl = ClientPost("/parental/disable", "")
			err := cl.ToggleParental(false)
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Context("Protection", func() {
		It("should enable protection", func() {
			ts, cl = ClientPost("/dns_config", `{"protection_enabled":true}`)
			err := cl.ToggleProtection(true)
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("should disable protection", func() {
			ts, cl = ClientPost("/dns_config", `{"protection_enabled":false}`)
			err := cl.ToggleProtection(false)
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Context("Services", func() {
		It("should read Services", func() {
			ts, cl = ClientGet("blockedservices-list.json", "/blocked_services/list")
			s, err := cl.Services()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(*s).Should(HaveLen(2))
		})
		It("should set Services", func() {
			ts, cl = ClientPost("/blocked_services/set", `["foo","bar"]`)
			err := cl.SetServices(&model.BlockedServicesArray{"foo", "bar"})
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Context("Clients", func() {
		It("should read Clients", func() {
			ts, cl = ClientGet("clients.json", "/clients")
			c, err := cl.Clients()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(*c.Clients).Should(HaveLen(2))
		})
		It("should add Clients", func() {
			ts, cl = ClientPost("/clients/add",
				`{"ids":["id"],"name":"foo"}`,
			)
			err := cl.AddClients(&model.Client{Name: utils.Ptr("foo"), Ids: utils.Ptr([]string{"id"})})
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("should update Clients", func() {
			ts, cl = ClientPost("/clients/update",
				`{"data":{"ids":["id"],"name":"foo"},"name":"foo"}`,
			)
			err := cl.UpdateClients(&model.Client{Name: utils.Ptr("foo"), Ids: utils.Ptr([]string{"id"})})
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("should delete Clients", func() {
			ts, cl = ClientPost("/clients/delete",
				`{"ids":["id"],"name":"foo"}`,
			)
			err := cl.DeleteClients(&model.Client{Name: utils.Ptr("foo"), Ids: utils.Ptr([]string{"id"})})
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Context("QueryLogConfig", func() {
		It("should read QueryLogConfig", func() {
			ts, cl = ClientGet("querylog_info.json", "/querylog_info")
			qlc, err := cl.QueryLogConfig()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(qlc.Enabled).ShouldNot(BeNil())
			Ω(*qlc.Enabled).Should(BeTrue())
			Ω(qlc.Interval).ShouldNot(BeNil())
			Ω(*qlc.Interval).Should(Equal(model.QueryLogConfigInterval(90)))
		})
		It("should set QueryLogConfig", func() {
			ts, cl = ClientPost("/querylog_config", `{"anonymize_client_ip":true,"enabled":true,"interval":123}`)

			var interval model.QueryLogConfigInterval = 123
			err := cl.SetQueryLogConfig(&model.QueryLogConfig{AnonymizeClientIp: ptr.To(true), Interval: &interval, Enabled: ptr.To(true)})
			Ω(err).ShouldNot(HaveOccurred())
		})
	})
	Context("StatsConfig", func() {
		It("should read StatsConfig", func() {
			ts, cl = ClientGet("stats_info.json", "/stats_info")
			sc, err := cl.StatsConfig()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(sc.Interval).ShouldNot(BeNil())
			Ω(*sc.Interval).Should(Equal(model.StatsConfigInterval(1)))
		})
		It("should set StatsConfig", func() {
			ts, cl = ClientPost("/stats_config", `{"interval":123}`)

			var interval model.StatsConfigInterval = 123
			err := cl.SetStatsConfig(&model.StatsConfig{Interval: &interval})
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Context("helper functions", func() {
		var cl client.Client
		BeforeEach(func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			}))
			var err error
			cl, err = client.New(types.AdGuardInstance{URL: ts.URL})
			Ω(err).ShouldNot(HaveOccurred())
		})
		Context("doGet", func() {
			It("should return an error on status code != 200", func() {
				_, err := cl.Status()
				Ω(err).Should(HaveOccurred())
				Ω(err.Error()).Should(Equal("401 Unauthorized"))
			})
		})

		Context("doPost", func() {
			It("should return an error on status code != 200", func() {
				var interval model.StatsConfigInterval = 123
				err := cl.SetStatsConfig(&model.StatsConfig{Interval: &interval})
				Ω(err).Should(HaveOccurred())
				Ω(err.Error()).Should(Equal("401 Unauthorized"))
			})
		})
	})
})

func ClientGet(file string, path string) (*httptest.Server, client.Client) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ω(r.URL.Path).Should(Equal(types.DefaultAPIPath + path))
		b, err := os.ReadFile(filepath.Join("../../testdata", file))
		Ω(err).ShouldNot(HaveOccurred())
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(b)
		Ω(err).ShouldNot(HaveOccurred())
	}))
	cl, err := client.New(types.AdGuardInstance{URL: ts.URL})
	Ω(err).ShouldNot(HaveOccurred())
	return ts, cl
}

func ClientPost(path string, content ...string) (*httptest.Server, client.Client) {
	index := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ω(r.URL.Path).Should(Equal(types.DefaultAPIPath + path))
		body, err := io.ReadAll(r.Body)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(body).Should(Equal([]byte(content[index])))
		index++
	}))

	cl, err := client.New(types.AdGuardInstance{URL: ts.URL, Username: username, Password: password})
	Ω(err).ShouldNot(HaveOccurred())
	return ts, cl
}
