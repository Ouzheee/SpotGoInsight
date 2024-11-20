package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
)

var (
	clientID     = "592fa46f290e4f1aa8b5768bbb802177" // 替換為您的 Spotify Client ID
	clientSecret = "4ddd10a13f2a4c00af97c1916b21a8c2" // 替換為您的 Spotify Client Secret
	redirectURI  = "http://localhost:8086/callback"   // 替換為您設定的 Redirect URI
	artistName   = "King gnu"
	//state        = "randomStateString"   // 隨機字串，用於防止 CSRF 攻擊
)

type user struct {
	Name       string
	SpotifyURL string
	ImageURL   string
	UserID     string
}

var userdata user

func getUserInfo(userInfo map[string]interface{}) {
	externalUrls := userInfo["external_urls"].(map[string]interface{})
	images := userInfo["images"].([]interface{})

	userdata = user{
		Name:       userInfo["display_name"].(string),
		SpotifyURL: externalUrls["spotify"].(string),
		ImageURL:   images[0].(map[string]interface{})["url"].(string),
		UserID:     userInfo["id"].(string),
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("userInfo.html")
	if err != nil {
		http.Error(w, "無法加載模板", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	err = temp.Execute(w, userdata)
	if err != nil {
		http.Error(w, "無法加載模板", http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

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
		//fmt.Fprintf(w, "請點擊以下連結進行授權：<a href='%s'>Spotify 授權</a>", authURL)
		http.Redirect(w, r, authURL, http.StatusSeeOther)
	})

	http.HandleFunc("/userinfo", handler)

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
			http.Error(w, "無法調用 UserInfo API", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		signerID, err := searchArtist(artistName, token.AccessToken)
		if err != nil {
			http.Error(w, "無法調用 Signer ID API", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		http.Redirect(w, r, "/userinfo", http.StatusSeeOther)
		fmt.Fprintf(w, "Spotify API 呼叫成功！用戶資訊：%v", userInfo)
		fmt.Fprintf(w, "Signer ID API 呼叫成功！歌手 %s 的ID：%v", artistName, signerID)
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
	// 設定重定向 URI，必須與 Spotify
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
	getUserInfo(userInfo)

	return userInfo, nil
}

func searchArtist(artistName string, accessToken string) (string, error) {
	// 編碼搜尋參數
	baseURL := "https://api.spotify.com/v1/search"
	params := url.Values{}
	params.Set("q", artistName)  // 搜索的關鍵字
	params.Set("type", "artist") // 資料類型為歌手
	params.Set("limit", "1")     // 只需要第一個結果

	// 建立完整的請求 URL
	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// 建立 HTTP 請求
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return "", err
	}

	// 設定 Authorization Header
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// 發送請求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 確認回應狀態碼
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("搜索 API 請求失敗，狀態碼: %d", resp.StatusCode)
	}

	// 解析 JSON 回應
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	// 提取歌手 ID
	artists := result["artists"].(map[string]interface{})
	items := artists["items"].([]interface{})
	if len(items) == 0 {
		return "", fmt.Errorf("未找到名為 '%s' 的歌手", artistName)
	}

	artist := items[0].(map[string]interface{})
	artistID := artist["id"].(string)
	return artistID, nil
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
