package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

var (
	clientID     = "592fa46f290e4f1aa8b5768bbb802177" // 替換為您的 Spotify Client ID
	clientSecret = "4ddd10a13f2a4c00af97c1916b21a8c2" // 替換為您的 Spotify Client Secret
	redirectURI  = "http://localhost:8086/callback"   // 替換為您設定的 Redirect URI
	//state        = "randomStateString"   // 隨機字串，用於防止 CSRF 攻擊
)

// 生成授權連結
func generateAuthURL() string {
	baseURL := "https://accounts.spotify.com/authorize"
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("response_type", "code")
	params.Set("redirect_uri", redirectURI)
	//params.Set("state", state)
	//params.Set("scope", "user-read-private user-read-email") // 權限範圍

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

// 啟動伺服器，接收授權碼並交換 Token
func startServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		authURL := generateAuthURL()
		fmt.Fprintf(w, "請點擊以下連結進行授權：<a href='%s'>Spotify 授權</a>", authURL)
	})

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// 驗證 state
		/*if r.URL.Query().Get("state") != state {
			http.Error(w, "狀態不符", http.StatusForbidden)
			return
		}*/

		// 獲取授權碼
		code := r.URL.Query().Get("code")
		log.Println("token=", code)
		if code == "" {
			http.Error(w, "未獲取到授權碼", http.StatusBadRequest)
			return
		}

		// 使用授權碼交換 Token
		token, err := exchangeCodeForToken(code)
		if err != nil {
			http.Error(w, "無法獲取 Token", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		// 調用 Spotify API
		userInfo, err := getCurrentUserInfo(token.AccessToken)
		if err != nil {
			http.Error(w, "無法調用 Spotify API", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		fmt.Fprintf(w, "Spotify API 呼叫成功！用戶資訊：%v", userInfo)
	})

	fmt.Println("伺服器啟動於 http://localhost:8086")
	log.Fatal(http.ListenAndServe(":8086", nil))
}

// 使用授權碼交換訪問 Token
func exchangeCodeForToken(code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token 請求失敗，狀態碼: %d", resp.StatusCode)
	}

	var tokenResponse TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		return nil, err
	}

	return &tokenResponse, nil
}

// 調用 Spotify API 獲取當前用戶資訊
func getCurrentUserInfo(accessToken string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 請求失敗，狀態碼: %d", resp.StatusCode)
	}

	var userInfo map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}

// Token 回應結構體
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func main() {
	startServer()
}
