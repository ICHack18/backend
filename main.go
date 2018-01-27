package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
)

const (
	key      = "3c9bda420b1f4c7d81ee65210b55fe11"
	fvkey = ""
	endpoint = "https://westcentralus.api.cognitive.microsoft.com/vision/v1.0/analyze?visualFeatures=categories,description&language=en"
	fvendpoint = "https://westcentralus.api.cognitive.microsoft.com/face/v1.0"
	persongid = "banned_users"
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

	response := &Response{
		Images: make([]ImageResponse, len(req.Urls)),
	}

	for index, url := range req.Urls {
		var cvResponse CVResponse
		var fetchNew = true

		if req.UseCache {
			val, err := client.Get(url).Result()
			if err == nil {
				fetchNew = false;
				marshalErr := json.Unmarshal([]byte(val), &cvResponse)
				if marshalErr != nil {
					http.Error(w, marshalErr.Error(), 500)
					return
				}
			}
		}

		if fetchNew {
			cvResponse = getDescriptionFromCognitiveServices(url)
			responseJson, err := json.Marshal(cvResponse)
			if err == nil {
				client.Set(url, responseJson, 0).Err()
			}
		}

		imageResponse := ImageResponse{
			Url:             url,
			Hide:            shouldBlockImage(req.Tags, cvResponse),
			SubstituteImage: url,
		}

		response.Images[index] = imageResponse
	}

	responseJson, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	w.Header().Set("content-type", "application/json")
	w.Write(responseJson)
}

func cvHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	response := getDescriptionFromCognitiveServices(url)
	responseJson, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJson)
}

func getDescriptionFromCognitiveServices(url string) CVResponse {
	var msReq = &CVRequest{
		url,
	}

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

	return response
}

func shouldBlockImage(blockTags []string, cvResponse CVResponse) bool {
	imageTags := cvResponse.Description.Tags
	set := make(map[string]bool)
	for _, imageTag := range imageTags {
		set[strings.ToLower(imageTag)] = true
	}

	for _, blockTag := range blockTags {
		_, tagInImage := set[strings.ToLower(blockTag)]
		if tagInImage {
			return true
		}
	}
	return false
}

func getFaceVerification(url string) *FVResponse {
	var msReq = FVRequest{}

	postData, err := json.Marshal(msReq)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", fvendpoint, bytes.NewReader(postData))
	req.Header.Set("Ocp-Apim-Subscription-Key", fvkey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var response FVResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return &response
}

type NPGRequest struct {
	Name string
	userData string
}

func createPersonGroup(name string, info string) bool {
	var endpoint = fvendpoint + "/persongroups/" + persongid

	//var msReq = NPGRequest{username, data}
	var msReq = NPGRequest{name, info}

	postData, err := json.Marshal(msReq)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(postData))
	req.Header.Set("Ocp-Apim-Subscription-Key", fvkey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	if len(body) != 0 {
		log.Fatal("response was not 200!")
	}
	return true
}

type PGRequest struct {
	Username string
	UserData string
}

type Person struct {
	PersonId string
}

func createPerson(username string, userinfo string) *Person {
	var endpoint = fvendpoint + "/persongroups/" + persongid + "/persons"

	var msReq = PGRequest{username, userinfo}

	postData, err := json.Marshal(msReq)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(postData))
	req.Header.Set("Ocp-Apim-Subscription-Key", fvkey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var person Person
	json.NewDecoder(resp.Body).Decode(&person)
	return &person
}


type NewFaceRequest struct {
	Url string
}

type NewFaceResponse struct {
	PersistedFaceId string
}


func addPersonFace(personid string, url string) *NewFaceResponse {
	var endpoint = fvendpoint + "/persongroups/" + persongid + "/persons/" + personid + "/persistedFaces"

	var msReq = NewFaceRequest{url}

	postData, err := json.Marshal(msReq)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(postData))
	req.Header.Set("Ocp-Apim-Subscription-Key", fvkey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var face NewFaceResponse
	json.NewDecoder(resp.Body).Decode(&face)
	return &face
}


func trainPersonGroup() bool {
	var endpoint = fvendpoint + "/persongroups/" + persongid + "/train"

	req, err := http.NewRequest("POST", endpoint, nil)
	req.Header.Set("Ocp-Apim-Subscription-Key", fvkey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	if len(body) != 0 {
		log.Fatal("response was not 200!")
	}
	return true
}
