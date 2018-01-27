package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Fatal("Couldn't connect to Redis. Please make sure that Redis is running.")
	}

	r := mux.NewRouter()
	r.HandleFunc("/", apiHealthHandler)
	r.HandleFunc("/hide", redisHandler(client, hideHandler)).Methods("POST")

	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":2048", nil))
}

func apiHealthHandler(w http.ResponseWriter, r *http.Request) {
	healthResponse := &HealthResponse{
		Status:    "OK!",
		Timestamp: time.Now().Unix(),
	}

	jsonResp, err := json.Marshal(healthResponse)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprintf(w, string(jsonResp))
}

func redisHandler(c *redis.Client,
	f func(c *redis.Client, w http.ResponseWriter, r *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) { f(c, w, r) }
}

func hideHandler(client *redis.Client, w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var req Request
	err = json.Unmarshal(b, &req)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if !req.Cache {
		fmt.Println("Fresh request")
	}

	fmt.Println("Checking cache")
}
