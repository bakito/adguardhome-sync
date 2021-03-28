package client

import (
	"fmt"
	"github.com/bakito/adguardhome-sync/pkg/log"
	"github.com/bakito/adguardhome-sync/pkg/types"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"net/url"
)

var (
	l = log.GetLogger("client")
)

func New(apiURL string, username string, password string) (Client, error) {

	cl := resty.New().SetHostURL(apiURL).SetDisableWarn(true)
	if username != "" && password != "" {
		cl = cl.SetBasicAuth(username, password)
	}

	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}

	return &client{
		client: cl,
		log:    l.With("host", u.Host),
	}, nil
}

type Client interface {
	Status() (*types.Status, error)
	RewriteList() (*types.RewriteEntries, error)
	AddRewriteEntries(e ...types.RewriteEntry) error
	DeleteRewriteEntries(e ...types.RewriteEntry) error

	Filtering() (*types.FilteringStatus, error)
	ToggleFiltering(enabled bool, interval int) error
	AddFilters(whitelist bool, e ...types.Filter) error
	DeleteFilters(whitelist bool, e ...types.Filter) error
	RefreshFilters(whitelist bool) error
	SetCustomRules(rules types.UserRules) error

	ToggleSaveBrowsing(enable bool) error
	ToggleParental(enable bool) error
	ToggleSafeSearch(enable bool) error
}

type client struct {
	client *resty.Client
	log    *zap.SugaredLogger
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

func (cl *client) ToggleSaveBrowsing(enable bool) error {
	return cl.toggle("safebrowsing", enable)
}

func (cl *client) ToggleParental(enable bool) error {
	return cl.toggle("parental", enable)
}

func (cl *client) ToggleSafeSearch(enable bool) error {
	return cl.toggle("safesearch", enable)
}

func (cl *client) toggle(mode string, enable bool) error {
	cl.log.With("mode", mode, "enable", enable).Info("Toggle")
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

func (cl *client) SetCustomRules(rules types.UserRules) error {
	cl.log.With("rules", len(rules)).Info("Set user rules")
	_, err := cl.client.R().EnableTrace().SetBody(rules.String()).Post("/filtering/set_rules")
	return err
}

func (cl *client) ToggleFiltering(enabled bool, interval int) error {
	cl.log.With("enabled", enabled, "interval", interval).Info("Toggle filtering")
	_, err := cl.client.R().EnableTrace().SetBody(&types.FilteringConfig{Enabled: enabled, Interval: interval}).Post("/filtering/config")
	return err
}
