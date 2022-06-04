package main

import "log"

const (
	HOST            = "localhost"
	PORT            = "8888"
	READ_CHUNK_SIZE = 128
	BUFF_ARR_CAP    = 128
	CACHE_SIZE      = 2
)

func DCHECK(cond bool, out string) {
	if !cond {
		log.Fatal(out)
	}
}
