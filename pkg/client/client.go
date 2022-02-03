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

	"github.com/bakito/adguardhome-sync/pkg/log"
	"github.com/bakito/adguardhome-sync/pkg/types"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

var (
	l = log.GetLogger("client")
	// ErrSetupNeeded custom error
	ErrSetupNeeded = errors.New("setup needed")
)

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

	if config.Username != "" && config.Password != "" {
		cl = cl.SetBasicAuth(config.Username, config.Password)
	}

	if v, ok := os.LookupEnv("REDIRECT_POLICY_NO_OF_REDIRECTS"); ok {
		nbr, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("error parsing env var %q value must be an integer", "REDIRECT_POLICY_NO_OF_REDIRECTS")
		}
		cl.SetRedirectPolicy(resty.FlexibleRedirectPolicy(nbr))
	} else {
		// no redirect
		cl.SetRedirectPolicy(resty.NoRedirectPolicy())
	}

	return &client{
		host:   u.Host,
		client: cl,
		log:    l.With("host", u.Host),
	}, nil
}

// Client AdguardHome API client interface
type Client interface {
	Host() string
	Status() (*types.Status, error)
	ToggleProtection(enable bool) error
	RewriteList() (*types.RewriteEntries, error)
	AddRewriteEntries(e ...types.RewriteEntry) error
	DeleteRewriteEntries(e ...types.RewriteEntry) error
	Filtering() (*types.FilteringStatus, error)
	ToggleFiltering(enabled bool, interval float64) error
	AddFilters(whitelist bool, e ...types.Filter) error
	DeleteFilters(whitelist bool, e ...types.Filter) error
	UpdateFilters(whitelist bool, e ...types.Filter) error
	RefreshFilters(whitelist bool) error
	SetCustomRules(rules types.UserRules) error
	SafeBrowsing() (bool, error)
	ToggleSafeBrowsing(enable bool) error
	Parental() (bool, error)
	ToggleParental(enable bool) error
	SafeSearch() (bool, error)
	ToggleSafeSearch(enable bool) error
	Services() (types.Services, error)
	SetServices(services types.Services) error
	Clients() (*types.Clients, error)
	AddClients(client ...types.Client) error
	UpdateClients(client ...types.Client) error
	DeleteClients(client ...types.Client) error
	QueryLogConfig() (*types.QueryLogConfig, error)
	SetQueryLogConfig(enabled bool, interval float64, anonymizeClientIP bool) error
	StatsConfig() (*types.IntervalConfig, error)
	SetStatsConfig(interval float64) error
	Setup() error
	AccessList() (*types.AccessList, error)
	SetAccessList(*types.AccessList) error
	DNSConfig() (*types.DNSConfig, error)
	SetDNSConfig(*types.DNSConfig) error
	DHCPServerConfig() (*types.DHCPServerConfig, error)
	SetDHCPServerConfig(*types.DHCPServerConfig) error
	AddDHCPStaticLeases(leases ...types.Lease) error
	DeleteDHCPStaticLeases(leases ...types.Lease) error
}

type client struct {
	client *resty.Client
	log    *zap.SugaredLogger
	host   string
}

func (cl *client) Host() string {
	return cl.host
}

func (cl *client) doGet(req *resty.Request, url string) error {
	rl := cl.log.With("method", "GET", "path", url)
	if cl.client.UserInfo != nil {
		rl = rl.With("username", cl.client.UserInfo.Username)
	}
	rl.Debug("do get")
	resp, err := req.Get(url)
	if err != nil {
		if resp != nil && resp.StatusCode() == http.StatusFound {
			loc := resp.Header().Get("Location")
			if loc == "/install.html" {
				return ErrSetupNeeded
			}
		}
		rl.With("status", resp.StatusCode(), "body", string(resp.Body()), "error", err).Debug("error in do get")
		return err
	}
	rl.With("status", resp.StatusCode(), "body", string(resp.Body())).Debug("got response")
	if resp.StatusCode() != http.StatusOK {
		return errors.New(resp.Status())
	}
	return nil
}

