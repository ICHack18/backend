package main

import (
	"encoding/json"
	"net/http"
	"bytes"
	"io/ioutil"
	"log"
)

func getFaceVerification(url string, faceids []string) (*FVResponse, error) {
	var msReq = FVRequest{faceids, persongid}

	postData, err := json.Marshal(msReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fvendpoint, bytes.NewReader(postData))
	req.Header.Set("Ocp-Apim-Subscription-Key", fvkey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := makeAPICallWithBackoff(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var response FVResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return &response, nil
}


func createPersonGroup(name string, info string) (bool, error) {
	var endpoint = fvendpoint + "/persongroups/" + persongid

	//var msReq = NewPGRequest{username, data}
	var msReq = NewPGRequest{name, info}

	postData, err := json.Marshal(msReq)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(postData))
	req.Header.Set("Ocp-Apim-Subscription-Key", fvkey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := makeAPICallWithBackoff(req)

	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if len(body) != 0 {
		log.Fatal("response was not 200!")
	}
	return true, nil
}


func createPerson(username string, userinfo string) (*Person, error) {
	var endpoint = fvendpoint + "/persongroups/" + persongid + "/persons"

	var msReq = PGRequest{username, userinfo}

	postData, err := json.Marshal(msReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(postData))
	req.Header.Set("Ocp-Apim-Subscription-Key", fvkey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := makeAPICallWithBackoff(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var person Person
	json.NewDecoder(resp.Body).Decode(&person)
	return &person, nil
}


func addPersonFace(personid string, url string) (*NewFaceResponse, error) {
	var endpoint = fvendpoint + "/persongroups/" + persongid + "/persons/" + personid + "/persistedFaces"

	var msReq = NewFaceRequest{url}

	postData, err := json.Marshal(msReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(postData))
	req.Header.Set("Ocp-Apim-Subscription-Key", fvkey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := makeAPICallWithBackoff(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var face NewFaceResponse
	json.NewDecoder(resp.Body).Decode(&face)
	return &face, nil
}


func trainPersonGroup() (bool, error) {
	var endpoint = fvendpoint + "/persongroups/" + persongid + "/train"

	req, err := http.NewRequest("POST", endpoint, nil)
	req.Header.Set("Ocp-Apim-Subscription-Key", fvkey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := makeAPICallWithBackoff(req)

	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if len(body) != 0 {
		log.Fatal("response was not 200!")
	}
	return true, nil
}


func detectFaces(url string) (*[]Faces, error) {
	var endpoint = fvendpoint + "/detect"

	var msReq = NewFaceRequest{url}

	postData, err := json.Marshal(msReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(postData))
	req.Header.Set("Ocp-Apim-Subscription-Key", fvkey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := makeAPICallWithBackoff(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var faces []Faces
	json.NewDecoder(resp.Body).Decode(&faces)
	return &faces, nil
}
