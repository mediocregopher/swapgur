package backend

import (
	"github.com/fzzy/radix/redis"
	"log"
)

type swapReq struct {
	category string
	giving   string
	retCh    chan string
}

var swapCh = make(chan *swapReq)

func init() {
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatal(err)
	}
	go redisSpin(conn)
}

func redisSpin(conn *redis.Client) {
	var rep *redis.Reply
	var receiving string
	var err error
	for {
		select {
		case req := <-swapCh:
			key := "catbuffer:" + req.category
			if req.giving == "" {
				rep = conn.Cmd("get", key)
			} else {
				rep = conn.Cmd("getset", key, req.giving)
			}
			if rep.Type == redis.NilReply {
				receiving = ""
			} else {
				receiving, err = rep.Str()
				if err != nil {
					log.Printf(
						"redis error: %s - category:%s giving:%s",
						err,
						req.category,
						req.giving,
					)
				}
			}
			req.retCh <- receiving
		}
	}
}

func Swap(category, giving string) string {
	req := swapReq{category, giving, make(chan string)}
	swapCh <- &req
	return <-req.retCh
}
