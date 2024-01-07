package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/bakito/adguardhome-sync/pkg/client/model"
	"github.com/bakito/adguardhome-sync/pkg/log"
	"github.com/bakito/adguardhome-sync/pkg/types"
	"github.com/bakito/adguardhome-sync/pkg/utils"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const envRedirectPolicyNoOfRedirects = "REDIRECT_POLICY_NO_OF_REDIRECTS"

var (
	l = log.GetLogger("client")
	// ErrSetupNeeded custom error
	ErrSetupNeeded = errors.New("setup needed")
)

func detailedError(resp *resty.Response, err error) error {
	e := resp.Status()
	if len(resp.Body()) > 0 {
		e += fmt.Sprintf("(%s)", string(resp.Body()))
	}
	if err != nil {
		e += fmt.Sprintf(": %s", err.Error())
	}
	return errors.New(e)
}

// New create a new client
func New(config types.AdGuardInstance) (Client, error) {
	var apiURL string
	if config.APIPath == "" {
		apiURL = fmt.Sprintf("%s/control", config.URL)
	} else {
		apiURL = fmt.Sprintf("%s/%s", config.URL, config.APIPath)
	}
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Clean(u.Path)
	cl := resty.New().SetBaseURL(u.String()).SetDisableWarn(true)

	if config.InsecureSkipVerify {
		// #nosec G402 has to be explicitly enabled
		cl.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	cookieParts := strings.Split(config.Cookie, "=")
	if len(cookieParts) == 2 {
		cl.SetCookie(&http.Cookie{
			Name:  cookieParts[0],
			Value: cookieParts[1],
		})
	} else if config.Username != "" && config.Password != "" {
		cl = cl.SetBasicAuth(config.Username, config.Password)
	}

	if v, ok := os.LookupEnv(envRedirectPolicyNoOfRedirects); ok {
		nbr, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("error parsing env var %q value must be an integer", envRedirectPolicyNoOfRedirects)
		}
		cl.SetRedirectPolicy(resty.FlexibleRedirectPolicy(nbr))
	} else {
		// no redirect
		cl.SetRedirectPolicy(resty.NoRedirectPolicy())
	}

	return &client{
		host:   config.Host,
		client: cl,
		log:    l.With("host", config.Host),
	}, nil
}

// Client AdguardHome API client interface
type Client interface {
	Host() string
	Status() (*model.ServerStatus, error)
	ToggleProtection(enable bool) error
	RewriteList() (*model.RewriteEntries, error)
	AddRewriteEntries(e ...model.RewriteEntry) error
	DeleteRewriteEntries(e ...model.RewriteEntry) error
	Filtering() (*model.FilterStatus, error)
	ToggleFiltering(enabled bool, interval int) error
	AddFilters(whitelist bool, e ...model.Filter) error
	DeleteFilters(whitelist bool, e ...model.Filter) error
	UpdateFilters(whitelist bool, e ...model.Filter) error
	RefreshFilters(whitelist bool) error
	SetCustomRules(rules *[]string) error
	SafeBrowsing() (bool, error)
	ToggleSafeBrowsing(enable bool) error
	Parental() (bool, error)
	ToggleParental(enable bool) error
	SafeSearchConfig() (*model.SafeSearchConfig, error)
	SetSafeSearchConfig(settings *model.SafeSearchConfig) error
	ProfileInfo() (*model.ProfileInfo, error)
	SetProfileInfo(settings *model.ProfileInfo) error
	BlockedServices() (*model.BlockedServicesArray, error)
	BlockedServicesSchedule() (*model.BlockedServicesSchedule, error)
	SetBlockedServices(services *model.BlockedServicesArray) error
	SetBlockedServicesSchedule(schedule *model.BlockedServicesSchedule) error
	Clients() (*model.Clients, error)
	AddClients(client ...*model.Client) error
	UpdateClients(client ...*model.Client) error
	DeleteClients(client ...*model.Client) error
	QueryLogConfig() (*model.QueryLogConfig, error)
	SetQueryLogConfig(*model.QueryLogConfig) error
	StatsConfig() (*model.StatsConfig, error)
	SetStatsConfig(sc *model.StatsConfig) error
	Setup() error
	AccessList() (*model.AccessList, error)
	SetAccessList(*model.AccessList) error
	DNSConfig() (*model.DNSConfig, error)
	SetDNSConfig(*model.DNSConfig) error
	DhcpConfig() (*model.DhcpStatus, error)
	SetDhcpConfig(*model.DhcpStatus) error
	AddDHCPStaticLeases(leases ...model.DhcpStaticLease) error
	DeleteDHCPStaticLeases(leases ...model.DhcpStaticLease) error
}

