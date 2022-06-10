package main

const (
	HOST              = "localhost"
	PORT              = "8888"
	READ_CHUNK_SIZE   = 128 // Max line width when reading logs
	BUFF_ARR_CAP      = 128 // Max number of lines
	STREAM_QUEUE_SIZE = 64  // Websocket backlog write buffer
	LOG_DB_PATH       = "/home/anand/tmp/"
	MAX_TOKEN_NUM     = 16 * 1000 * 1000
)
