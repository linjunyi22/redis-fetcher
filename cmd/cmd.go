package cmd

import (
	"errors"
	"flag"
	"fmt"
	"github/linjunyi22/redis-fetcher/handler"
	"os"
)

var (
	help     bool
	host     string
	port     int
	//password string
	//format   string // json normal
)

type Command struct{}

func (c *Command) GetParams() {}

func (c *Command) Init() {
	c.flagSetting()
	flag.Usage = c.usage
	flag.Parse()

	if help {
		flag.Usage()
		return
	}
	c.parse()
}

func (c *Command) flagSetting() {
	flag.BoolVar(&help, "H", false, "show help")
	flag.StringVar(&host, "h", "", "server host, e.g. 127.0.0.1")
	flag.IntVar(&port, "p", 6379, "server port")
	//flag.StringVar(&password, "a", "", "server password,default is a empty string")
	//flag.StringVar(&format, "f", "json", "output format,'json' and 'normal' are in the choices")
}

func (c *Command) usage() {
	fmt.Fprintf(os.Stderr, `redis fetcher Usage: redis-fetcher [-t redisType] [-h host] [-p port] [-a password] [-f format] 
Options:
`)
	flag.PrintDefaults()
}

func (c *Command) errorOutput(err error) {
	fmt.Fprintf(os.Stderr, err.Error()+"\n")
	os.Exit(2)
}

func (c *Command) parse() {
	if len(host) == 0 {
		c.errorOutput(errors.New("the value of \"-h\" is necessary"))
	}

	//if format != "json" && format != "normal" {
	//	c.errorOutput(errors.New("the value of \"-f\" is necessary, must be 'json' or 'normal'"))
	//}

	rh := &handler.RedisHandler{
		Host:     host,
		Port:     port,
		//Password: password,
		//Format:   format,
	}

	rh.FetchRedisInfo()
}
