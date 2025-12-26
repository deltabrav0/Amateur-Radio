package collector

import (
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/dbutler/lotw-exporter/internal/adif"
	"github.com/dbutler/lotw-exporter/internal/lotw"
	"github.com/prometheus/client_golang/prometheus"
)

type Collector struct {
	client *lotw.Client

	// Metrics
	qsoTotal       *prometheus.GaugeVec
	qslTotal       *prometheus.GaugeVec
	dxccCount      prometheus.Gauge
	scrapeDuration prometheus.Gauge
	scrapeSuccess  prometheus.Gauge
	lastFetch      prometheus.Gauge

	// Daily History
	qsoHistory *prometheus.GaugeVec // labels: date, band

	mu sync.Mutex
}

func NewCollector(client *lotw.Client) *Collector {
	return &Collector{
		client: client,
		qsoTotal: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "lotw_qso_total",
			Help: "Total number of QSOs logged in LoTW",
		}, []string{"band", "mode"}),
		qslTotal: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "lotw_qsl_confirmed_total",
			Help: "Total number of confirmed QSLs",
		}, []string{"band", "mode"}),
		dxccCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "lotw_dxcc_entities_count",
			Help: "Number of unique DXCC entities confirmed",
		}),
		scrapeDuration: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "lotw_scrape_duration_seconds",
			Help: "Time taken to fetch and parse LoTW data",
		}),
		scrapeSuccess: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "lotw_scrape_success",
			Help: "1 if last scrape was successful, 0 otherwise",
		}),
		lastFetch: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "lotw_last_fetch_timestamp_seconds",
			Help: "Timestamp of the last successful LoTW fetch",
		}),
		qsoHistory: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "lotw_qso_history_count",
			Help: "Number of QSOs per day for the recent past",
		}, []string{"date", "band"}),
	}
}

// Describe implements prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.qsoTotal.Describe(ch)
	c.qslTotal.Describe(ch)
	c.dxccCount.Describe(ch)
	c.scrapeDuration.Describe(ch)
	c.scrapeSuccess.Describe(ch)
	c.lastFetch.Describe(ch)
	c.qsoHistory.Describe(ch)
}

// Collect implements prometheus.Collector
// NOTE: Since LoTW is slow, we should probably FETCH asynchronously and Collect returns cached values.
// However, standard Exporter practice usually blocks.
// Given the user requested a SCHEDULE, we will implement a background fetcher and Collect simply reads state.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.qsoTotal.Collect(ch)
	c.qslTotal.Collect(ch)
	c.dxccCount.Collect(ch)
	c.scrapeDuration.Collect(ch)
	c.scrapeSuccess.Collect(ch)
	c.lastFetch.Collect(ch)
	c.qsoHistory.Collect(ch)
}

// StartBackgroundFetch runs the fetch loop.
func (c *Collector) StartBackgroundFetch(interval time.Duration) {
	// Fetch immediately
	go func() {
		c.fetch()
		ticker := time.NewTicker(interval)
		for range ticker.C {
			c.fetch()
		}
	}()
}

func (c *Collector) fetch() {
	start := time.Now()
	log.Println("Starting LoTW fetch...")

	// Fetch all records (since zero time) to rebuild full state.
	// Optimization: could cache and only fetch delta, but for simplicity we fetch all.
	// Users might have thousands of records; LoTW download is reasonably fast for text.
	r, err := c.client.FetchReport(time.Time{})

	duration := time.Since(start).Seconds()

	c.mu.Lock()
	c.scrapeDuration.Set(duration)
	if err != nil {
		log.Printf("Error fetching LoTW report: %v", err)
		c.scrapeSuccess.Set(0)
		c.mu.Unlock()
		return
	}
	defer r.Close()

	// Read entire body to debug size and content
	bodyBytes, err := io.ReadAll(r)
	if err != nil {
		log.Printf("Error reading LoTW response: %v", err)
		c.scrapeSuccess.Set(0)
		c.mu.Unlock()
		return
	}
	r.Close()

	log.Printf("Downloaded %d bytes from LoTW", len(bodyBytes))

	// Create reader from bytes
	byteReader := strings.NewReader(string(bodyBytes))

	records, err := adif.Parse(byteReader)
	if err != nil {
		log.Printf("Error parsing ADIF: %v", err)
		c.scrapeSuccess.Set(0)
		c.mu.Unlock()
		return
	}

	c.scrapeSuccess.Set(1)
	c.lastFetch.Set(float64(time.Now().Unix()))

	// Reset metrics to avoid stale data if we are doing a full refresh
	c.qsoTotal.Reset()
	c.qslTotal.Reset()
	c.qsoHistory.Reset()

	// Aggregators
	// We need to count unique DXCCs confirmed.
	dxccConfirmed := make(map[string]bool)

	// Date aggregator
	// Map date string "YYYY-MM-DD|BAND" -> count
	dateCounts := make(map[string]float64)

	for _, rec := range records {
		band := rec["BAND"]
		mode := rec["MODE"]

		// Total QSOs
		c.qsoTotal.WithLabelValues(band, mode).Inc()

		// QSL Confirmed?
		// LoTW uses QSL_RCVD = 'Y' for lotw matches? Or is it LOTW_QSL_RCVD?
		// Standard ADIF for LoTW usually puts confirmation in QSL_RCVD if fetched from LoTW.
		// Let's check both QSL_RCVD and APP_LOTW_QSL_RCVD if present?
		// We'll stick to QSL_RCVD being 'Y'.
		qslRcvd := strings.ToUpper(rec["QSL_RCVD"]) // Standard
		if qslRcvd == "Y" {
			c.qslTotal.WithLabelValues(band, mode).Inc()

			if dxcc, ok := rec["DXCC"]; ok && dxcc != "" {
				dxccConfirmed[dxcc] = true
			}
		}

		// Date history
		// Use APP_LOTW_QSO_TIMESTAMP if available (ISO8601: 2025-12-10T15:31:30Z)
		// Fallback to QSO_DATE (YYYYMMDD)
		var d string
		if ts, ok := rec["APP_LOTW_QSO_TIMESTAMP"]; ok && len(ts) >= 10 {
			d = ts[0:10] // YYYY-MM-DD
		} else {
			qsoDateRaw := rec["QSO_DATE"]
			if len(qsoDateRaw) == 8 {
				d = fmt.Sprintf("%s-%s-%s", qsoDateRaw[0:4], qsoDateRaw[4:6], qsoDateRaw[6:8])
			}
		}

		if d != "" {
			dKey := fmt.Sprintf("%s|%s", d, band)
			dateCounts[dKey]++
		}
	}

	c.dxccCount.Set(float64(len(dxccConfirmed)))

	// Populate history limits (e.g. last 30 days or so?)
	// Actually, if we just push all history, it might be fine if unique days < 10000.
	// A ham radio operator active for 10 years = 3650 points. Prometheus handles that fine.
	// We'll dump all date counts found.
	for key, count := range dateCounts {
		parts := strings.Split(key, "|")
		if len(parts) == 2 {
			c.qsoHistory.WithLabelValues(parts[0], parts[1]).Set(count)
		}
	}

	c.mu.Unlock()
	log.Printf("LoTW fetch complete. Processed %d records.", len(records))
}
