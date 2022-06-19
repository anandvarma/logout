package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Config struct {
}

const (
	HOST              = "0.0.0.0"
	PUBLIC_IP         = "logout.my.to"
	TCP_PORT          = "8888"
	WEB_PORT          = "8080"
	READ_CHUNK_SIZE   = 128 // Max line width when reading logs
	BUFF_ARR_CAP      = 128 // Max number of lines
	STREAM_QUEUE_SIZE = 64  // Websocket backlog write buffer
	LOG_DB_PATH       = "/tmp/logout"
	MAX_TOKEN_NUM     = 16 * 1000 * 1000
	PRINT_QR_CODE     = true
)

var (
	numHttpHome = promauto.NewCounter(prometheus.CounterOpts{
		Name: "num_http_home",
		Help: "The total number of HTTP requests made to /",
	})

	numHttpView = promauto.NewCounter(prometheus.CounterOpts{
		Name: "num_http_view",
		Help: "The total number of HTTP requests made to /view",
	})

	numHttpStream = promauto.NewCounter(prometheus.CounterOpts{
		Name: "num_http_stream",
		Help: "The total number of HTTP requests made to /stream",
	})

	numHttp404 = promauto.NewCounter(prometheus.CounterOpts{
		Name: "num_http_404",
		Help: "The total number of HTTP requests that resulted in a 404 error code",
	})

	numTcpAccepts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "num_tcp_accepts",
		Help: "The total number of TCP accepts",
	})
)
