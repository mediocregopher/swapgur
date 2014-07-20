package config

import (
	"github.com/mediocregopher/flagconfig"
)

var fc *flagconfig.FlagConfig

var (
	RedisAddr string
	RedisConns int
	ListenAddr string
	SwapsPerDay int
)

func init() {
	fc = flagconfig.New("swapgur")

	fc.StrParam("redis-addr", "TCP address of redis", "localhost:6379")
	fc.IntParam("redis-conns", "Number of redis connections to make", 10)
	fc.StrParam("listen-addr", "Address to listen on", ":8787")
	fc.IntParam("max-swaps-per-day", "Number of times a particular ip address can submit a particular url per-day. Leave 0 to not do this check", 0)

	fc.Parse()

	RedisAddr = fc.GetStr("redis-addr")
	RedisConns = fc.GetInt("redis-conns")
	ListenAddr = fc.GetStr("listen-addr")
	SwapsPerDay = fc.GetInt("max-swaps-per-day")

}
