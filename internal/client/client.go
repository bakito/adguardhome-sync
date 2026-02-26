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

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"github.com/bakito/adguardhome-sync/internal/client/model"
	"github.com/bakito/adguardhome-sync/internal/log"
	"github.com/bakito/adguardhome-sync/internal/types"
)

const envRedirectPolicyNoOfRedirects = "REDIRECT_POLICY_NO_OF_REDIRECTS"

type Error struct {
	message   string
	errorCode int
}

func (e *Error) Error() string {
	return e.message
}

func (e *Error) Code() int {
	return e.errorCode
}

var (
	l = log.GetLogger("client")
	// ErrSetupNeeded custom error.
	ErrSetupNeeded = errors.New("setup needed")
)

func detailedError(resp *resty.Response, err error) error {
	e := resp.Status()
	if len(resp.Body()) > 0 {
		e += fmt.Sprintf("(%s)", string(resp.Body()))
	}
	if err != nil {
		e += ": " + err.Error()
	}
	return &Error{
		message:   e,
		errorCode: resp.StatusCode(),
	}
}

// New create a new client.
func New(config types.AdGuardInstance) (Client, error) {
	var apiURL string
	if config.APIPath == "" {
		apiURL = config.URL + "/control"
	} else {
		apiURL = fmt.Sprintf("%s/%s", config.URL, config.APIPath)
	}
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Clean(u.Path)
	cl := resty.New().SetBaseURL(u.String()).SetDisableWarn(true).SetHeaders(config.RequestHeaders)

	// #nosec G402 has to be explicitly enabled
	cl.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: config.InsecureSkipVerify})

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

// Client AdguardHome API client interface.
//
//nolint:interfacebloat
type Client interface {
	Host() string
	Status() (*model.ServerStatus, error)
	Stats() (*model.Stats, error)
	QueryLog(limit int) (*model.QueryLog, error)
	ToggleProtection(enable bool) error
	RewriteList() (*model.RewriteEntries, error)
	AddRewriteEntries(e ...model.RewriteEntry) error
	DeleteRewriteEntries(e ...model.RewriteEntry) error
	Filtering() (*model.FilterStatus, error)
	ToggleFiltering(enabled bool, interval int) error
	AddFilter(whitelist bool, f model.Filter) error
	DeleteFilter(whitelist bool, f model.Filter) error
	UpdateFilter(whitelist bool, f model.Filter) error
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
	BlockedServicesSchedule() (*model.BlockedServicesSchedule, error)
	SetBlockedServicesSchedule(schedule *model.BlockedServicesSchedule) error
	Clients() (*model.Clients, error)
	AddClient(client *model.Client) error
	UpdateClient(client *model.Client) error
	DeleteClient(client *model.Client) error
	QueryLogConfig() (*model.QueryLogConfigWithIgnored, error)
	SetQueryLogConfig(ql *model.QueryLogConfigWithIgnored) error
	StatsConfig() (*model.GetStatsConfigResponse, error)
	SetStatsConfig(sc *model.PutStatsConfigUpdateRequest) error
	Setup() error
	AccessList() (*model.AccessList, error)
	SetAccessList(accessList *model.AccessList) error
	DNSConfig() (*model.DNSConfig, error)
	SetDNSConfig(config *model.DNSConfig) error
	DhcpConfig() (*model.DhcpStatus, error)
	SetDhcpConfig(status *model.DhcpStatus) error
	AddDHCPStaticLease(lease model.DhcpStaticLease) error
	DeleteDHCPStaticLease(lease model.DhcpStaticLease) error
	TLSConfig() (*model.TlsConfig, error)
	SetTLSConfig(tls *model.TlsConfig) error
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

func (cl *client) Stats() (*model.Stats, error) {
	stats := &model.Stats{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(stats), "stats")
	return stats, err
}

func (cl *client) QueryLog(limit int) (*model.QueryLog, error) {
	ql := &model.QueryLog{}
	err := cl.doGet(
		cl.client.R().EnableTrace().SetResult(ql),
		fmt.Sprintf(`querylog?limit=%d&response_status="all"`, limit),
	)
	return ql, err
}

func (cl *client) RewriteList() (*model.RewriteEntries, error) {
	rewrites := &model.RewriteEntries{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(&rewrites), "/rewrite/list")
	return rewrites, err
}

func (cl *client) AddRewriteEntries(entries ...model.RewriteEntry) error {
	for _, e := range entries {
		cl.log.With("domain", e.Domain, "answer", e.Answer).Info("Add DNS rewrite entry")
		err := cl.doPost(cl.client.R().EnableTrace().SetBody(&e), "/rewrite/add")
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *client) DeleteRewriteEntries(entries ...model.RewriteEntry) error {
	for _, e := range entries {
		cl.log.With("domain", e.Domain, "answer", e.Answer).Info("Delete DNS rewrite entry")
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
	cl.log.With("enable", enable).Info("Toggle " + mode)
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

func (cl *client) AddFilter(whitelist bool, f model.Filter) error {
	cl.log.With("url", f.Url, "whitelist", whitelist, "enabled", f.Enabled).Info("Add filter")
	ff := &model.AddUrlRequest{Name: new(f.Name), Url: new(f.Url), Whitelist: new(whitelist)}
	return cl.doPost(cl.client.R().EnableTrace().SetBody(ff), "/filtering/add_url")
}

func (cl *client) DeleteFilter(whitelist bool, f model.Filter) error {
	cl.log.With("url", f.Url, "whitelist", whitelist, "enabled", f.Enabled).Info("Delete filter")
	ff := &model.RemoveUrlRequest{Url: new(f.Url), Whitelist: new(whitelist)}
	return cl.doPost(cl.client.R().EnableTrace().SetBody(ff), "/filtering/remove_url")
}

func (cl *client) UpdateFilter(whitelist bool, f model.Filter) error {
	cl.log.With("url", f.Url, "whitelist", whitelist, "enabled", f.Enabled).Info("Update filter")
	fu := &model.FilterSetUrl{
		Whitelist: new(whitelist), Url: new(f.Url),
		Data: &model.FilterSetUrlData{Name: f.Name, Url: f.Url, Enabled: f.Enabled},
	}
	return cl.doPost(cl.client.R().EnableTrace().SetBody(fu), "/filtering/set_url")
}

func (cl *client) RefreshFilters(whitelist bool) error {
	cl.log.With("whitelist", whitelist).Info("Refresh filter")
	return cl.doPost(
		cl.client.R().EnableTrace().SetBody(&model.FilterRefreshRequest{Whitelist: new(whitelist)}),
		"/filtering/refresh",
	)
}

func (cl *client) ToggleProtection(enable bool) error {
	cl.log.With("enable", enable).Info("Toggle protection")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(&types.Protection{ProtectionEnabled: enable}), "/dns_config")
}

func (cl *client) SetCustomRules(rules *[]string) error {
	var l int
	if rules != nil {
		l = len(*rules)
	}
	cl.log.With("rules", l).Info("Set user rules")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(&model.SetRulesRequest{Rules: rules}), "/filtering/set_rules")
}

func (cl *client) ToggleFiltering(enabled bool, interval int) error {
	cl.log.With("enabled", enabled, "interval", interval).Info("Toggle filtering")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(&model.FilterConfig{
		Enabled:  new(enabled),
		Interval: new(interval),
	}), "/filtering/config")
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

func (cl *client) AddClient(client *model.Client) error {
	cl.log.With("name", *client.Name).Info("Add client settings")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(client), "/clients/add")
}

func (cl *client) UpdateClient(client *model.Client) error {
	cl.log.With("name", *client.Name).Info("Update client settings")
	return cl.doPost(
		cl.client.R().EnableTrace().SetBody(&model.ClientUpdate{Name: client.Name, Data: client}),
		"/clients/update",
	)
}

func (cl *client) DeleteClient(client *model.Client) error {
	cl.log.With("name", *client.Name).Info("Delete client settings")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(client), "/clients/delete")
}