type client struct {
	client  *resty.Client
	log     *zap.SugaredLogger
	host    string
	version string
}

func (cl *client) Host() string {
	return cl.host
}

func contentType(resp *resty.Response) string {
	if ct, ok := resp.Header()["Content-Type"]; ok {
		if len(ct) != 1 {
			return fmt.Sprintf("%v", ct)
		}
		return ct[0]
	}
	return ""
}

func (cl *client) Status() (*model.ServerStatus, error) {
	status := &model.ServerStatus{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(status), "status")
	cl.version = status.Version
	return status, err
}

func (cl *client) RewriteList() (*model.RewriteEntries, error) {
	rewrites := &model.RewriteEntries{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(&rewrites), "/rewrite/list")
	return rewrites, err
}

func (cl *client) AddRewriteEntries(entries ...model.RewriteEntry) error {
	for i := range entries {
		e := entries[i]
		cl.log.With("domain", e.Domain, "answer", e.Answer).Info("Add rewrite entry")
		err := cl.doPost(cl.client.R().EnableTrace().SetBody(&e), "/rewrite/add")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) DeleteRewriteEntries(entries ...model.RewriteEntry) error {
	for i := range entries {
		e := entries[i]
		cl.log.With("domain", e.Domain, "answer", e.Answer).Info("Delete rewrite entry")
		err := cl.doPost(cl.client.R().EnableTrace().SetBody(&e), "/rewrite/delete")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) SafeBrowsing() (bool, error) {
	return cl.toggleStatus("safebrowsing")
}

func (cl *client) ToggleSafeBrowsing(enable bool) error {
	return cl.toggleBool("safebrowsing", enable)
}

func (cl *client) Parental() (bool, error) {
	return cl.toggleStatus("parental")
}

func (cl *client) ToggleParental(enable bool) error {
	return cl.toggleBool("parental", enable)
}

func (cl *client) toggleStatus(mode string) (bool, error) {
	fs := &model.EnableConfig{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(fs), fmt.Sprintf("/%s/status", mode))
	return fs.Enabled, err
}

func (cl *client) toggleBool(mode string, enable bool) error {
	cl.log.With("enable", enable).Info(fmt.Sprintf("Toggle %s", mode))
	var target string
	if enable {
		target = "enable"
	} else {
		target = "disable"
	}
	return cl.doPost(cl.client.R().EnableTrace(), fmt.Sprintf("/%s/%s", mode, target))
}

func (cl *client) Filtering() (*model.FilterStatus, error) {
	f := &model.FilterStatus{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(f), "/filtering/status")
	return f, err
}

func (cl *client) AddFilters(whitelist bool, filters ...model.Filter) error {
	for _, f := range filters {
		cl.log.With("url", f.Url, "whitelist", whitelist, "enabled", f.Enabled).Info("Add filter")
		ff := &model.AddUrlRequest{Name: utils.Ptr(f.Name), Url: utils.Ptr(f.Url), Whitelist: utils.Ptr(whitelist)}
		err := cl.doPost(cl.client.R().EnableTrace().SetBody(ff), "/filtering/add_url")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) DeleteFilters(whitelist bool, filters ...model.Filter) error {
	for _, f := range filters {
		cl.log.With("url", f.Url, "whitelist", whitelist, "enabled", f.Enabled).Info("Delete filter")
		ff := &model.RemoveUrlRequest{Url: utils.Ptr(f.Url), Whitelist: utils.Ptr(whitelist)}
		err := cl.doPost(cl.client.R().EnableTrace().SetBody(ff), "/filtering/remove_url")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) UpdateFilters(whitelist bool, filters ...model.Filter) error {
	for _, f := range filters {
		cl.log.With("url", f.Url, "whitelist", whitelist, "enabled", f.Enabled).Info("Update filter")
		fu := &model.FilterSetUrl{
			Whitelist: utils.Ptr(whitelist), Url: utils.Ptr(f.Url),
			Data: &model.FilterSetUrlData{Name: f.Name, Url: f.Url, Enabled: f.Enabled},
		}
		err := cl.doPost(cl.client.R().EnableTrace().SetBody(fu), "/filtering/set_url")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) RefreshFilters(whitelist bool) error {
	cl.log.With("whitelist", whitelist).Info("Refresh filter")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(&model.FilterRefreshRequest{Whitelist: utils.Ptr(whitelist)}), "/filtering/refresh")
}

func (cl *client) ToggleProtection(enable bool) error {
	cl.log.With("enable", enable).Info("Toggle protection")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(&types.Protection{ProtectionEnabled: enable}), "/dns_config")
}

func (cl *client) SetCustomRules(rules *[]string) error {
	cl.log.With("rules", len(*rules)).Info("Set user rules")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(&model.SetRulesRequest{Rules: rules}), "/filtering/set_rules")
}

func (cl *client) ToggleFiltering(enabled bool, interval int) error {
	cl.log.With("enabled", enabled, "interval", interval).Info("Toggle filtering")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(&model.FilterConfig{
		Enabled:  utils.Ptr(enabled),
		Interval: utils.Ptr(interval),
	}), "/filtering/config")
}

func (cl *client) BlockedServices() (*model.BlockedServicesArray, error) {
	svcs := &model.BlockedServicesArray{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(svcs), "/blocked_services/list")
	return svcs, err
}

func (cl *client) SetBlockedServices(services *model.BlockedServicesArray) error {
	cl.log.With("services", model.ArrayString(services)).Info("Set blocked services")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(services), "/blocked_services/set")
}

