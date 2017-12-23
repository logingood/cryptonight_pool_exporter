package main

import (
	"encoding/json"
	"io/ioutil"
	"flag"
	"compress/flate"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type PoolStats struct {
	TotalPayments	 int `json:totalPayments`
	Hashrate			 float64 `json:hashrate`
	RoundHashes		 int `json:roundHashes`
	LastBlockFound string `json:lastBlockFound`
}

type NetworkStats struct {
	Difficulty		int `json:difficulty`
}


type CpoolStats struct {
	Pool *PoolStats
	Network *NetworkStats
}

type expConf struct {
	Dial_Addr []string
	Port      string
	Proto     string
	Method    string
}

func fillDefaults() *expConf {
	confDefault := &expConf{
		Dial_Addr: []string{"127.0.0.1"},
		Port:      "8117",
		Proto:     "tcp",
		Method:    "stats",
	}
	return confDefault
}

func readConf() *expConf {
	conf := fillDefaults()

	dial_addr := os.Getenv("CPOOL_DIAL_ADDR")
	if len(dial_addr) == 0 {
		panic("DIAL_ADDR env must be set, e.g.: export CPOOL_DIAL_ADDR=192.168.1.1;192.168.1.2;..")
	}

	dial_addr_slice := strings.Split(dial_addr, ";")
	conf.Dial_Addr = dial_addr_slice

	port := os.Getenv("CPOOL_PORT")
	if len(port) != 0 {
		conf.Port = port
	}

	proto := os.Getenv("CPOOL_PROTO")
	if len(proto) != 0 {
		conf.Proto = proto
	}

	method := os.Getenv("CPOOL_STATS")
	if len(method) != 0 {
		conf.Method = method
	}

	return conf
}


func callCpool(addr string, conf *expConf) (response *CpoolStats) {

	client  := http.Client{}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%s/stats", addr, conf.Port), nil)
	req.Header.Add("Accept-Encoding", "deflate, gzip")

  if err != nil {
    log.Fatal(err)
  }

	// Synchronous call

	res, getErr := client.Do(req)
	if getErr != nil {
			log.Fatal(getErr)
	}

	res.Header.Del("Content-Length")
	re := flate.NewReader(res.Body)
	defer re.Close()

	if err != nil {
			panic(err)
	}

	if err != nil {
		log.Fatal("Can't parse response:", err)
	}

	enflate, err := ioutil.ReadAll(re)

	err = json.Unmarshal(enflate, &response)
	if err != nil {
		panic(err)
  }
		
	return response
}

type CpoolStatsCollector struct{}

func NewCpoolStatsCollector() *CpoolStatsCollector {
	return &CpoolStatsCollector{}
}

var (
	totalPaymentsDesc = prometheus.NewDesc(
		"TotalPayments",
		"Total Payments made by pool",
		[]string{"Pool"},
		nil)

	HashrateDesc = prometheus.NewDesc(
		"hashrate",
		"Total hashrate",
		[]string{"Pool"},
		nil)

	RoundHashesDesc = prometheus.NewDesc(
		"roundhashes",
		"Amount of hashes submitted",
		[]string{"Pool"},
		nil)

	LastBlockFoundDesc = prometheus.NewDesc(
		"lastblockfound",
		"Timestamp when last block found",
		[]string{"Pool"},
		nil)

	DifficultyDesc = prometheus.NewDesc(
		"difficulty",
		"Network Difficulty",
		[]string{"Pool"},
		nil)
)

func (c *CpoolStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- totalPaymentsDesc
	ch <- HashrateDesc
	ch <- RoundHashesDesc
	ch <- LastBlockFoundDesc
	ch <- DifficultyDesc
}

func (c *CpoolStatsCollector) Collect(ch chan<- prometheus.Metric) {

	conf := readConf()
	for _, addr := range conf.Dial_Addr {

		stats := callCpool(addr, conf)

		fmt.Printf("%+v", stats)

		totalPayments  := float64(stats.Pool.TotalPayments)

		ch <- prometheus.MustNewConstMetric(totalPaymentsDesc,
			prometheus.GaugeValue,
			totalPayments,
			addr)

		Hashrate := float64(stats.Pool.Hashrate)
		ch <- prometheus.MustNewConstMetric(HashrateDesc,
			prometheus.GaugeValue,
			Hashrate,
			addr)

		RoundHashes := float64(stats.Pool.RoundHashes)
		ch <- prometheus.MustNewConstMetric(RoundHashesDesc,
			prometheus.GaugeValue,
			RoundHashes,
			addr)

		LastBlockFound, _ := strconv.ParseFloat(stats.Pool.LastBlockFound, 32)
		ch <- prometheus.MustNewConstMetric(LastBlockFoundDesc,
			prometheus.GaugeValue,
			LastBlockFound,
			addr)

		Difficulty := float64(stats.Network.Difficulty)
		ch <- prometheus.MustNewConstMetric(DifficultyDesc,
			prometheus.GaugeValue,
			Difficulty,
			addr)

	}

}

func main() {

	var (
		listenAddress = flag.String("web.listen-address", ":10335", "Address on which to expose metrics and web interface.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	)

	claymore_collector := NewCpoolStatsCollector()

	prometheus.MustRegister(claymore_collector)

	http.Handle(*metricsPath, prometheus.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Cryptonight Pool Stats Exporter</title></head>
			<body>
			<h1>Claymore Stasts Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})
	http.ListenAndServe(*listenAddress, nil)

}
