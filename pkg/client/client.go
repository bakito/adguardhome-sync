package client

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"path"

	"github.com/bakito/adguardhome-sync/pkg/log"
	"github.com/bakito/adguardhome-sync/pkg/types"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

var (
	l = log.GetLogger("client")
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
	cl := resty.New().SetHostURL(u.String()).SetDisableWarn(true)

	if config.InsecureSkipVerify {
		cl.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	if config.Username != "" && config.Password != "" {
		cl = cl.SetBasicAuth(config.Username, config.Password)
	}

	return &client{
		host:   u.Host,
		client: cl,
		log:    l.With("host", u.Host),
	}, nil
}

// Client AdGuard Home API client interface
type Client interface {
	Host() string

	Status() (*types.Status, error)
	ToggleProtection(enable bool) error
	RewriteList() (*types.RewriteEntries, error)
	AddRewriteEntries(e ...types.RewriteEntry) error
	DeleteRewriteEntries(e ...types.RewriteEntry) error

	Filtering() (*types.FilteringStatus, error)
	ToggleFiltering(enabled bool, interval int) error
	AddFilters(whitelist bool, e ...types.Filter) error
	DeleteFilters(whitelist bool, e ...types.Filter) error
	RefreshFilters(whitelist bool) error
	SetCustomRules(rules types.UserRules) error

	SafeBrowsing() (bool, error)
	ToggleSafeBrowsing(enable bool) error
	Parental() (bool, error)
	ToggleParental(enable bool) error
	SafeSearch() (bool, error)
	ToggleSafeSearch(enable bool) error

	Services() (*types.Services, error)
	SetServices(services types.Services) error

	Clients() (*types.Clients, error)
	AddClients(client ...types.Client) error
	UpdateClients(client ...types.Client) error
	DeleteClients(client ...types.Client) error

	QueryLogConfig() (*types.QueryLogConfig, error)
	SetQueryLogConfig(enabled bool, interval int, anonymizeClientIP bool) error
	StatsConfig() (*types.IntervalConfig, error)
	SetStatsConfig(interval int) error
}

type client struct {
	client *resty.Client
	log    *zap.SugaredLogger
	host   string
}

func (cl *client) Host() string {
	return cl.host
}
func (cl *client) Status() (*types.Status, error) {
	status := &types.Status{}
	_, err := cl.client.R().EnableTrace().SetResult(status).Get("status")
	return status, err

}

func (cl *client) RewriteList() (*types.RewriteEntries, error) {
	rewrites := &types.RewriteEntries{}
	_, err := cl.client.R().EnableTrace().SetResult(&rewrites).Get("/rewrite/list")
	return rewrites, err
}

func (cl *client) AddRewriteEntries(entries ...types.RewriteEntry) error {
	for _, e := range entries {
		cl.log.With("domain", e.Domain, "answer", e.Answer).Info("Add rewrite entry")
		_, err := cl.client.R().EnableTrace().SetBody(&e).Post("/rewrite/add")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) DeleteRewriteEntries(entries ...types.RewriteEntry) error {
	for _, e := range entries {
		cl.log.With("domain", e.Domain, "answer", e.Answer).Info("Delete rewrite entry")
		_, err := cl.client.R().EnableTrace().SetBody(&e).Post("/rewrite/delete")
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
	_, err := cl.client.R().EnableTrace().SetResult(fs).Get(fmt.Sprintf("/%s/status", mode))
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
	_, err := cl.client.R().EnableTrace().Post(fmt.Sprintf("/%s/%s", mode, target))
	return err
}

func (cl *client) Filtering() (*types.FilteringStatus, error) {
	f := &types.FilteringStatus{}
	_, err := cl.client.R().EnableTrace().SetResult(f).Get("/filtering/status")
	return f, err
}

func (cl *client) AddFilters(whitelist bool, filters ...types.Filter) error {
	for _, f := range filters {
		cl.log.With("url", f.URL, "whitelist", whitelist).Info("Add filter")
		ff := &types.Filter{Name: f.Name, URL: f.URL, Whitelist: whitelist}
		_, err := cl.client.R().EnableTrace().SetBody(ff).Post("/filtering/add_url")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) DeleteFilters(whitelist bool, filters ...types.Filter) error {
	for _, f := range filters {
		cl.log.With("url", f.URL, "whitelist", whitelist).Info("Delete filter")
		ff := &types.Filter{URL: f.URL, Whitelist: whitelist}
		_, err := cl.client.R().EnableTrace().SetBody(ff).Post("/filtering/remove_url")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) RefreshFilters(whitelist bool) error {
	cl.log.With("whitelist", whitelist).Info("Refresh filter")
	_, err := cl.client.R().EnableTrace().SetBody(&types.RefreshFilter{Whitelist: whitelist}).Post("/filtering/refresh")
	return err
}

func (cl *client) ToggleProtection(enable bool) error {
	cl.log.With("enable", enable).Info("Toggle protection")
	_, err := cl.client.R().EnableTrace().SetBody(&types.Protection{ProtectionEnabled: enable}).Post("/dns_config")
	return err
}

func (cl *client) SetCustomRules(rules types.UserRules) error {
	cl.log.With("rules", len(rules)).Info("Set user rules")
	_, err := cl.client.R().EnableTrace().SetBody(rules.String()).Post("/filtering/set_rules")
	return err
}

func (cl *client) ToggleFiltering(enabled bool, interval int) error {
	cl.log.With("enabled", enabled, "interval", interval).Info("Toggle filtering")
	_, err := cl.client.R().EnableTrace().SetBody(&types.FilteringConfig{
		EnableConfig:   types.EnableConfig{Enabled: enabled},
		IntervalConfig: types.IntervalConfig{Interval: interval},
	}).Post("/filtering/config")
	return err
}

func (cl *client) Services() (*types.Services, error) {
	svcs := &types.Services{}
	_, err := cl.client.R().EnableTrace().SetResult(svcs).Get("/blocked_services/list")
	return svcs, err
}

func (cl *client) SetServices(services types.Services) error {
	cl.log.With("services", len(services)).Info("Set services")
	_, err := cl.client.R().EnableTrace().SetBody(&services).Post("/blocked_services/set")
	return err
}

func (cl *client) Clients() (*types.Clients, error) {
	clients := &types.Clients{}
	_, err := cl.client.R().EnableTrace().SetResult(clients).Get("/clients")
	return clients, err
}

func (cl *client) AddClients(clients ...types.Client) error {
	for _, client := range clients {
		cl.log.With("name", client.Name).Info("Add client")
		_, err := cl.client.R().EnableTrace().SetBody(&client).Post("/clients/add")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) UpdateClients(clients ...types.Client) error {
	for _, client := range clients {
		cl.log.With("name", client.Name).Info("Update client")
		_, err := cl.client.R().EnableTrace().SetBody(&types.ClientUpdate{Name: client.Name, Data: client}).Post("/clients/update")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) DeleteClients(clients ...types.Client) error {
	for _, client := range clients {
		cl.log.With("name", client.Name).Info("Delete client")
		_, err := cl.client.R().EnableTrace().SetBody(&client).Post("/clients/delete")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) QueryLogConfig() (*types.QueryLogConfig, error) {
	qlc := &types.QueryLogConfig{}
	_, err := cl.client.R().EnableTrace().SetResult(qlc).Get("/querylog_info")
	return qlc, err
}

func (cl *client) SetQueryLogConfig(enabled bool, interval int, anonymizeClientIP bool) error {
	cl.log.With("enabled", enabled, "interval", interval, "anonymizeClientIP", anonymizeClientIP).Info("Set query log config")
	_, err := cl.client.R().EnableTrace().SetBody(&types.QueryLogConfig{
		EnableConfig:      types.EnableConfig{Enabled: enabled},
		IntervalConfig:    types.IntervalConfig{Interval: interval},
		AnonymizeClientIP: anonymizeClientIP,
	}).Post("/querylog_config")
	return err
}

func (cl *client) StatsConfig() (*types.IntervalConfig, error) {
	stats := &types.IntervalConfig{}
	_, err := cl.client.R().EnableTrace().SetResult(stats).Get("/stats_info")
	return stats, err
}

func (cl *client) SetStatsConfig(interval int) error {
	cl.log.With("interval", interval).Info("Set stats config")
	_, err := cl.client.R().EnableTrace().SetBody(&types.IntervalConfig{Interval: interval}).Post("/stats_config")
	return err
}
