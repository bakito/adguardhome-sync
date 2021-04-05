package types_test

import (
	"encoding/json"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/bakito/adguardhome-sync/pkg/types"
)

var _ = Describe("Types", func() {
	Context("FilteringStatus", func() {
		It("should correctly parse json", func() {
			b, err := ioutil.ReadFile("../..//testdata/filtering-status.json")
			fs := &types.FilteringStatus{}
			Ω(err).ShouldNot(HaveOccurred())
			err = json.Unmarshal(b, fs)
			Ω(err).ShouldNot(HaveOccurred())
		})
	})
})
