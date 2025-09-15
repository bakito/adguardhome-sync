package sync

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Percent", func() {
	DescribeTable("calculating percentage",
		func(a, b *int, want string) {
			Expect(percent(a, b)).To(Equal(want))
		},
		Entry("both inputs are nil", nil, nil, "0.00"),
		Entry("a is nil, b is non-zero", nil, intPtr(10), "0.00"),
		Entry("b is nil, a is non-zero", intPtr(10), nil, "0.00"),
		Entry("b is zero", intPtr(10), intPtr(0), "0.00"),
		Entry("normal case with positive int values", intPtr(25), intPtr(100), "25.00"),
		Entry("a and b are equal", intPtr(50), intPtr(50), "100.00"),
		Entry("a is zero, b is positive", intPtr(0), intPtr(50), "0.00"),
		Entry("large positive values", intPtr(1000), intPtr(4000), "25.00"),
		Entry("a greater than b", intPtr(150), intPtr(100), "150.00"),
		Entry("negative values for a and b", intPtr(-25), intPtr(-50), "50.00"),
		Entry("a is positive, b is negative", intPtr(25), intPtr(-50), "-50.00"),
		Entry("a is negative, b is positive", intPtr(-25), intPtr(50), "-50.00"),
	)
})

func intPtr(i int) *int {
	return &i
}
