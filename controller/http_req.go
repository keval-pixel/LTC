package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/fatih/color"
)

type Request_body struct {
	Data map[string]interface{} `json:"data"`
}

func http_get(endpoint string) []byte {
	url := "http://localhost:6565/"

	response, err := http.Get(url + endpoint)
	if err != nil {
		color.Red("The HTTP request failed with error %s\n", err)
		return nil
	}

	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		color.Red("Failed to read the response body: %s\n", err.Error())
		return nil
	}

	return body

}

func http_send(Endpoint string, Method string, jsonData []byte) {
	if Method != "PUT" && Method != "POST" && Method != "PATCH" {
		color.Red("Invalid HTTP Method. Should be PUT, POST or PATCH")
		return
	}
	url := "http://localhost:6565/" + Endpoint

	req, err := http.NewRequest(Method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		color.Red("Failed to create request: %s\n", err)
		return
	}

	// Set custom headers
	req.Header.Set("Content-Type", "application/json")

	// Create an HTTP client and set timeout
	client := &http.Client{}

	response, err := client.Do(req)
	if err != nil {
		color.Red("The HTTP request failed with error %s\n", err)
		return
	}

	defer response.Body.Close()

	// Read the response body
	_, err = io.ReadAll(response.Body)
	if err != nil {
		color.Red("Failed to read the response body: %s\n", err)
		return
	}
	// Print the response body
	//k6_get_status()

}

func k6_stop() {
	requestData := Request_body{
		Data: map[string]interface{}{
			"type": "status",
			"id":   "default",
			"attributes": map[string]interface{}{
				"stopped": true,
			},
		},
	}

	// Marshal the struct to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		color.Red("Failed to marshal JSON: %s\n", err)
		return
	}

	http_send("v1/status", "PATCH", jsonData)

	color.Yellow("Stopping...")

}

func k6_continue() {
	requestData := Request_body{
		Data: map[string]interface{}{
			"type": "status",
			"id":   "default",
			"attributes": map[string]interface{}{
				"running": true,
			},
		},
	}

	// Marshal the struct to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		color.Red("Failed to marshal JSON: %s\n", err)
		return
	}

	http_send("v1/status", "PATCH", jsonData)

}

func k6_update_user(next_vus int32) {

	if next_vus < 0 {
		color.Red("Error: Cannot specify a non-positive value for K6 users")
		return
	}
	requestData := Request_body{
		Data: map[string]interface{}{
			"type": "status",
			"id":   "default",
			"attributes": map[string]interface{}{
				"vus": next_vus,
			},
		},
	}

	// Marshal the struct to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		color.Red("Failed to marshal JSON: %s\n", err)
		return
	}
	http_send("v1/status", "PATCH", jsonData)
}

func get_vus() int32 {
	data := http_get("v1/status")
	if data == nil {
		color.Red("Error getting status from K6")
	}
	vus := get_key_value("vus", data)
	if vus < 0 {
		color.Red("Error getting vus from K6")
	}

	return vus
}

func k6_get_status() {
	var data interface{}

	jsonData := http_get("v1/status")

	err := json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	// Marshal the data again with indentation for pretty-printing
	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}

	// Print the pretty JSON output
	fmt.Println(string(prettyJSON))
}
