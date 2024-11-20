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
	clientID     = "ab43d6c3cbdc479ca53096f213e19f2a" // 替換為您的 Spotify Client ID
	clientSecret = "5e3c41d08b5f467799668a62b566aa18" // 替換為您的 Spotify Client Secret
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
		fmt.Println("token: ", code)
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
	// 設定請求參數：授權類型為 "authorization_code"
	data.Set("grant_type", "authorization_code")
	// 將從回調中獲取的授權碼加入請求參數
	data.Set("code", code)
	// 設定重定向 URI，必須與 Spotify 開發者應用程式設定中的一致
	data.Set("redirect_uri", redirectURI)

	// 建立一個新的 POST 請求，目標為 Spotify 的 Token 端點
	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}

	// 設定請求的內容類型為 "application/x-www-form-urlencoded"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// 使用基本身份驗證將 Client ID 和 Client Secret 加入請求頭
	req.SetBasicAuth(clientID, clientSecret)

	// 建立一個 HTTP 客戶端來發送請求
	client := &http.Client{}
	// 發送請求並接收回應
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	// 確保在函式結束前關閉回應主體
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token 請求失敗，狀態碼: %d", resp.StatusCode)
	}

	var tokenResponse TokenResponse
	// 將回應的 JSON 數據解碼到 tokenResponse 結構體中
	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		return nil, err
	}

	return &tokenResponse, nil
}

// 調用 Spotify API 獲取當前用戶資訊
func getCurrentUserInfo(accessToken string) (map[string]interface{}, error) {
	// 建立一個新的 GET 請求，目標為 Spotify 的當前用戶端點
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me", nil)
	if err != nil {
		return nil, err
	}

	// 設置授權標頭，使用 Bearer Token 認證
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// 創建一個 HTTP 客戶端用於發送請求
	client := &http.Client{}
	// 發送請求並接收回應
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	// 確保在函式結束時關閉回應的主體，避免資源泄漏
	defer resp.Body.Close()

	// 檢查回應的 HTTP 狀態碼
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 請求失敗，狀態碼: %d", resp.StatusCode)
	}

	// 定義一個通用的 map 用於存放解析後的用戶資訊
	var userInfo map[string]interface{}
	// 使用 JSON 解碼器將回應主體中的 JSON 資料解碼到 userInfo map 中
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return nil, err
	}

	//test print
	printUserInfo(userInfo)

	return userInfo, nil
}

func printUserInfo(userInfo map[string]interface{}) {
	name := userInfo["display_name"].(string)
	externalUrls := userInfo["external_urls"].(map[string]interface{})
	spotifyURL := externalUrls["spotify"].(string)
	images := userInfo["images"].([]interface{})
	imageURL := images[0].(map[string]interface{})["url"].(string)
	userID := userInfo["id"].(string)

	fmt.Println("1. user name:", name)
	fmt.Println("2. external url: ", spotifyURL)
	fmt.Println("3. image url: ", imageURL)
	fmt.Println("4. user ID: ", userID)
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
