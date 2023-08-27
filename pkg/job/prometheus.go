package job

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	integerBase int = 10
)

// ServeHTTP writes promtheues styled metrics about the last executed `mtr`
// run, see https://prometheus.io/docs/instrumenting/exposition_formats/#line-format
//
// NOTE: at the moment, no use of github.com/prometheus/client_golang/prometheus
// because overhead in size and complexity. once mtr-exporter requires features
// like push-gateway-export or graphite export or the like, we switch.
func (c *Collector) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.jobs) == 0 {
		fmt.Fprintln(w, "# no mtr jobs defined (yet).")
		return
	}

	fmt.Fprintf(w, "# %d mtr jobs defined\n", len(c.jobs))

	for _, job := range c.jobs {

		if len(job.Report.Hubs) == 0 {
			continue
		}

		// the original job.Report might be changed in the
		// background by a successful run of mtr. copy (pointer to) the report
		// to have something safe to work on
		report := job.Report
		ts := job.Launched.UTC()
		d := job.Duration

		labels := report.Mtr.Labels()
		labels["mtr_exporter_job"] = job.Label
		tsMs := ts.UnixNano() / int64(time.Millisecond)

		fmt.Fprintf(w, "# mtr run: %s\n", ts.Format(time.RFC3339Nano))
		// FIXME: remove or rewrite: fmt.Fprintf(w, "# cmdline: %s\n", job.cmdLine)
		fmt.Fprintf(w, "mtr_report_duration_ms_gauge{%s} %d %d\n",
			labels2Prom(labels), d/time.Millisecond, tsMs)
		fmt.Fprintf(w, "mtr_report_count_hubs_gauge{%s} %d %d\n",
			labels2Prom(labels), len(report.Hubs), tsMs)

		for i, hub := range report.Hubs {
			labels["host"] = hub.Host
			labels["count"] = strconv.FormatInt(hub.Count, integerBase)
			// mark last hub to have it easily identified
			if i < (len(report.Hubs) - 1) {
				hub.WriteMetrics(w, labels2Prom(labels), tsMs)
			} else {
				labels["last"] = "true"
				hub.WriteMetrics(w, labels2Prom(labels), tsMs)
				delete(labels, "last")
			}
		}
	}

}

func labels2Prom(labels map[string]string) string {
	sl := make(sort.StringSlice, 0, len(labels))
	for k, v := range labels {
		sl = append(sl, fmt.Sprintf("%s=%q", k, v))
	}
	sl.Sort()
	return strings.Join(sl, ",")
}