func (cl *client) doPost(req *resty.Request, url string) error {
	rl := cl.log.With("method", "POST", "path", url)
	if cl.client.UserInfo != nil {
		rl = rl.With("username", cl.client.UserInfo.Username)
	}
	rl.Debug("do post")
	resp, err := req.Post(url)
	if err != nil {
		rl.With("status", resp.StatusCode(), "body", string(resp.Body()), "error", err).Debug("error in do post")
		return err
	}
	rl.With("status", resp.StatusCode(), "body", string(resp.Body())).Debug("got response")
	if resp.StatusCode() != http.StatusOK {
		return errors.New(resp.Status())
	}
	return nil
}

func (cl *client) Status() (*types.Status, error) {
	status := &types.Status{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(status), "status")
	return status, err
}

func (cl *client) RewriteList() (*types.RewriteEntries, error) {
	rewrites := &types.RewriteEntries{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(&rewrites), "/rewrite/list")
	return rewrites, err
}

func (cl *client) AddRewriteEntries(entries ...types.RewriteEntry) error {
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

func (cl *client) DeleteRewriteEntries(entries ...types.RewriteEntry) error {
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

func (cl *client) SafeSearch() (bool, error) {
	return cl.toggleStatus("safesearch")
}

func (cl *client) ToggleSafeSearch(enable bool) error {
	return cl.toggleBool("safesearch", enable)
}

func (cl *client) toggleStatus(mode string) (bool, error) {
	fs := &types.EnableConfig{}
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

func (cl *client) Filtering() (*types.FilteringStatus, error) {
	f := &types.FilteringStatus{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(f), "/filtering/status")
	return f, err
}

func (cl *client) AddFilters(whitelist bool, filters ...types.Filter) error {
	for _, f := range filters {
		cl.log.With("url", f.URL, "whitelist", whitelist, "enabled", f.Enabled).Info("Add filter")
		ff := &types.Filter{Name: f.Name, URL: f.URL, Whitelist: whitelist}
		err := cl.doPost(cl.client.R().EnableTrace().SetBody(ff), "/filtering/add_url")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) DeleteFilters(whitelist bool, filters ...types.Filter) error {
	for _, f := range filters {
		cl.log.With("url", f.URL, "whitelist", whitelist, "enabled", f.Enabled).Info("Delete filter")
		ff := &types.Filter{URL: f.URL, Whitelist: whitelist}
		err := cl.doPost(cl.client.R().EnableTrace().SetBody(ff), "/filtering/remove_url")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) UpdateFilters(whitelist bool, filters ...types.Filter) error {
	for _, f := range filters {
		cl.log.With("url", f.URL, "whitelist", whitelist, "enabled", f.Enabled).Info("Update filter")
		fu := &types.FilterUpdate{Whitelist: whitelist, URL: f.URL, Data: types.Filter{ID: f.ID, Name: f.Name, URL: f.URL, Whitelist: whitelist, Enabled: f.Enabled}}
		err := cl.doPost(cl.client.R().EnableTrace().SetBody(fu), "/filtering/set_url")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) RefreshFilters(whitelist bool) error {
	cl.log.With("whitelist", whitelist).Info("Refresh filter")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(&types.RefreshFilter{Whitelist: whitelist}), "/filtering/refresh")
}

func (cl *client) ToggleProtection(enable bool) error {
	cl.log.With("enable", enable).Info("Toggle protection")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(&types.Protection{ProtectionEnabled: enable}), "/dns_config")
}

func (cl *client) SetCustomRules(rules types.UserRules) error {
	cl.log.With("rules", len(rules)).Info("Set user rules")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(rules.String()), "/filtering/set_rules")
}

func (cl *client) ToggleFiltering(enabled bool, interval float64) error {
	cl.log.With("enabled", enabled, "interval", interval).Info("Toggle filtering")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(&types.FilteringConfig{
		EnableConfig:   types.EnableConfig{Enabled: enabled},
		IntervalConfig: types.IntervalConfig{Interval: interval},
	}), "/filtering/config")
}

func (cl *client) Services() (types.Services, error) {
	svcs := types.Services{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(&svcs), "/blocked_services/list")
	return svcs, err
}

func (cl *client) SetServices(services types.Services) error {
	cl.log.With("services", len(services)).Info("Set services")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(&services), "/blocked_services/set")
}

func (cl *client) Clients() (*types.Clients, error) {
	clients := &types.Clients{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(clients), "/clients")
	return clients, err
}

func (cl *client) AddClients(clients ...types.Client) error {
	for i := range clients {
		client := clients[i]
		cl.log.With("name", client.Name).Info("Add client")
		err := cl.doPost(cl.client.R().EnableTrace().SetBody(&client), "/clients/add")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) UpdateClients(clients ...types.Client) error {
	for _, client := range clients {
		cl.log.With("name", client.Name).Info("Update client")
		err := cl.doPost(cl.client.R().EnableTrace().SetBody(&types.ClientUpdate{Name: client.Name, Data: client}), "/clients/update")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) DeleteClients(clients ...types.Client) error {
	for i := range clients {
		client := clients[i]
		cl.log.With("name", client.Name).Info("Delete client")
		err := cl.doPost(cl.client.R().EnableTrace().SetBody(&client), "/clients/delete")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) QueryLogConfig() (*types.QueryLogConfig, error) {
	qlc := &types.QueryLogConfig{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(qlc), "/querylog_info")
	return qlc, err
}

func (cl *client) SetQueryLogConfig(enabled bool, interval float64, anonymizeClientIP bool) error {
	cl.log.With("enabled", enabled, "interval", interval, "anonymizeClientIP", anonymizeClientIP).Info("Set query log config")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(&types.QueryLogConfig{
		EnableConfig:      types.EnableConfig{Enabled: enabled},
		IntervalConfig:    types.IntervalConfig{Interval: interval},
		AnonymizeClientIP: anonymizeClientIP,
	}), "/querylog_config")
}

func (cl *client) StatsConfig() (*types.IntervalConfig, error) {
	stats := &types.IntervalConfig{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(stats), "/stats_info")
	return stats, err
}

func (cl *client) SetStatsConfig(interval float64) error {
	cl.log.With("interval", interval).Info("Set stats config")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(&types.IntervalConfig{Interval: interval}), "/stats_config")
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

func (cl *client) AccessList() (*types.AccessList, error) {
	al := &types.AccessList{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(al), "/access/list")
	return al, err
}

func (cl *client) SetAccessList(list *types.AccessList) error {
	cl.log.Info("Set access list")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(list), "/access/set")
}

func (cl *client) DNSConfig() (*types.DNSConfig, error) {
	cfg := &types.DNSConfig{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(cfg), "/dns_info")
	return cfg, err
}

func (cl *client) SetDNSConfig(config *types.DNSConfig) error {
	cl.log.Info("Set dns config list")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(config), "/dns_config")
}

func (cl *client) DHCPServerConfig() (*types.DHCPServerConfig, error) {
	cfg := &types.DHCPServerConfig{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(cfg), "/dhcp/status")
	return cfg, err
}

func (cl *client) SetDHCPServerConfig(config *types.DHCPServerConfig) error {
	cl.log.Info("Set dhcp server config")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(config), "/dhcp/set_config")
}

func (cl *client) AddDHCPStaticLeases(leases ...types.Lease) error {
	for _, l := range leases {
		cl.log.With("mac", l.HWAddr, "ip", l.IP, "hostname", l.Hostname).Info("Add static dhcp lease")
		err := cl.doPost(cl.client.R().EnableTrace().SetBody(l), "/dhcp/add_static_lease")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) DeleteDHCPStaticLeases(leases ...types.Lease) error {
	for _, l := range leases {
		cl.log.With("mac", l.HWAddr, "ip", l.IP, "hostname", l.Hostname).Info("Delete static dhcp lease")
		err := cl.doPost(cl.client.R().EnableTrace().SetBody(l), "/dhcp/remove_static_lease")
		if err != nil {
			return err
		}
	}
	return nil
}
