package client

import (
	"net/http"

	"github.com/go-resty/resty/v2"

	"github.com/bakito/adguardhome-sync/pkg/client/model"
)

var _ model.HttpRequestDoer = &adapter{}

func RestyAdapter(r *resty.Client) model.HttpRequestDoer {
	return &adapter{
		client: r,
	}
}

type adapter struct {
	client *resty.Client
}

func (a adapter) Do(req *http.Request) (*http.Response, error) {
	r, err := a.client.R().
		SetHeaderMultiValues(req.Header).
		Execute(req.Method, req.URL.String())
	return r.RawResponse, err
}
