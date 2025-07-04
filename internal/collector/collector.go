package collector

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/ezhische/qrator-exporter/internal/collector/entity"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	namespace = "qrator"
	gCount    = 10 //Количество горутин для сбора метрик
)

type Collector struct {
	config *config
	client *http.Client

	bypassedTraffic   prometheus.GaugeVec
	incomingTraffic   prometheus.GaugeVec
	outgoingTraffic   prometheus.GaugeVec
	bypassedPackets   prometheus.GaugeVec
	incomingPackets   prometheus.GaugeVec
	outgoingPackets   prometheus.GaugeVec
	requestRate       prometheus.GaugeVec
	slowRequestsCount prometheus.GaugeVec
	errorsCount       prometheus.GaugeVec
	bannedIPs         prometheus.GaugeVec
	billableTraffic   prometheus.GaugeVec

	totalScrapes            prometheus.Counter
	failedDomainScrapes     prometheus.Counter
	failedDomainHTTPScrapes prometheus.Counter
	failedDomainBillScrapes prometheus.Counter
	failedDomainIPScrapes   prometheus.Counter
	sync.Mutex
}

type config struct {
	aPIKey       string
	clientID     int
	qratorAPIURL string
	domainsList  []int
	proxyURL     string
	timeout      time.Duration
	logger       *logrus.Logger
	con          int
}

type Semaphore struct {
	C chan struct{}
}

func (s *Semaphore) Acquire() {
	s.C <- struct{}{}
}

func (s *Semaphore) Release() {
	<-s.C
}

func CollectorFromConfig(
	apiKey string,
	clientID int,
	apiURL string,
	domainsList []int,
	proxy string,
	timeout time.Duration,
	logger *logrus.Logger,
	con int,
) (*Collector, error) {
	conf := &config{
		aPIKey:       apiKey,
		clientID:     clientID,
		qratorAPIURL: apiURL,
		domainsList:  domainsList,
		timeout:      timeout,
		proxyURL:     proxy,
		logger:       logger,
		con:          con,
	}
	return NewCollector(conf)
}

func NewCollector(conf *config) (*Collector, error) {
	client, err := newClient(conf.proxyURL, conf.timeout)
	if err != nil {
		return &Collector{}, fmt.Errorf("error creating client: %w", err)
	}

	collector := &Collector{
		config: conf,
		client: client,
	}
	err = collector.qratorCheck()
	if err != nil {
		return nil, err
	}

	collector.totalScrapes = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "exporter_scrapes_total",
		Help:      "Count of total scrapes",
	})

	collector.failedDomainScrapes = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "exporter_failed_domain_scrapes_total",
		Help:      "Count of failed domains scrapes",
	})

	collector.failedDomainHTTPScrapes = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "exporter_failed_domain_http_stats_scrapes_total",
		Help:      "Count of failed stats scrapes",
	})

	collector.failedDomainIPScrapes = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "exporter_failed_domain_ip_stats_scrapes_total",
		Help:      "Count of failed stats scrapes",
	})

	collector.failedDomainBillScrapes = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "exporter_failed_domain_billable_stats_scrapes_total",
		Help:      "Count of failed stats scrapes",
	})

	collector.bypassedTraffic = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "bypassed_traffic",
			Help:      "Bypassed traffic (bps)",
		},
		[]string{
			"domain",
		},
	)

	collector.incomingTraffic = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "incoming_traffic",
			Help:      "Incoming traffic (bps)",
		},
		[]string{
			"domain",
		},
	)

	collector.outgoingTraffic = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "outgoing_traffic",
			Help:      "Outgoing traffic (bps)",
		},
		[]string{
			"domain",
		},
	)

	collector.bypassedPackets = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "bypassed_packets",
			Help:      "Bypassed packets (pps)",
		},
		[]string{
			"domain",
		},
	)

	collector.incomingPackets = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "incoming_packets",
			Help:      "Incoming packets (pps)",
		},
		[]string{
			"domain",
		},
	)
	collector.outgoingPackets = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "output_packets",
			Help:      "Output packets (pps)",
		},
		[]string{
			"domain",
		},
	)

	collector.requestRate = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "request_rate",
			Help:      "Request rate (rps)",
		},
		[]string{
			"domain",
		},
	)

	collector.slowRequestsCount = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "slow_requests_count",
			Help:      "Slow request count by treshold",
		},
		[]string{
			"domain",
			"treshold_seconds",
		},
	)

	collector.errorsCount = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "errors_count",
			Help:      "Errors count by code",
		},
		[]string{
			"domain",
			"code",
		},
	)

	collector.bannedIPs = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "banned_ip_addresses_count",
			Help:      "Number of IPs banned by Qrator",
		},
		[]string{
			"domain",
			"source",
		},
	)

	collector.billableTraffic = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "billable_traffic",
			Help:      "Billable traffic (Mbps)",
		},
		[]string{
			"domain",
		},
	)
	return collector, nil
}

