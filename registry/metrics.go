package registry

import (
	"fmt"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"

	"github.com/weaveworks/flux"
	fluxmetrics "github.com/weaveworks/flux/metrics"
)

type Metrics struct {
	// Latency of image fetch, that is getting *all* information about
	// an image
	FetchDuration metrics.Histogram
	// Counts of particular kinds of request
	RequestDuration metrics.Histogram
}

const (
	LabelRepository  = "repository"
	LabelRequestKind = "kind"

	RequestKindTags     = "tags"
	RequestKindMetadata = "metadata"
)

func NewMetrics() Metrics {
	return Metrics{
		FetchDuration: prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
			Namespace: "flux",
			Subsystem: "registry",
			Name:      "fetch_duration_seconds",
			Help:      "Duration of image metadata fetches, in seconds.",
			Buckets:   stdprometheus.DefBuckets,
		}, []string{fluxmetrics.LabelInstanceID, LabelRepository, fluxmetrics.LabelSuccess}),
		RequestDuration: prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
			Namespace: "flux",
			Subsystem: "registry",
			Name:      "request_duration_seconds",
			Help:      "Duration of HTTP requests made in the course of fetching image metadata",
		}, []string{fluxmetrics.LabelInstanceID, LabelRepository, LabelRequestKind, fluxmetrics.LabelSuccess}),
	}
}

func (m Metrics) WithInstanceID(instanceID flux.InstanceID) Metrics {
	return Metrics{
		FetchDuration:   m.FetchDuration.With(fluxmetrics.LabelInstanceID, string(instanceID)),
		RequestDuration: m.RequestDuration.With(fluxmetrics.LabelInstanceID, string(instanceID)),
	}
}

type instrumentedMemcacheClient struct {
	c MemcacheClient
	// Success rate of cache requests
	MemcacheRequestDuration metrics.Histogram
}

func InstrumentMemcacheClient(c MemcacheClient) MemcacheClient {
	return &instrumentedMemcacheClient{
		c: c,
		MemcacheRequestDuration: prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
			Namespace: "flux",
			Subsystem: "memcache",
			Name:      "request_duration_seconds",
			Help:      "Duration of memcache requests, in seconds.",
			Buckets:   stdprometheus.DefBuckets,
		}, []string{fluxmetrics.LabelMethod, fluxmetrics.LabelSuccess}),
	}
}

func (i *instrumentedMemcacheClient) Add(item *memcache.Item) (err error) {
	defer func(begin time.Time) {
		i.MemcacheRequestDuration.With(
			fluxmetrics.LabelMethod, "Add",
			fluxmetrics.LabelSuccess, fmt.Sprint(err == nil),
		).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return i.c.Add(item)
}

func (i *instrumentedMemcacheClient) CompareAndSwap(item *memcache.Item) (err error) {
	defer func(begin time.Time) {
		i.MemcacheRequestDuration.With(
			fluxmetrics.LabelMethod, "CompareAndSwap",
			fluxmetrics.LabelSuccess, fmt.Sprint(err == nil),
		).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return i.c.CompareAndSwap(item)
}

func (i *instrumentedMemcacheClient) Decrement(key string, delta uint64) (newValue uint64, err error) {
	defer func(begin time.Time) {
		i.MemcacheRequestDuration.With(
			fluxmetrics.LabelMethod, "Decrement",
			fluxmetrics.LabelSuccess, fmt.Sprint(err == nil),
		).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return i.c.Decrement(key, delta)
}

func (i *instrumentedMemcacheClient) Delete(key string) (err error) {
	defer func(begin time.Time) {
		i.MemcacheRequestDuration.With(
			fluxmetrics.LabelMethod, "Delete",
			fluxmetrics.LabelSuccess, fmt.Sprint(err == nil),
		).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return i.c.Delete(key)
}

func (i *instrumentedMemcacheClient) DeleteAll() (err error) {
	defer func(begin time.Time) {
		i.MemcacheRequestDuration.With(
			fluxmetrics.LabelMethod, "DeleteAll",
			fluxmetrics.LabelSuccess, fmt.Sprint(err == nil),
		).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return i.c.DeleteAll()
}

func (i *instrumentedMemcacheClient) FlushAll() (err error) {
	defer func(begin time.Time) {
		i.MemcacheRequestDuration.With(
			fluxmetrics.LabelMethod, "FlushAll",
			fluxmetrics.LabelSuccess, fmt.Sprint(err == nil),
		).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return i.c.FlushAll()
}

func (i *instrumentedMemcacheClient) Get(key string) (item *memcache.Item, err error) {
	defer func(begin time.Time) {
		i.MemcacheRequestDuration.With(
			fluxmetrics.LabelMethod, "Get",
			fluxmetrics.LabelSuccess, fmt.Sprint(err == nil),
		).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return i.c.Get(key)
}

func (i *instrumentedMemcacheClient) GetMulti(keys []string) (items map[string]*memcache.Item, err error) {
	defer func(begin time.Time) {
		i.MemcacheRequestDuration.With(
			fluxmetrics.LabelMethod, "GetMulti",
			fluxmetrics.LabelSuccess, fmt.Sprint(err == nil),
		).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return i.c.GetMulti(keys)
}

func (i *instrumentedMemcacheClient) Increment(key string, delta uint64) (newValue uint64, err error) {
	defer func(begin time.Time) {
		i.MemcacheRequestDuration.With(
			fluxmetrics.LabelMethod, "Increment",
			fluxmetrics.LabelSuccess, fmt.Sprint(err == nil),
		).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return i.c.Increment(key, delta)
}

func (i *instrumentedMemcacheClient) Replace(item *memcache.Item) (err error) {
	defer func(begin time.Time) {
		i.MemcacheRequestDuration.With(
			fluxmetrics.LabelMethod, "Replace",
			fluxmetrics.LabelSuccess, fmt.Sprint(err == nil),
		).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return i.c.Replace(item)
}

func (i *instrumentedMemcacheClient) Set(item *memcache.Item) (err error) {
	defer func(begin time.Time) {
		i.MemcacheRequestDuration.With(
			fluxmetrics.LabelMethod, "Set",
			fluxmetrics.LabelSuccess, fmt.Sprint(err == nil),
		).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return i.c.Set(item)
}

func (i *instrumentedMemcacheClient) Touch(key string, seconds int32) (err error) {
	defer func(begin time.Time) {
		i.MemcacheRequestDuration.With(
			fluxmetrics.LabelMethod, "Touch",
			fluxmetrics.LabelSuccess, fmt.Sprint(err == nil),
		).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return i.c.Touch(key, seconds)
}

func (i *instrumentedMemcacheClient) Stop() {
	defer func(begin time.Time) {
		i.MemcacheRequestDuration.With(
			fluxmetrics.LabelMethod, "Stop",
			fluxmetrics.LabelSuccess, "true",
		).Observe(time.Since(begin).Seconds())
	}(time.Now())
	i.c.Stop()
}
