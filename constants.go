package main

const (
	HOST              = "localhost"
	TCP_PORT          = "8888"
	WEB_PORT          = "8080"
	READ_CHUNK_SIZE   = 128 // Max line width when reading logs
	BUFF_ARR_CAP      = 128 // Max number of lines
	STREAM_QUEUE_SIZE = 64  // Websocket backlog write buffer
	LOG_DB_PATH       = "/tmp/logout"
	MAX_TOKEN_NUM     = 16 * 1000 * 1000
	PRINT_QR_CODE     = true
)
