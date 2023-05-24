package apiclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const timeout = 5 * time.Second

var (
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "controlplane",
			Subsystem: "apiclient",
			Name:      "requests_total",
			Help:      "Number of requests sent",
		},
		[]string{"method", "url", "code"},
	)

	requestsDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "controlplane",
			Subsystem: "apiclient",
			Name:      "requests_duration_seconds",
			Help:      "Timings for sent requests",
			Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.3, 0.5, 0.7, 0.9, 1, 1.1, 1.3, 1.5, 1.7, 1.9, 2.0, 2.5, 5, 10},
		},
		[]string{"method", "url"},
	)
)

type Sender struct {
	Timeout   time.Duration
	Transport http.RoundTripper
	Pacer     *Pacer
}

func NewSender() Sender {
	return Sender{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}

// {"errors":[{"title":"Seat already taken","detail":"Taken seats: [1]"}]}
type Error struct {
	Title  string `json:"title,omitempty"`
	Detail string `json:"detail,omitempty"`
	Source string `json:"source,omitempty"`
}

type Errors struct {
	StatusCode int     `json:"statusCode,omitempty"`
	Errors     []Error `json:"errors,omitempty"`
}

func (e Errors) Error() string {
	body, _ := json.Marshal(e)
	return string(body)
}

func (s Sender) send(ctx context.Context, method string, url string, bodyIn, bodyOut interface{}) error {
	if s.Pacer != nil {
		s.Pacer.Step()
	}

	furl := filterIDs(url)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var err error
	var rawBodyIn []byte
	if bodyIn != nil {
		if rawBodyIn, err = json.Marshal(bodyIn); err != nil {
			return fmt.Errorf("failed to encode body: %v", err)
		}
	}

	reader := bytes.NewReader(rawBodyIn)
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return fmt.Errorf("failed to build request: %v", err)
	}
	glog.V(1).Info(method, " to ", url)

	req.Header.Set("User-Agent", "controlplane-load-generator")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Slurm-RPC-Deadline", fmt.Sprintf("%d", time.Now().Add(timeout).Unix()))

	timer := prometheus.NewTimer(requestsDuration.WithLabelValues(method, furl))
	defer timer.ObserveDuration()

	client := http.Client{Transport: s.Transport}
	resp, err := client.Do(req)
	if err != nil {
		requestsTotal.WithLabelValues(method, furl, "").Inc()
		return fmt.Errorf("failed to do HTTP request: %v", err)
	}

	requestsTotal.WithLabelValues(method, furl, fmt.Sprintf("%d", resp.StatusCode)).Inc()

	defer resp.Body.Close()
	rawBodyOut, _ := ioutil.ReadAll(resp.Body)
	glog.V(1).Infof("Got response %d: %s", resp.StatusCode, rawBodyOut[:200])
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		if bodyOut != nil {
			if err := json.Unmarshal(rawBodyOut, bodyOut); err != nil {
				return fmt.Errorf("failed to decode body: %v", err)
			}
		}

		return nil
	}

	if rawBodyOut == nil {
		rawBodyOut = []byte("")
	}

	verrs := Errors{StatusCode: resp.StatusCode}
	if err := json.Unmarshal(rawBodyOut, &verrs); err != nil {
		return fmt.Errorf(
			"API returned unexpected result:\nstatus code: %d\nbody:%s",
			resp.StatusCode,
			string(rawBodyOut),
		)
	}

	return verrs
}

var numbersRegExp = regexp.MustCompile("[0-9]+")

func filterIDs(in string) string {
	return numbersRegExp.ReplaceAllString(in, ":id")
}
