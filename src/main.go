package main

import (
	"net/http"
	"github.com/golang/glog"
	"bufio"
	"strings"
	"crypto/md5"
	"encoding/hex"
	"github.com/go-redis/redis"
)

type Config struct {
	s3Url         string
	backendUrl    string
	redisHost     string
	redisPort     string
	redisPassword string
	redisDB       int
}

func getMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func handle(w http.ResponseWriter, r *http.Request) {
	config := Config{
		"https://www.sample-videos.com/img/Sample-png-image-500kb.png",
		"http://alfa.kadromierz.pl",
		"localhost",
		"6379",
		"",
		0,
	}

	client := redis.NewClient(&redis.Options{
		Addr:     config.redisHost + ":" + config.redisPort,
		Password: config.redisPassword,
		DB:       config.redisDB,
	})

	fileHash := getMD5Hash(r.RequestURI)

	_, err := client.Get(fileHash).Result()

	if err == redis.Nil {
		_, err := http.Get(config.backendUrl + r.RequestURI)

		if err != nil {
			glog.Error("Error on backend")
			return
		}

		client.Set(fileHash, nil, 0)
	}

	resp, err := http.Get(config.s3Url + r.RequestURI)

	if err != nil {
		glog.Error(err)
		return
	}

	for k, v := range resp.Header {
		w.Header().Set(k, strings.Join(v, ", "))
	}

	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadBytes('\n')

		if err != nil {
			return
		}

		w.Write(line)
	}
}

func main() {
	http.HandleFunc("/", handle)
	http.ListenAndServe(":8080", nil)
}