func newClient(proxy string, timeout time.Duration) (*http.Client, error) {
	if proxy == "" {
		return &http.Client{
			Timeout: timeout,
		}, nil
	}
	proxyUrl, err := url.Parse(proxy)
	if err != nil {
		return nil, fmt.Errorf("error parsing proxy url: %w", err)
	}
	return &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
	}, nil
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.Lock()
	defer c.Unlock()

	c.totalScrapes.Inc()
	qds, err := c.getQratorDomains()
	if err != nil {
		c.failedDomainScrapes.Inc()
		c.config.logger.Errorf("error getting domains:%s", err)
	}

	sem := Semaphore{
		C: make(chan struct{}, c.config.con),
	}
	wg := &sync.WaitGroup{}
	for _, qd := range qds {
		//IPStat API
		wg.Add(1)
		go func(qd entity.QratorDomain, ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
			sem.Acquire()
			defer sem.Release()
			defer wg.Done()

			iPStat, err := c.getQratorDomainIPStats(qd)
			if err != nil {
				c.failedDomainIPScrapes.Inc()
				c.config.logger.Errorf("failed to get ip stats: %s", err)
				return
			}

			c.bypassedTraffic.WithLabelValues(qd.Name).Set(float64(iPStat.Result.Bandwidth.Passed))
			c.incomingTraffic.WithLabelValues(qd.Name).Set(float64(iPStat.Result.Bandwidth.Input))
			c.outgoingTraffic.WithLabelValues(qd.Name).Set(float64(iPStat.Result.Bandwidth.Output))
			c.bypassedPackets.WithLabelValues(qd.Name).Set(float64(iPStat.Result.Packets.Passed))
			c.incomingPackets.WithLabelValues(qd.Name).Set(float64(iPStat.Result.Packets.Input))
			c.outgoingPackets.WithLabelValues(qd.Name).Set(float64(iPStat.Result.Packets.Output))
			c.bannedIPs.WithLabelValues(qd.Name, "Qrator").Set(float64(iPStat.Result.Blacklist.Qrator))
			c.bannedIPs.WithLabelValues(qd.Name, "Qrator.API").Set(float64(iPStat.Result.Blacklist.API))
			c.bannedIPs.WithLabelValues(qd.Name, "WAF").Set(float64(iPStat.Result.Blacklist.WAF))
			c.bannedIPs.WithLabelValues(qd.Name, "Custom").Set(float64(iPStat.Result.Blacklist.Custom))

			ch <- c.bypassedTraffic.WithLabelValues(qd.Name)
			ch <- c.incomingTraffic.WithLabelValues(qd.Name)
			ch <- c.outgoingTraffic.WithLabelValues(qd.Name)
			ch <- c.bypassedPackets.WithLabelValues(qd.Name)
			ch <- c.incomingPackets.WithLabelValues(qd.Name)
			ch <- c.outgoingPackets.WithLabelValues(qd.Name)
			ch <- c.bannedIPs.WithLabelValues(qd.Name, "Qrator")
			ch <- c.bannedIPs.WithLabelValues(qd.Name, "Qrator.API")
			ch <- c.bannedIPs.WithLabelValues(qd.Name, "WAF")
			ch <- c.bannedIPs.WithLabelValues(qd.Name, "Custom")
		}(qd, ch, wg)

		//HTTP Stat API
		wg.Add(1)
		go func(qd entity.QratorDomain, ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
			sem.Acquire()
			defer sem.Release()
			defer wg.Done()

			httpStat, err := c.getQratorDomainHTTPStats(qd)
			if err != nil {
				c.failedDomainHTTPScrapes.Inc()
				c.config.logger.Errorf("failed to get http stats: %s", err)
				return
			}

			c.requestRate.WithLabelValues(qd.Name).Set(float64(httpStat.Result.Requests))
			c.slowRequestsCount.WithLabelValues(qd.Name, "0.2").Set(float64(httpStat.Result.Responses.Duration0000_0200))
			c.slowRequestsCount.WithLabelValues(qd.Name, "0.5").Set(float64(httpStat.Result.Responses.Duration0200_0500))
			c.slowRequestsCount.WithLabelValues(qd.Name, "0.7").Set(float64(httpStat.Result.Responses.Duration0500_0700))
			c.slowRequestsCount.WithLabelValues(qd.Name, "1.0").Set(float64(httpStat.Result.Responses.Duration0700_1000))
			c.slowRequestsCount.WithLabelValues(qd.Name, "1.5").Set(float64(httpStat.Result.Responses.Duration1000_1500))
			c.slowRequestsCount.WithLabelValues(qd.Name, "2.0").Set(float64(httpStat.Result.Responses.Duration1500_2000))
			c.slowRequestsCount.WithLabelValues(qd.Name, "5.0").Set(float64(httpStat.Result.Responses.Duration2000_5000))
			c.slowRequestsCount.WithLabelValues(qd.Name, ">5").Set(float64(httpStat.Result.Responses.Duration5000_Inf))
			c.errorsCount.WithLabelValues(qd.Name, "Total").Set(float64(httpStat.Result.Errors.Total))
			c.errorsCount.WithLabelValues(qd.Name, "500").Set(float64(httpStat.Result.Errors.Code500))
			c.errorsCount.WithLabelValues(qd.Name, "501").Set(float64(httpStat.Result.Errors.Code501))
			c.errorsCount.WithLabelValues(qd.Name, "502").Set(float64(httpStat.Result.Errors.Code502))
			c.errorsCount.WithLabelValues(qd.Name, "503").Set(float64(httpStat.Result.Errors.Code503))
			c.errorsCount.WithLabelValues(qd.Name, "504").Set(float64(httpStat.Result.Errors.Code504))
			c.errorsCount.WithLabelValues(qd.Name, "4XX").Set(float64(httpStat.Result.Errors.Code4xx))

			ch <- c.requestRate.WithLabelValues(qd.Name)
			ch <- c.slowRequestsCount.WithLabelValues(qd.Name, "0.2")
			ch <- c.slowRequestsCount.WithLabelValues(qd.Name, "0.5")
			ch <- c.slowRequestsCount.WithLabelValues(qd.Name, "0.7")
			ch <- c.slowRequestsCount.WithLabelValues(qd.Name, "1.0")
			ch <- c.slowRequestsCount.WithLabelValues(qd.Name, "1.5")
			ch <- c.slowRequestsCount.WithLabelValues(qd.Name, "2.0")
			ch <- c.slowRequestsCount.WithLabelValues(qd.Name, "5.0")
			ch <- c.slowRequestsCount.WithLabelValues(qd.Name, ">5")
			ch <- c.errorsCount.WithLabelValues(qd.Name, "Total")
			ch <- c.errorsCount.WithLabelValues(qd.Name, "500")
			ch <- c.errorsCount.WithLabelValues(qd.Name, "501")
			ch <- c.errorsCount.WithLabelValues(qd.Name, "502")
			ch <- c.errorsCount.WithLabelValues(qd.Name, "503")
			ch <- c.errorsCount.WithLabelValues(qd.Name, "504")
			ch <- c.errorsCount.WithLabelValues(qd.Name, "4XX")
		}(qd, ch, wg)

		// Billable API
		wg.Add(1)
		go func(qd entity.QratorDomain, ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
			sem.Acquire()
			defer sem.Release()
			defer wg.Done()

			billStat, err := c.getQratorDomainBillableStats(qd)
			if err != nil {
				c.failedDomainBillScrapes.Inc()
				c.config.logger.Errorf("failed to get billable stats: %s", err)
				return
			}

			c.billableTraffic.WithLabelValues(qd.Name).Set(float64(billStat.Result))
			ch <- c.billableTraffic.WithLabelValues(qd.Name)
		}(qd, ch, wg)

	}

	wg.Wait()
	ch <- c.totalScrapes
	ch <- c.failedDomainScrapes
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}