func (cl *client) BlockedServicesSchedule() (*model.BlockedServicesSchedule, error) {
	sched := &model.BlockedServicesSchedule{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(sched), "/blocked_services/get")
	return sched, err
}

func (cl *client) SetBlockedServicesSchedule(schedule *model.BlockedServicesSchedule) error {
	cl.log.With("services", schedule.ServicesString()).Info("Set blocked services schedule")
	return cl.doPut(cl.client.R().EnableTrace().SetBody(schedule), "/blocked_services/update")
}

func (cl *client) Clients() (*model.Clients, error) {
	clients := &model.Clients{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(clients), "/clients")
	return clients, err
}

func (cl *client) AddClients(clients ...*model.Client) error {
	for i := range clients {
		client := clients[i]
		cl.log.With("name", *client.Name).Info("Add client")
		err := cl.doPost(cl.client.R().EnableTrace().SetBody(client), "/clients/add")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) UpdateClients(clients ...*model.Client) error {
	for _, client := range clients {
		cl.log.With("name", *client.Name).Info("Update client")
		err := cl.doPost(cl.client.R().EnableTrace().SetBody(&model.ClientUpdate{Name: client.Name, Data: client}), "/clients/update")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) DeleteClients(clients ...*model.Client) error {
	for i := range clients {
		client := clients[i]
		cl.log.With("name", *client.Name).Info("Delete client")
		err := cl.doPost(cl.client.R().EnableTrace().SetBody(client), "/clients/delete")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) QueryLogConfig() (*model.QueryLogConfig, error) {
	qlc := &model.QueryLogConfig{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(qlc), "/querylog_info")
	return qlc, err
}

func (cl *client) SetQueryLogConfig(qlc *model.QueryLogConfig) error {
	cl.log.With("enabled", *qlc.Enabled, "interval", *qlc.Interval, "anonymizeClientIP", *qlc.AnonymizeClientIp).Info("Set query log config")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(qlc), "/querylog_config")
}

func (cl *client) StatsConfig() (*model.StatsConfig, error) {
	stats := &model.StatsConfig{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(stats), "/stats_info")
	return stats, err
}

func (cl *client) SetStatsConfig(sc *model.StatsConfig) error {
	cl.log.With("interval", *sc.Interval).Info("Set stats config")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(sc), "/stats_config")
}

func (cl *client) Setup() error {
	cl.log.Info("Setup new AdguardHome instance")
	cfg := &types.InstallConfig{
		Web: types.InstallPort{
			IP:         "0.0.0.0",
			Port:       3000,
			Status:     "",
			CanAutofix: false,
		},
		DNS: types.InstallPort{
			IP:         "0.0.0.0",
			Port:       53,
			Status:     "",
			CanAutofix: false,
		},
	}

	if cl.client.UserInfo != nil {
		cfg.Username = cl.client.UserInfo.Username
		cfg.Password = cl.client.UserInfo.Password
	}
	req := cl.client.R().EnableTrace().SetBody(cfg)
	req.UserInfo = nil
	return cl.doPost(req, "/install/configure")
}

func (cl *client) AccessList() (*model.AccessList, error) {
	al := &model.AccessList{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(al), "/access/list")
	return al, err
}

func (cl *client) SetAccessList(list *model.AccessList) error {
	cl.log.Info("Set access list")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(list), "/access/set")
}

