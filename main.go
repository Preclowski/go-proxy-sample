package go_static_sample

import (
	"net/http"
	"github.com/golang/glog"
	"bufio"
	"strings"
	"crypto/md5"
	"encoding/hex"
	"github.com/go-redis/redis"
	"os"
	"strconv"
	"log"
)

type Config struct {
	s3Url         string
	backendUrl    string
	redisHost     string
	redisPort     string
	redisPassword string
	redisDB       int
}

var config Config
var client redis.Client

func getMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func handle(w http.ResponseWriter, r *http.Request) {
	fileHash := getMD5Hash(r.RequestURI)
	resp := fetchImage(fileHash, r.RequestURI, w)
	reader := bufio.NewReader(resp.Body)

	for k, v := range resp.Header {
		w.Header().Set(k, strings.Join(v, ", "))
	}

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}
		w.Write(line)
	}
}

func existInRedis(fileHash string) bool {
	_, err := client.Get(fileHash).Result()
	return err == nil
}

func fetchImage(fileHash string, uri string, w http.ResponseWriter) *http.Response {
	if !existInRedis(fileHash) {
		return resolveFromBackend(uri, fileHash, w)
	} else {
		return resolveFromS3(uri, w)
	}
}

func resolveFromS3(uri string, w http.ResponseWriter) *http.Response {
	resp, err := http.Get(config.s3Url + uri)
	if err != nil {
		glog.Error(err)
		http.Error(w, "File not found", 404)
	}
	return resp
}

func resolveFromBackend(uri string, fileHash string, w http.ResponseWriter) *http.Response {
	resp, err := http.Get(config.backendUrl + uri)
	if err != nil {
		glog.Error(err)
		http.Error(w, "File not found", 404)
	}
	client.Set(fileHash, nil, 0)
	return resp
}

func main() {
	client = *redis.NewClient(&redis.Options{
		Addr:     config.redisHost + ":" + config.redisPort,
		Password: config.redisPassword,
		DB:       config.redisDB,
	})

	redisDb, err := strconv.Atoi(os.Getenv("REDIS_DB"))

	if err != nil {
		log.Fatal("Invalid redis DB")
	}

	config = Config{
		os.Getenv("S3_URL"),
		os.Getenv("BACKEND_URL"),
		os.Getenv("REDIS_HOST"),
		os.Getenv("REDIS_PORT"),
		os.Getenv("REDIS_PASSWORD"),
		redisDb,
	}

	http.HandleFunc("/", handle)
	http.ListenAndServe(":8080", nil)
}
