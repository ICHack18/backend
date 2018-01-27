package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
)

const (
	key      = "3c9bda420b1f4c7d81ee65210b55fe11"
	endpoint = "https://westcentralus.api.cognitive.microsoft.com/vision/v1.0/analyze?visualFeatures=categories,description&language=en"
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
	r.HandleFunc("/call-ms-cv/", cvHandler)

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

func cvHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	response := getDescriptionFromCognitiveServices(url)
	responseJson, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJson)
}

func getDescriptionFromCognitiveServices(url string) *CVResponse {
	var msReq = CVRequest{url}
	postData, err := json.Marshal(msReq)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(postData))
	req.Header.Set("Ocp-Apim-Subscription-Key", key)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var response CVResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return &response
}
