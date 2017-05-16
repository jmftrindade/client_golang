// Copyright 2015 The Prometheus Authors
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

// A simple example exposing fictional RPC latencies with different types of
// random distributions (uniform, normal, and exponential) as Prometheus
// metrics.
package main

import (
	"flag"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	addr              = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	uniformDomain     = flag.Float64("uniform.domain", 0.0002, "The domain for the uniform distribution.")
	normDomain        = flag.Float64("normal.domain", 0.0002, "The domain for the normal distribution.")
	normMean          = flag.Float64("normal.mean", 0.00001, "The mean for the normal distribution.")
	oscillationPeriod = flag.Duration("oscillation-period", 10*time.Minute, "The duration of the rate oscillation period.")
	sizesNormDomain   = flag.Float64("normal.sdomain", 10240, "The domain for the normal distribution.")
	sizesNormMean     = flag.Float64("normal.smean", 1024, "The mean for the normal distribution.")
)

var (
	/*
		// Create a summary to track fictional interservice RPC latencies for three
		// distinct services with different latency distributions. These services are
		// differentiated via a "service" label.
		rpcDurations = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       "rpc_durations_seconds",
				Help:       "RPC latency distributions.",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{"service"},
		)
		// The same as above, but now as a histogram, and only for the normal
		// distribution. The buckets are targeted to the parameters of the
		// normal distribution, with 20 buckets centered on the mean, each
		// half-sigma wide.
		rpcDurationsHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "rpc_durations_histogram_seconds",
			Help:    "RPC latency distributions.",
			Buckets: prometheus.LinearBuckets(*normMean-5**normDomain, .5**normDomain, 20),
		})
	*/

	lineageOpDurations = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "lineage_op_durations_seconds",
			Help: "Data lineage op durations distribution (in seconds).",
		},
		[]string{"user", "op", "dataset"},
	)

	lineageOpSizes = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "lineage_op_sizes_bytes",
			Help: "Data lineage op sizes distribution (in bytes).",
		},
		[]string{"user", "op", "dataset"},
	)

/*
	// Data lineage events sizes in data bytes.
	lineageOpSizesHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "lineage_op_sizes_histogram_bytes",
			Help:    "Data lineage op sizes distributions (in bytes).",
			Buckets: prometheus.LinearBuckets(*sizesNormMean-5**sizesNormDomain, .5**sizesNormDomain, 20),
		},
		[]string{"user", "op", "dataset"},
	)

	// Data lineage event durations in seconds.
	lineageOpDurationsHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "lineage_op_durations_histogram_seconds",
			Help:    "Data lineage op duration distributions (in seconds).",
			Buckets: prometheus.LinearBuckets(*normMean-5**normDomain, .5**normDomain, 20),
		},
		[]string{"user", "op", "dataset"},
	)
*/
)

func init() {
	// Register the data lineage metrics with Prometheus's default registry.
	prometheus.MustRegister(lineageOpSizes)
	prometheus.MustRegister(lineageOpDurations)
}

func main() {
	flag.Parse()

	start := time.Now()

	oscillationFactor := func() float64 {
		return 2 + math.Sin(math.Sin(2*math.Pi*float64(time.Since(start))/float64(*oscillationPeriod)))
	}
	/*
		// Periodically record some sample latencies for the three services.
		go func() {
			for {
				v := rand.Float64() * *uniformDomain
				rpcDurations.WithLabelValues("uniform").Observe(v)
				time.Sleep(time.Duration(100*oscillationFactor()) * time.Millisecond)
			}
		}()

		go func() {
			for {
				v := (rand.NormFloat64() * *normDomain) + *normMean
				rpcDurations.WithLabelValues("normal").Observe(v)
				rpcDurationsHistogram.Observe(v)
				time.Sleep(time.Duration(75*oscillationFactor()) * time.Millisecond)
			}
		}()

		go func() {
			for {
				v := rand.ExpFloat64() / 1e6
				rpcDurations.WithLabelValues("exponential").Observe(v)
				time.Sleep(time.Duration(50*oscillationFactor()) * time.Millisecond)
			}
		}()
	*/
	// Periodically record randomly generated sample data lineage events.
	go func() {
		for {
			users := []string{"alice", "bob"}
			datasets := []string{"file_a", "file_b", "file_c"}
			opTypes := []string{"read", "write"}
			opSizes := []float64{512.0, 1024.0, 2048.0}
			opDurations := []float64{1.0, 2.0, 3.0}
			/*
				v := (rand.NormFloat64() * *sizesNormDomain) + *sizesNormMean
				lineageOpSizesHistogram.WithLabelValues(
					users[rand.Intn(len(users))],
					opTypes[rand.Intn(len(opTypes))],
					datasets[rand.Intn(len(datasets))]).Observe(v)

				// Use a different random value, but still generate both of these in side-step.
				v = (rand.NormFloat64() * *normDomain) + *normMean
				lineageOpDurationsHistogram.WithLabelValues(users[rand.Intn(len(users))],
					opTypes[rand.Intn(len(opTypes))],
					datasets[rand.Intn(len(datasets))]).Observe(v)
			*/
			// Use a different random value, but still generate both of these in side-step.
			opIdx := rand.Intn(len(opSizes))
			userIdx := rand.Intn(len(users))
			typeIdx := rand.Intn(len(opTypes))
			datasetIdx := rand.Intn(len(datasets))

			v := opSizes[opIdx]
			lineageOpSizes.WithLabelValues(
				users[userIdx],
				opTypes[typeIdx],
				datasets[datasetIdx]).Observe(v)

			v = opDurations[opIdx]
			lineageOpDurations.WithLabelValues(
				users[userIdx],
				opTypes[typeIdx],
				datasets[datasetIdx]).Observe(v)

			// Sleep a little to add variance.
			time.Sleep(time.Duration(75*oscillationFactor()) * time.Millisecond)
		}
	}()

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
