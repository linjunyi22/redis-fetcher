package handler

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
)

type RedisHandler struct {
	Host string
	Port int
	//Password string
	//Format     string
	masterName string
}

func printError(err string) {
	_, _ = fmt.Fprintf(os.Stderr, err+"\n")
}

func (rh *RedisHandler) logToFile(dirName string, fileParams map[string]string) {
	dir, err := os.Getwd()
	if err != nil {
		printError(err.Error())
		return
	}
	dir = path.Join(dir, "redis-fetcher", dirName)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		printError(err.Error())
		return
	}

	for k, v := range fileParams {
		fp := path.Join(dir, k)
		err := ioutil.WriteFile(fp, []byte(v), os.ModePerm)
		if err != nil {
			printError(err.Error())
			return
		}
	}
	return
}

func (rh *RedisHandler) connection() (redis.Conn, error) {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", rh.Host, rh.Port))
	if err != nil {
		printError(fmt.Sprintf(`fail to connect to %s:%d
Error:%s `, rh.Host, rh.Port, err))
		return nil, err
	}
	return conn, nil
}

func (rh *RedisHandler) infoHandler(info []interface{}) map[string]string {
	fileParams := make(map[string]string)
	for _, v := range info {
		currentIP := ""
		tmpInfo := ""
		flag := false
		for ii, vv := range v.([]interface{}) {
			tmpInfo += string(vv.([]byte))
			if ii%2 == 0 {
				tmpInfo += ":"
			} else {
				tmpInfo += "\n"
			}

			if flag {
				currentIP = string(vv.([]byte))
				flag = false
			}

			if string(vv.([]byte)) == "ip" {
				flag = true
			}
		}
		if len(currentIP) != 0 {
			fileParams[currentIP] = tmpInfo
		}
	}
	return fileParams
}

func (rh *RedisHandler) getSentinelInfo() {
	conn, err := rh.connection()
	defer conn.Close()
	if err != nil {
		return
	}
	s, err := redis.String(conn.Do("info", "sentinel"))
	if err != nil {
		printError(fmt.Sprintf("Error : %s", err.Error()))
		return
	}

	if len(s) == 0 {
		printError(fmt.Sprintf("wrong host,can not find the sentinel info"))
		return
	}

	for _, v := range strings.Split(s, "\n") {
		if strings.Contains(v, "master0:name") {
			re, err := regexp.Compile("master0:name=(.*?),")
			if err != nil {
				printError(fmt.Sprintf("Error : %s", err.Error()))
				return
			}
			tmpMatch := re.FindStringSubmatch(v)
			if len(tmpMatch) < 2 {
				printError("can not find the master name,please check your redis configuration")
				return
			}
			rh.masterName = tmpMatch[1]
			break
		}
	}
	if len(rh.masterName) == 0 {
		printError("can not find the master name,please check your redis configuration")
		return
	}

	fileParams := make(map[string]string)
	currentHostInfo, err := redis.String(conn.Do("info"))
	if err != nil {
		printError(err.Error())
		return
	}

	fileParams[rh.Host] = currentHostInfo

	sentinelInfo, err := redis.Values(conn.Do("sentinel", "sentinels", rh.masterName))
	if err != nil {
		printError(err.Error())
		return
	}

	info := rh.infoHandler(sentinelInfo)
	for k, v := range info {
		fileParams[k] = v
	}

	rh.logToFile("sentinel", fileParams)
	fmt.Println("download sentinel info done!")
}

func (rh *RedisHandler) getServerInfo() {
	conn, err := rh.connection()
	defer conn.Close()
	if err != nil {
		return
	}

	if len(rh.masterName) == 0 {
		return
	}

	fileParams := make(map[string]string)

	masterParams, err := redis.Values(conn.Do("sentinel", "masters"))
	if err != nil {
		printError(err.Error())
		return
	}

	mInfo := rh.infoHandler(masterParams)
	for k, v := range mInfo {
		fileParams[k] = v
	}

	slaveParams, err := redis.Values(conn.Do("sentinel", "slaves", rh.masterName))
	if err != nil {
		printError(err.Error())
		return
	}

	sInfo := rh.infoHandler(slaveParams)
	for k, v := range sInfo {
		fileParams[k] = v
	}
	rh.logToFile("server", fileParams)
	fmt.Println("download server info done!")
}

func (rh *RedisHandler) FetchRedisInfo() {
	rh.getSentinelInfo()
	rh.getServerInfo()
}
