package backend

import (
	"github.com/fzzy/radix/redis"
	"log"

	"swapgur/config"
)

var connGetCh = make(chan *redis.Client, config.RedisConns)
var connPutCh = make(chan *redis.Client, config.RedisConns)

func init() {
	log.Printf("Creating %d redis connections", config.RedisConns)
	for i := 0; i < config.RedisConns; i++ {
		conn, err := redis.Dial("tcp", config.RedisAddr)
		if err != nil {
			log.Fatal(err)
		}
		connPutCh<- conn
	}
	log.Println("Redis conns created")
	go connSpin()
}

func connSpin() {
	var conn *redis.Client
	for {
		conn = <-connPutCh
		connGetCh<- conn
	}
}

func Swap(category, giving string) string {
	conn := <-connGetCh

	key := "catbuffer:" + category
	var rep *redis.Reply
	if giving == "" {
		rep = conn.Cmd("get", key)
	} else {
		rep = conn.Cmd("getset", key, giving)
	}

	var receiving string
	var err error
	if rep.Type == redis.NilReply {
		receiving = ""
	} else {
		receiving, err = rep.Str()
		if err != nil {
			log.Printf(
				"redis error: %s - category:%s giving:%s",
				err,
				category,
				giving,
			)
		}
	}
	connPutCh<- conn
	return receiving
}
