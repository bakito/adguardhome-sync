package metrics

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/google/go-cmp/cmp"

	"github.com/bakito/adguardhome-sync/internal/client/model"
)

func TestUpdateInstances_getStats(t *testing.T) {
	stats = make(OverallStats)
	UpdateInstances(InstanceMetricsList{Metrics: []InstanceMetrics{
		{HostName: "foo", Status: &model.ServerStatus{}, Stats: &model.Stats{
			NumDnsQueries: new(100),
			DnsQueries:    &[]int{10, 20, 30, 40, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}},
		{HostName: "bar", Status: &model.ServerStatus{}, Stats: &model.Stats{
			NumDnsQueries: new(200),
			DnsQueries:    &[]int{20, 40, 60, 80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}},
		{HostName: "aaa", Status: &model.ServerStatus{}, Stats: &model.Stats{
			NumDnsQueries: new(300),
			DnsQueries:    &[]int{30, 60, 90, 120, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}},
	}})

	if _, ok := stats["foo"]; !ok {
		t.Error("stats should have key 'foo'")
	}
	if *stats["foo"].NumDnsQueries != 100 {
		t.Errorf("stats['foo'].NumDnsQueries = %v, want 100", *stats["foo"].NumDnsQueries)
	}
	if _, ok := stats["bar"]; !ok {
		t.Error("stats should have key 'bar'")
	}
	if *stats["bar"].NumDnsQueries != 200 {
		t.Errorf("stats['bar'].NumDnsQueries = %v, want 200", *stats["bar"].NumDnsQueries)
	}
	if _, ok := stats["aaa"]; !ok {
		t.Error("stats should have key 'aaa'")
	}
	if *stats["aaa"].NumDnsQueries != 300 {
		t.Errorf("stats['aaa'].NumDnsQueries = %v, want 300", *stats["aaa"].NumDnsQueries)
	}

	os := getStats()
	tot := os.Total()
	if *tot.NumDnsQueries != 600 {
		t.Errorf("os.Total().NumDnsQueries = %v, want 600", *tot.NumDnsQueries)
	}
	wantDNSQueries := []int{60, 120, 180, 240, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	if diff := cmp.Diff(wantDNSQueries, *tot.DnsQueries); diff != "" {
		t.Errorf("os.Total().DnsQueries mismatch (-want +got):\n%s", diff)
	}

	foo := os["foo"]
	bar := os["bar"]
	aaa := os["aaa"]

	if *foo.NumDnsQueries != 100 {
		t.Errorf("foo.NumDnsQueries = %v, want 100", *foo.NumDnsQueries)
	}
	if *bar.NumDnsQueries != 200 {
		t.Errorf("bar.NumDnsQueries = %v, want 200", *bar.NumDnsQueries)
	}
	if *aaa.NumDnsQueries != 300 {
		t.Errorf("aaa.NumDnsQueries = %v, want 300", *aaa.NumDnsQueries)
	}
}

func TestStatsGraph(t *testing.T) {
	var metrics InstanceMetricsList
	err := faker.FakeData(&metrics)
	if err != nil {
		t.Fatalf("faker.FakeData error = %v", err)
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
	if diff := cmp.Diff(total.Data, sum); diff != "" {
		t.Errorf("sum mismatch (-want +got):\n%s", diff)
	}
}
