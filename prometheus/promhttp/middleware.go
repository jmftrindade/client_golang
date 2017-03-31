// Copyright 2016 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Copyright (c) 2013, The Prometheus Authors
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package promhttp

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// ClientTrace adds middleware providing a histogram of outgoing request
// latencies, partitioned by http client, request host and httptrace event.
func ClientTrace(obs prometheus.ObserverVec, next httpClient) httpClient {
	// The supplied histogram NEEDS a label for the httptrace event.
	// TODO: Using `event` for now, but any other name is acceptable.

	checkEventLabel(obs)
	return ClientMiddleware(func(r *http.Request) (*http.Response, error) {
		var (
			start = time.Now()
		)

		trace := &httptrace.ClientTrace{
			DNSStart: func(_ httptrace.DNSStartInfo) {
				obs.WithLabelValues("DNSStart").Observe(time.Since(start).Seconds())
			},
			DNSDone: func(_ httptrace.DNSDoneInfo) {
				obs.WithLabelValues("DNSDone").Observe(time.Since(start).Seconds())
			},
			ConnectStart: func(_, _ string) {
				obs.WithLabelValues("ConnectStart").Observe(time.Since(start).Seconds())
			},
			ConnectDone: func(net, addr string, err error) {
				if err != nil {
					return
				}
				obs.WithLabelValues("ConnectDone").Observe(time.Since(start).Seconds())
			},
			GotConn: func(_ httptrace.GotConnInfo) {
				obs.WithLabelValues("GotConn").Observe(time.Since(start).Seconds())
			},
			GotFirstResponseByte: func() {
				obs.WithLabelValues("GotFirstResponseByte").Observe(time.Since(start).Seconds())
			},
			TLSHandshakeStart: func() {
				obs.WithLabelValues("TLSHandshakeStart").Observe(time.Since(start).Seconds())
			},
			TLSHandshakeDone: func(_ tls.ConnectionState, err error) {
				if err != nil {
					return
				}
				obs.WithLabelValues("TLSHandshakeDone").Observe(time.Since(start).Seconds())
			},
			WroteRequest: func(_ httptrace.WroteRequestInfo) {
				obs.WithLabelValues("WroteRequest").Observe(time.Since(start).Seconds())
			},
		}
		r = r.WithContext(httptrace.WithClientTrace(context.Background(), trace))

		return next.Do(r)
	})
}

// InFlight is middleware that instruments number of open requests partitioned
// by http client and request host.
func InFlightC(gauge prometheus.Gauge, next httpClient) httpClient {
	return ClientMiddleware(func(r *http.Request) (*http.Response, error) {
		gauge.Inc()
		resp, err := next.Do(r)
		if err != nil {
			return nil, err
		}
		gauge.Dec()
		return resp, err
	})
}

func CounterC(counter *prometheus.CounterVec, next httpClient) httpClient {
	code, method := checkLabels(counter)

	return ClientMiddleware(func(r *http.Request) (*http.Response, error) {
		resp, err := next.Do(r)
		if err != nil {
			return nil, err
		}
		counter.With(labels(code, method, r.Method, resp.StatusCode)).Inc()
		return resp, err
	})
}

func checkEventLabel(c prometheus.Collector) {
	var (
		desc *prometheus.Desc
		pm   dto.Metric
	)

	descc := make(chan *prometheus.Desc, 1)
	c.Describe(descc)

	select {
	case desc = <-descc:
	default:
		panic("no description provided by collector")
	}

	m, err := prometheus.NewConstMetric(desc, prometheus.UntypedValue, 0, "")
	if err != nil {
		panic("error checking metric for labels")
	}

	if err := m.Write(&pm); err != nil {
		panic("error checking metric for labels")
	}

	name := *pm.Label[0].Name
	if name != "event" {
		panic("metric partitioned with non-supported label")
	}
}
