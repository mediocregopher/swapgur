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
	key := "catbuffer:" + category

	conn := <-connGetCh
	rep := conn.Cmd("getset", key, giving)
	connPutCh<- conn

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
	return receiving
}

func Get(category string) string {
	key := "catbuffer:" + category

	conn := <-connGetCh
	rep := conn.Cmd("get", key)
	connPutCh<- conn

	var receiving string
	var err error
	if rep.Type == redis.NilReply {
		receiving = ""
	} else {
		receiving, err = rep.Str()
		if err != nil {
			log.Printf("redis error: %s - category:%s", err, category)
		}
	}
	return receiving
}

func IPCanSwap(ip, url string) bool {
	if config.SwapsPerDay == 0 {
		return true
	}

	key := "ipurlcount:" + ip + ":" + url

	conn := <-connGetCh
	rep := conn.Cmd("incr", key)
	conn.Cmd("expire", key, 86400)
	connPutCh<- conn

	count, err := rep.Int()
	if err != nil {
		log.Printf("redis error: %s - ip:%s url:%s", err, ip, url)
		return true
	}

	if count > config.SwapsPerDay {
		log.Printf(
			"ip %s has tried to swap %s %d times today, rejecting",
			ip,
			url,
			count,
		)
		return false
	}

	return true
}
