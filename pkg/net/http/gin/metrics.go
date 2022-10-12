package gin

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/peterlearn/kratos/pkg/net/metadata"
	"github.com/peterlearn/kratos/pkg/stat/metric"
	"strconv"
	"time"
)

const (
	//clientNamespace = "http_client"
	serverNamespace = "http_server"
)

var (
	_metricServerReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "http server requests duration(ms).",
		Labels:    []string{"path", "caller", "method"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})
	_metricServerReqCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "http server requests error count.",
		Labels:    []string{"path", "caller", "method", "code"},
	})
	_metricServerBBR = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: serverNamespace,
		Subsystem: "",
		Name:      "bbr_total",
		Help:      "http server bbr total.",
		Labels:    []string{"url", "method"},
	})
	//_metricClientReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
	//	Namespace: clientNamespace,
	//	Subsystem: "requests",
	//	Name:      "duration_ms",
	//	Help:      "http client requests duration(ms).",
	//	Labels:    []string{"path", "method"},
	//	Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	//})
	//_metricClientReqCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
	//	Namespace: clientNamespace,
	//	Subsystem: "requests",
	//	Name:      "code_total",
	//	Help:      "http client requests code count.",
	//	Labels:    []string{"path", "method", "code"},
	//})
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

var skipPath = []string{"/metrics", "/debug", "/healthcheck"}

func ServerReqMetric() gin.HandlerFunc {
	const noUser = "no_user"

	var skip map[string]struct{}
	if length := len(skipPath); length > 0 {
		skip = make(map[string]struct{}, length)
		for _, path := range skipPath {
			skip[path] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		path := c.FullPath()
		if _, ok := skip[path]; ok {
			c.Next()
			return
		}

		now := time.Now()
		req := c.Request
		w := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w

		c.Next()

		var code int
		resp := &JSON{}
		err := json.Unmarshal(w.body.Bytes(), resp)
		if err != nil {
			code = c.Writer.Status()
			//log.Error("Metric get response error, can't catch response code")
			//return
		} else {
			code = resp.Code
		}

		dt := time.Since(now)
		caller := metadata.String(c, metadata.Caller)
		if caller == "" {
			caller = noUser
		}

		if len(path) > 0 {
			_metricServerReqCodeTotal.Inc(path[1:], caller, req.Method, strconv.Itoa(code))
			_metricServerReqDur.Observe(int64(dt/time.Millisecond), path[1:], caller, req.Method)
		}
	}
}
