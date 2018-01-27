package main

import (
	"encoding/json"
	"net/http"
	"bytes"
	"io/ioutil"
	"log"
)

func getFaceVerification(url string, faceids []string) *FVResponse {
	var msReq = FVRequest{faceids, persongid}

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


func createPersonGroup(name string, info string) bool {
	var endpoint = fvendpoint + "/persongroups/" + persongid

	//var msReq = NewPGRequest{username, data}
	var msReq = NewPGRequest{name, info}

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


func detectFaces(url string) *[]Faces {
	var endpoint = fvendpoint + "/detect"

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

	var faces []Faces
	json.NewDecoder(resp.Body).Decode(&faces)
	return &faces
}