package client

import (
	"encoding/json"
	"net/http"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

func (cl *client) doGet(req *resty.Request, url string) error {
	rl := cl.log.With("method", "GET", "path", url)
	if cl.client.UserInfo != nil {
		rl = rl.With("username", cl.client.UserInfo.Username)
	}
	req.ForceContentType("application/json")
	rl.Debug("do get")
	resp, err := req.Get(url)
	if err != nil {
		l := rl
		if resp != nil {
			if resp.StatusCode() == http.StatusFound {
				loc := resp.Header().Get("Location")
				if loc == "/install.html" || loc == "/control/install.html" {
					return ErrSetupNeeded
				}
			}
			l = l.With("status", resp.StatusCode(), "body", string(resp.Body()), "error", err)
		}

		l.Debug("error in do get")
		return detailedError(resp, err)
	}

	checkAuthenticationIssue(resp, rl)

	rl.With(
		"status", resp.StatusCode(),
		"body", string(resp.Body()),
		"content-type", resp.Header()["Content-Type"],
	).Debug("got response")
	if resp.StatusCode() != http.StatusOK {
		return detailedError(resp, nil)
	}
	return nil
}

func (cl *client) doPost(req *resty.Request, url string) error {
	rl := cl.log.With("method", "POST", "path", url)
	if cl.client.UserInfo != nil {
		rl = rl.With("username", cl.client.UserInfo.Username)
	}
	b, _ := json.Marshal(req.Body)
	rl.With("body", string(b)).Debug("do post")
	resp, err := req.Post(url)
	if err != nil {
		rl.With("status", resp.StatusCode(), "body", string(resp.Body()), "error", err).Debug("error in do post")
		return detailedError(resp, err)
	}

	checkAuthenticationIssue(resp, rl)

	rl.With(
		"status", resp.StatusCode(),
		"body", string(resp.Body()),
		"content-type", contentType(resp),
	).Debug("got response")
	if resp.StatusCode() != http.StatusOK {
		return detailedError(resp, nil)
	}
	return nil
}

func (cl *client) doPut(req *resty.Request, url string) error {
	rl := cl.log.With("method", "PUT", "path", url)
	if cl.client.UserInfo != nil {
		rl = rl.With("username", cl.client.UserInfo.Username)
	}
	b, _ := json.Marshal(req.Body)
	rl.With("body", string(b)).Debug("do put")
	resp, err := req.Put(url)
	if err != nil {
		rl.With("status", resp.StatusCode(), "body", string(resp.Body()), "error", err).Debug("error in do put")
		return detailedError(resp, err)
	}

	checkAuthenticationIssue(resp, rl)

	rl.With(
		"status", resp.StatusCode(),
		"body", string(resp.Body()),
		"content-type", contentType(resp),
	).Debug("got response")
	if resp.StatusCode() != http.StatusOK {
		return detailedError(resp, nil)
	}
	return nil
}

func checkAuthenticationIssue(resp *resty.Response, rl *zap.SugaredLogger) {
	if resp != nil && (resp.StatusCode() == http.StatusUnauthorized || resp.StatusCode() == http.StatusForbidden) {
		rl.With("status", resp.StatusCode()).Error("there seems to be an authentication issue - " +
			"please check https://github.com/bakito/adguardhome-sync/wiki/FAQ")
	}
}
