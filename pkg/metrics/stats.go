package metrics

import (
	"slices"
	"strings"

	"github.com/bakito/adguardhome-sync/pkg/client/model"
)

const labelTotal = "Total"

var (
	blue             = []int{78, 141, 245}
	blueAlternatives = [][]int{
		{44, 95, 163},
		{122, 166, 247},
		{30, 61, 92},
		{93, 158, 255},
		{58, 123, 213},
	}

	red             = []int{255, 94, 94}
	redAlternatives = [][]int{
		{204, 59, 59},
		{255, 127, 127},
		{140, 36, 36},
		{255, 153, 153},
		{255, 66, 66},
	}

	yellow             = []int{232, 198, 78}
	yellowAlternatives = [][]int{
		{196, 163, 60},
		{255, 220, 110},
		{140, 114, 36},
		{250, 233, 156},
		{212, 180, 84},
	}

	green             = []int{110, 224, 122}
	greenAlternatives = [][]int{
		{68, 160, 80},
		{142, 255, 158},
		{44, 140, 63},
		{163, 255, 192},
		{85, 198, 102},
	}
)

func StatsGraph() (t *model.Stats, dns, blocked, malware, adult []Line) {
	s := getStats()
	t = s.Total()
	dns = graphLines(t, s, blue, blueAlternatives, func(s *model.Stats) []int {
		return safeStats(s.DnsQueries)
	})
	blocked = graphLines(t, s, red, redAlternatives, func(s *model.Stats) []int {
		return safeStats(s.BlockedFiltering)
	})
	malware = graphLines(t, s, green, greenAlternatives, func(s *model.Stats) []int {
		return safeStats(s.ReplacedSafebrowsing)
	})
	adult = graphLines(t, s, yellow, yellowAlternatives, func(s *model.Stats) []int {
		return safeStats(s.ReplacedParental)
	})

	return t, dns, blocked, malware, adult
}

func safeStats(stats *[]int) []int {
	if stats == nil {
		return make([]int, 0)
	}
	return *stats
}

func graphLines(
	t *model.Stats,
	s OverallStats,
	baseColor []int,
	altColors [][]int,
	dataCB func(s *model.Stats) []int,
) []Line {
	g := &graph{
		total: Line{
			Fill:  true,
			Title: labelTotal,
			Data:  dataCB(t),
			R:     baseColor[0],
			G:     baseColor[1],
			B:     baseColor[2],
		},
	}

	var i int
	for name, data := range s {
		if name != StatsTotal {
			g.replicas = append(g.replicas, Line{
				Fill:  false,
				Title: name,
				Data:  dataCB(data),
				R:     altColors[i%len(altColors)][0],
				G:     altColors[i%len(altColors)][1],
				B:     altColors[i%len(altColors)][2],
			})
			i++
		}
	}

	lines := []Line{g.total}

	slices.SortFunc(g.replicas, func(a, b Line) int {
		return strings.Compare(a.Title, b.Title)
	})
	lines = append(lines, g.replicas...)
	return lines
}

type graph struct {
	total    Line
	replicas []Line
}

type Line struct {
	Data  []int  `json:"data"`
	R     int    `json:"r"`
	G     int    `json:"g"`
	B     int    `json:"b"`
	Title string `json:"title"`
	Fill  bool   `json:"fill"`
}
