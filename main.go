package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const (
	endpoint = "https://westcentralus.api.cognitive.microsoft.com/vision/v1.0/analyze?visualFeatures=categories,description&language=en"
	fvendpoint = "https://westcentralus.api.cognitive.microsoft.com/face/v1.0"
	persongid = "banned_users"
)
var key = os.Getenv("VISION_KEY")
var fvkey = os.Getenv("FACE_VISION_KEY")


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
	r.HandleFunc("/newuser", newUser)
	r.HandleFunc("/adduserphoto", newUserPhoto)
	r.HandleFunc("/traingroups", trainGroups)
	r.HandleFunc("/listusers", listUsers)

	loggedRouter := handlers.LoggingHandler(os.Stdout, r)

	log.Fatal(http.ListenAndServe(":2048", loggedRouter))
}

func ensure_groups_created() {
	// TODO
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
				fetchNew = false
				marshalErr := json.Unmarshal([]byte(val), &cvResponse)
				if marshalErr != nil {
					http.Error(w, marshalErr.Error(), 500)
					return
				}
			}
		}

		if fetchNew {
			cvResponse, err = getDescriptionFromCognitiveServices(url)
			if err != nil {
				http.Error(w, err.Error(), 429)
				return
			}
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
	response, err := getDescriptionFromCognitiveServices(url)
	if err != nil {
		http.Error(w, err.Error(), 429)
		return
	}
	responseJson, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJson)
}

func getDescriptionFromCognitiveServices(url string) (cvr CVResponse, err error) {
	var msReq = &CVRequest{
		url,
	}

	postData, err := json.Marshal(msReq)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(postData))
	req.Header.Set("Ocp-Apim-Subscription-Key", key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := makeAPICallWithBackoff(req)

	if err != nil {
		return
	}

	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&cvr)

	return
}

// TODO: add face rec to this
func shouldBlockImage(blockTags []string, cvResponse CVResponse) bool {
	// If image is too small then the width or height is set to 0 but shouldn't be hidden
	if cvResponse.Metadata.Width == 0 || cvResponse.Metadata.Height == 0 {
		return false
	}

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

	// NOTE: Where to put this?
	// TODO: detect faces, get first 10 faceids, and then
	//faces := detectFaces(photoUrl)
	//for face := range *faces {
	//	face.FaceId
	//}
	// TODO: then call
	//getFaceVerification(photoUrl)
	return false
}

func newUser(w http.ResponseWriter, r *http.Request) {
	//username := r.URL.Query().Get("username")
	//userinfo := r.URL.Query().Get("userinfo")
	//person := createPerson(username, userinfo)
	// TODO: something with person ID?
	// TODO: respond with updated list of users-> redirect to listUsers?
}

func newUserPhoto(w http.ResponseWriter, r *http.Request) {
	//personId  := r.URL.Query().Get("personId")
	//photoUrl := r.URL.Query().Get("url")
	//newFace := addPersonFace(personId, photoUrl)
	// TODO: store persistedFaceId somewhere -> list of faces that can occur in a photo-> max 10 cos demo
}

func trainGroups(w http.ResponseWriter, r *http.Request) {
	// TODO: loop over all keys and retrain group for each
	//trainPersonGroup()
	// TODO: some spinner or something to mark busy? idk
}

func listUsers(w http.ResponseWriter, r *http.Request) {
	// each user needs to have its personId

}

func makeAPICallWithBackoff(req *http.Request) (resp *http.Response, err error) {
	client := &http.Client{}

	var backoffTimeout = 0.5
	resp, err = client.Do(req)
	for {
		if err == nil {
			return
		}
		time.Sleep(time.Duration(backoffTimeout) * time.Second)
		resp, err = client.Do(req)
		backoffTimeout *= 2
		if backoffTimeout == 8 {
			err = errors.New("API Timeout")
			return
		}
	}
}
