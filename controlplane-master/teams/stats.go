package teams

import "github.com/prometheus/client_golang/prometheus"

type collector struct{}

var desc = prometheus.NewDesc(
	"controlplane_teams_available",
	"List of available teams",
	[]string{"team_name", "namespace"},
	nil,
)

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- desc
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	for _, team := range GetAll() {
		m, _ := prometheus.NewConstMetric(desc, prometheus.GaugeValue, 1, team.Name, team.Namespace)
		ch <- m
	}
}

func init() {
	prometheus.Register(&collector{})
}
