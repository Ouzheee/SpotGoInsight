package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	clientID     = "7178993402ab428dbdcb349934eda41e"  // 替換為您的 Spotify Client ID
	clientSecret = "80d3360b46a04436b412f42aa3d26029"  // 替換為您的 Spotify Client Secret
	redirectURI  = "http://localhost:8086"             // Spotify 授權完成後的返回網址
	scope        = "user-read-private user-read-email" // 所需權限
	authURL      = "https://accounts.spotify.com/authorize"
	tokenURL     = "https://accounts.spotify.com/api/token"
)

var state = "randomstate" // 用於防止 CSRF 攻擊

func main() {
	// 設定 HTTP 路由
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/callback", handleCallback)

	// 啟動伺服器
	fmt.Println("伺服器啟動於 http://localhost:8086")
	log.Fatal(http.ListenAndServe(":8086", nil))
}

// 首頁處理：引導使用者到 Spotify 授權頁面
func handleIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Redirecting user to Spotify login page")
	authURL := fmt.Sprintf("%s?client_id=%s&response_type=code&redirect_uri=%s&scope=%s&state=%s",
		authURL, clientID, url.QueryEscape(redirectURI), url.QueryEscape(scope), state)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// 授權回呼處理：接收 Spotify 的授權碼並交換 Access Token
func handleCallback(w http.ResponseWriter, r *http.Request) {
	// 檢查是否有錯誤或缺少授權碼
	if r.URL.Query().Get("error") != "" {
		http.Error(w, "授權失敗", http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	receivedState := r.URL.Query().Get("state")
	if code == "" || receivedState != state {
		http.Error(w, "授權碼或狀態無效", http.StatusBadRequest)
		return
	}

	// 交換 Access Token
	token, err := exchangeToken(code)
	if err != nil {
		http.Error(w, "無法交換 Access Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 回傳 Access Token 結果
	response := struct {
		AccessToken string `json:"access_token"`
	}{
		AccessToken: token.AccessToken,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 與 Spotify 伺服器交換 Access Token
func exchangeToken(code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

// Access Token 回應結構
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}