func (cl *client) DNSConfig() (*model.DNSConfig, error) {
	cfg := &model.DNSConfig{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(cfg), "/dns_info")
	return cfg, err
}

func (cl *client) SetDNSConfig(config *model.DNSConfig) error {
	cl.log.Info("Set dns config list")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(config), "/dns_config")
}

func (cl *client) DhcpConfig() (*model.DhcpStatus, error) {
	cfg := &model.DhcpStatus{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(cfg), "/dhcp/status")
	return cfg, err
}

func (cl *client) SetDhcpConfig(config *model.DhcpStatus) error {
	cl.log.Info("Set dhcp server config")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(config), "/dhcp/set_config")
}

func (cl *client) AddDHCPStaticLeases(leases ...model.DhcpStaticLease) error {
	for _, l := range leases {
		cl.log.With("mac", l.Mac, "ip", l.Ip, "hostname", l.Hostname).Info("Add static dhcp lease")
		err := cl.doPost(cl.client.R().EnableTrace().SetBody(l), "/dhcp/add_static_lease")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) DeleteDHCPStaticLeases(leases ...model.DhcpStaticLease) error {
	for _, l := range leases {
		cl.log.With("mac", l.Mac, "ip", l.Ip, "hostname", l.Hostname).Info("Delete static dhcp lease")
		err := cl.doPost(cl.client.R().EnableTrace().SetBody(l), "/dhcp/remove_static_lease")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) SafeSearchConfig() (*model.SafeSearchConfig, error) {
	sss := &model.SafeSearchConfig{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(sss), "/safesearch/status")
	return sss, err
}

func (cl *client) SetSafeSearchConfig(settings *model.SafeSearchConfig) error {
	cl.log.With("enabled", *settings.Enabled).Info("Set safesearch settings")
	return cl.doPut(cl.client.R().EnableTrace().SetBody(settings), "/safesearch/settings")
}

func (cl *client) ProfileInfo() (*model.ProfileInfo, error) {
	p := &model.ProfileInfo{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(p), "/profile")
	return p, err
}

func (cl *client) SetProfileInfo(profile *model.ProfileInfo) error {
	cl.log.With("language", profile.Language, "theme", profile.Theme).Info("Set profile")
	return cl.doPut(cl.client.R().EnableTrace().SetBody(profile), "/profile/update")
}
