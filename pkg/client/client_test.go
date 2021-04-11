package client_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"

	"github.com/bakito/adguardhome-sync/pkg/client"
	"github.com/bakito/adguardhome-sync/pkg/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {

	var (
		cl client.Client
	)

	Context("Filtering", func() {
		It("should reade filtering status", func() {
			cl = Serve("filtering-status.json")
			_, err := cl.Filtering()
			Ω(err).ShouldNot(HaveOccurred())
		})
	})
})

func Serve(file string) client.Client {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadFile(filepath.Join("../../testdata", file))
		Ω(err).ShouldNot(HaveOccurred())
		_, err = w.Write(b)
		Ω(err).ShouldNot(HaveOccurred())
	}))
	cl, err := client.New(types.AdGuardInstance{URL: ts.URL})
	Ω(err).ShouldNot(HaveOccurred())
	return cl
}
