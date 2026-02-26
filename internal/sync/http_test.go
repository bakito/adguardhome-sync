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
		Entry("a is nil, b is non-zero", nil, new(10), "0.00"),
		Entry("b is nil, a is non-zero", new(10), nil, "0.00"),
		Entry("b is zero", new(10), new(0), "0.00"),
		Entry("normal case with positive int values", new(25), new(100), "25.00"),
		Entry("a and b are equal", new(50), new(50), "100.00"),
		Entry("a is zero, b is positive", new(0), new(50), "0.00"),
		Entry("large positive values", new(1000), new(4000), "25.00"),
		Entry("a greater than b", new(150), new(100), "150.00"),
		Entry("negative values for a and b", new(-25), new(-50), "50.00"),
		Entry("a is positive, b is negative", new(25), new(-50), "-50.00"),
		Entry("a is negative, b is positive", new(-25), new(50), "-50.00"),
	)
})