func (cl *client) QueryLogConfig() (*model.QueryLogConfigWithIgnored, error) {
	qlc := &model.QueryLogConfigWithIgnored{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(qlc), "/querylog/config")
	return qlc, err
}

func (cl *client) SetQueryLogConfig(qlc *model.QueryLogConfigWithIgnored) error {
	cl.log.With("enabled", *qlc.Enabled, "interval", *qlc.Interval, "anonymizeClientIP", *qlc.AnonymizeClientIp).
		Info("Set query log config")
	return cl.doPut(cl.client.R().EnableTrace().SetBody(qlc), "/querylog/config/update")
}

func (cl *client) StatsConfig() (*model.GetStatsConfigResponse, error) {
	stats := &model.GetStatsConfigResponse{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(stats), "/stats/config")
	return stats, err
}

func (cl *client) SetStatsConfig(sc *model.PutStatsConfigUpdateRequest) error {
	cl.log.With("interval", sc.Interval).Info("Set stats config")
	return cl.doPut(cl.client.R().EnableTrace().SetBody(sc), "/stats/config/update")
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

func (cl *client) AddDHCPStaticLease(l model.DhcpStaticLease) error {
	cl.log.With("mac", l.Mac, "ip", l.Ip, "hostname", l.Hostname).Info("Add static dhcp lease")
	err := cl.doPost(cl.client.R().EnableTrace().SetBody(l), "/dhcp/add_static_lease")
	if err != nil {
		return err
	}
	return nil
}

func (cl *client) DeleteDHCPStaticLease(l model.DhcpStaticLease) error {
	cl.log.With("mac", l.Mac, "ip", l.Ip, "hostname", l.Hostname).Info("Delete static dhcp lease")
	err := cl.doPost(cl.client.R().EnableTrace().SetBody(l), "/dhcp/remove_static_lease")
	if err != nil {
		return err
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

func (cl *client) TLSConfig() (*model.TlsConfig, error) {
	tlsc := &model.TlsConfig{}
	err := cl.doGet(cl.client.R().EnableTrace().SetResult(tlsc), "/tls/status")
	return tlsc, err
}

func (cl *client) SetTLSConfig(tlsc *model.TlsConfig) error {
	cl.log.With("enabled", tlsc.Enabled).Info("Set TLS config")
	return cl.doPost(cl.client.R().EnableTrace().SetBody(tlsc), "/tls/configure")
}
