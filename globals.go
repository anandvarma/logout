package main

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Config struct {
}

var ErrLogNotFound = errors.New("log identified by key does not exist")
var ErrKeyExists = errors.New("log identified by key already exists")
var ErrInvalidKey = errors.New("key is an invalid log identifier")

const (
	HOST               = "0.0.0.0"
	PUBLIC_IP          = "logout.a8a.dev"
	TCP_PORT           = "8888"
	WEB_PORT           = "8080"
	READ_CHUNK_SIZE    = 128 // Max line width when reading logs
	BUFF_ARR_CAP       = 128 // Max number of lines
	STREAM_QUEUE_SIZE  = 64  // Websocket backlog write buffer
	FS_LOG_DB_PATH     = "/tmp/logout"
	BOLT_CONF_DB_PATH  = "/tmp/logout/bolt.db"
	MAX_TOKEN_NUM      = 16 * 1000 * 1000
	PRINT_QR_CODE      = true
	ANONYMOUS_OWNER_ID = 0
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
