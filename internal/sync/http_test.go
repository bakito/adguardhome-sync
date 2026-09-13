package sync

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	gm "go.uber.org/mock/gomock"

	"github.com/bakito/adguardhome-sync/internal/client"
	"github.com/bakito/adguardhome-sync/internal/client/model"
	clientmock "github.com/bakito/adguardhome-sync/internal/mocks/client"
	"github.com/bakito/adguardhome-sync/internal/types"
)

func TestPercent(t *testing.T) {
	tests := []struct {
		name string
		a    *int
		b    *int
		want string
	}{
		{name: "both inputs are nil", a: nil, b: nil, want: "0.00"},
		{name: "a is nil, b is non-zero", a: nil, b: new(10), want: "0.00"},
		{name: "b is nil, a is non-zero", a: new(10), b: nil, want: "0.00"},
		{name: "b is zero", a: new(10), b: new(0), want: "0.00"},
		{name: "normal case with positive int values", a: new(25), b: new(100), want: "25.00"},
		{name: "a and b are equal", a: new(50), b: new(50), want: "100.00"},
		{name: "a is zero, b is positive", a: new(0), b: new(50), want: "0.00"},
		{name: "large positive values", a: new(1000), b: new(4000), want: "25.00"},
		{name: "a greater than b", a: new(150), b: new(100), want: "150.00"},
		{name: "negative values for a and b", a: new(-25), b: new(-50), want: "50.00"},
		{name: "a is positive, b is negative", a: new(25), b: new(-50), want: "-50.00"},
		{name: "a is negative, b is positive", a: new(-25), b: new(50), want: "-50.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := percent(tt.a, tt.b); got != tt.want {
				t.Errorf("percent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorker_handleLivez(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := &worker{}

	endpoints := []string{"/livez", "/liveness", "/healthz"}
	for _, ep := range endpoints {
		for _, method := range []string{http.MethodGet, http.MethodHead} {
			t.Run(method+" "+ep, func(t *testing.T) {
				r := gin.New()
				r.GET("/livez", w.handleLivez)
				r.HEAD("/livez", w.handleLivez)
				r.GET("/liveness", w.handleLivez)
				r.HEAD("/liveness", w.handleLivez)
				r.GET("/healthz", w.handleLivez)
				r.HEAD("/healthz", w.handleLivez)

				req := httptest.NewRequestWithContext(t.Context(), method, ep, http.NoBody)
				rec := httptest.NewRecorder()
				r.ServeHTTP(rec, req)

				if rec.Code != http.StatusOK {
					t.Errorf("%s %s status = %v, want %v", method, ep, rec.Code, http.StatusOK)
				}
			})
		}
	}
}

func TestWorker_handleReadyz(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return 200 OK when origin and replicas are success", func(t *testing.T) {
		ctrl := gm.NewController(t)
		cl := clientmock.NewMockClient(ctrl)
		cl.EXPECT().Status().Return(&model.ServerStatus{ProtectionEnabled: true}, nil).AnyTimes()

		w := &worker{
			createClient: func(_ types.AdGuardInstance, _ time.Duration) (client.Client, error) {
				return cl, nil
			},
			cfg: &types.Config{
				Origin:   &types.AdGuardInstance{WebURL: "http://origin"},
				Replicas: []types.AdGuardInstance{{WebURL: "http://replica1"}},
			},
		}

		r := gin.New()
		r.GET("/readyz", w.handleReadyz)
		r.GET("/readiness", w.handleReadyz)

		for _, ep := range []string{"/readyz", "/readiness"} {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, ep, http.NoBody)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("GET %s status = %v, want %v", ep, rec.Code, http.StatusOK)
			}
		}
	})

	t.Run("should return 503 when origin fails", func(t *testing.T) {
		ctrl := gm.NewController(t)
		cl := clientmock.NewMockClient(ctrl)
		cl.EXPECT().Status().Return(nil, errors.New("connection refused")).AnyTimes()

		w := &worker{
			createClient: func(_ types.AdGuardInstance, _ time.Duration) (client.Client, error) {
				return cl, nil
			},
			cfg: &types.Config{
				Origin:   &types.AdGuardInstance{WebURL: "http://origin"},
				Replicas: []types.AdGuardInstance{{WebURL: "http://replica1"}},
			},
		}

		r := gin.New()
		r.GET("/readyz", w.handleReadyz)

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/readyz", http.NoBody)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("GET /readyz status = %v, want %v", rec.Code, http.StatusServiceUnavailable)
		}
	})

	t.Run("should return 503 when replica fails", func(t *testing.T) {
		ctrl := gm.NewController(t)
		clOrigin := clientmock.NewMockClient(ctrl)
		clOrigin.EXPECT().Status().Return(&model.ServerStatus{ProtectionEnabled: true}, nil).AnyTimes()

		clReplica := clientmock.NewMockClient(ctrl)
		clReplica.EXPECT().Status().Return(nil, errors.New("replica down")).AnyTimes()

		w := &worker{
			createClient: func(inst types.AdGuardInstance, _ time.Duration) (client.Client, error) {
				if inst.WebURL == "http://origin" {
					return clOrigin, nil
				}
				return clReplica, nil
			},
			cfg: &types.Config{
				Origin:   &types.AdGuardInstance{WebURL: "http://origin"},
				Replicas: []types.AdGuardInstance{{WebURL: "http://replica1"}},
			},
		}

		r := gin.New()
		r.GET("/readyz", w.handleReadyz)

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/readyz", http.NoBody)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("GET /readyz status = %v, want %v", rec.Code, http.StatusServiceUnavailable)
		}
	})
}
