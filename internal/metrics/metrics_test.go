package metrics

import (
	"reflect"
	"testing"

	"github.com/go-faker/faker/v4"

	"github.com/bakito/adguardhome-sync/internal/client/model"
)

func ptr[T any](v T) *T {
	return &v
}

func TestMetrics_UpdateInstances_getStats(t *testing.T) {
	stats = make(OverallStats)
	UpdateInstances(InstanceMetricsList{[]InstanceMetrics{
		{HostName: "foo", Status: &model.ServerStatus{}, Stats: &model.Stats{
			NumDnsQueries: ptr(100),
			DnsQueries: ptr(
				[]int{10, 20, 30, 40, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			),
		}},
		{HostName: "bar", Status: &model.ServerStatus{}, Stats: &model.Stats{
			NumDnsQueries: ptr(200),
			DnsQueries: ptr(
				[]int{20, 40, 60, 80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			),
		}},
		{HostName: "aaa", Status: &model.ServerStatus{}, Stats: &model.Stats{
			NumDnsQueries: ptr(300),
			DnsQueries: ptr(
				[]int{30, 60, 90, 120, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			),
		}},
	}})

	if _, ok := stats["foo"]; !ok {
		t.Error("expected key 'foo' in stats")
	}
	if *stats["foo"].NumDnsQueries != 100 {
		t.Errorf("expected 100 but got %d", *stats["foo"].NumDnsQueries)
	}
	if _, ok := stats["bar"]; !ok {
		t.Error("expected key 'bar' in stats")
	}
	if *stats["bar"].NumDnsQueries != 200 {
		t.Errorf("expected 200 but got %d", *stats["bar"].NumDnsQueries)
	}
	if _, ok := stats["aaa"]; !ok {
		t.Error("expected key 'aaa' in stats")
	}
	if *stats["aaa"].NumDnsQueries != 300 {
		t.Errorf("expected 300 but got %d", *stats["aaa"].NumDnsQueries)
	}

	os := getStats()
	tot := os.Total()
	if *tot.NumDnsQueries != 600 {
		t.Errorf("expected 600 but got %d", *tot.NumDnsQueries)
	}
	expectedQueries := []int{60, 120, 180, 240, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	if !reflect.DeepEqual(*tot.DnsQueries, expectedQueries) {
		t.Errorf("expected %v but got %v", expectedQueries, *tot.DnsQueries)
	}

	foo := os["foo"]
	bar := os["bar"]
	aaa := os["aaa"]

	if *foo.NumDnsQueries != 100 {
		t.Errorf("expected 100 but got %d", *foo.NumDnsQueries)
	}
	if *bar.NumDnsQueries != 200 {
		t.Errorf("expected 200 but got %d", *bar.NumDnsQueries)
	}
	if *aaa.NumDnsQueries != 300 {
		t.Errorf("expected 300 but got %d", *aaa.NumDnsQueries)
	}
}

func TestMetrics_StatsGraph(t *testing.T) {
	stats = make(OverallStats)
	var metrics InstanceMetricsList
	err := faker.FakeData(&metrics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	UpdateInstances(metrics)

	_, dns, blocked, malware, adult := StatsGraph()

	verifyStats(t, dns)
	verifyStats(t, blocked)
	verifyStats(t, malware)
	verifyStats(t, adult)
}

func verifyStats(t *testing.T, lines []Line) {
	t.Helper()
	var total Line
	sum := make([]int, len(lines[0].Data))
	for _, l := range lines {
		if l.Title == labelTotal {
			total = l
		} else {
			for i, d := range l.Data {
				sum[i] += d
			}
		}
	}
	if !reflect.DeepEqual(sum, total.Data) {
		t.Errorf("expected sum %v but got %v", total.Data, sum)
	}
}
