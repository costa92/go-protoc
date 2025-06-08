package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusHandler 返回用于暴露Prometheus指标的HTTP处理器
func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}
