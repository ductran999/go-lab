package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func main() {
	backendURL := "http://localhost:8080"

	// Step 1: Request Access Token via Client Credentials Grant
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", "billing_service")
	data.Set("client_secret", "billing_secret_999")

	resp, err := http.PostForm(backendURL+"/oauth/token", data)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	var tokenResult map[string]any
	json.NewDecoder(resp.Body).Decode(&tokenResult)
	accessToken := tokenResult["access_token"].(string)

	fmt.Println("Successfully obtained Access Token!")

	// Step 2: Use the Access Token to call the protected /api/v1/reports API
	client := &http.Client{}
	req, _ := http.NewRequest("GET", backendURL+"/api/v1/reports", nil)

	// Attach the Token as a Bearer token
	req.Header.Set("Authorization", "Bearer "+accessToken)

	apiResp, err := client.Do(req)
	if err != nil {
		fmt.Println("API Error:", err)
		return
	}
	defer apiResp.Body.Close()

	body, _ := io.ReadAll(apiResp.Body)
	fmt.Println("API Response:\n", string(body))
}
