package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"time"
)

//clientID
//林亦潔 592fa46f290e4f1aa8b5768bbb802177
//歐哲熏 ab43d6c3cbdc479ca53096f213e19f2a

//clientSecret
//林亦潔 4ddd10a13f2a4c00af97c1916b21a8c2
//歐哲熏 5e3c41d08b5f467799668a62b566aa18 	

var (
	clientID     = "592fa46f290e4f1aa8b5768bbb802177"
	clientSecret = "4ddd10a13f2a4c00af97c1916b21a8c2"
	redirectURI  = "http://localhost:8080/callback"
	// 向使用者要求的授權範圍
	scope      = "user-read-private user-read-email  user-top-read playlist-modify-public playlist-modify-private"
	ARTISTNAME = "King gnu"
	TRACKNAME  = "Supernova"
	//state        = "randomStateString"   // 隨機字串，用於防止 CSRF 攻擊
)

var (
	currentAccessToken  string                  // 保存當前的 Access Token
	currentRefreshToken string                  // 保存 Refresh Token
	tokenExpiresAt      int64                   // Access Token 過期的 Unix 時間戳
	processedRequests   = make(map[string]bool) // 紀錄處理過的授權碼
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type User struct {
	Name       string
	SpotifyURL string
	ImageURL   string
	UserID     string
}

type Singer struct {
	SingerID  string
	Name      string
	TopTracks []Track
	ImageURL  string
}

type Playlist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	EmbedURL string `json:"embedurl"`
	ExternalURL string `json:"externalurl"`
	TrackURIs []string `json:"trackurls"`
}

type Track struct {
	Name       string
	ID         string
	URL        string
	ImageURL   string
	PreviewURL string
	Album      Album
	Singer     string
}

type TemplateData struct {
	UserData        User2
	SingerData      Singer
	PlaylistData    Playlist
	SearchTrackData Track
}
type Album struct {
	ID           string
	Name         string
	Release_date string
	Total_teacks int
}

var userdata User2
var singerdata Singer
var trackdata Track
var playlistdata  = Playlist{
	ExternalURL: "http://localhost:8080/favorite",
}
var playlistpointer *Playlist
var token *TokenResponse

// 連接html
/*func getUserInfo(userInfo map[string]interface{}) {
	externalUrls := userInfo["external_urls"].(map[string]interface{})
	images := userInfo["images"].([]interface{})

	userdata = User{
		Name:       userInfo["display_name"].(string),
		SpotifyURL: externalUrls["spotify"].(string),
		ImageURL:   images[0].(map[string]interface{})["url"].(string),
		UserID:     userInfo["id"].(string),
	}
}*/

func handler(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("userInfo.html")
	if err != nil {
		http.Error(w, "無法加載模板", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	data := TemplateData{
		UserData:        userdata,
		SingerData:      singerdata,
		SearchTrackData: trackdata,
		PlaylistData:    playlistdata,
	}

	err = temp.Execute(w, data)
	if err != nil {
		http.Error(w, "渲染模板失敗", http.StatusInternalServerError)
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
	params.Set("scope", scope) // 權限範圍

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

// 啟動伺服器，接收授權碼並交換 Token
func startServer() {

	http.HandleFunc("/userinfo", handler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		authURL := generateAuthURL()
		//fmt.Fprintf(w, "請點擊以下連結進行授權：<a href='%s'>Spotify 授權</a>", authURL)
		http.Redirect(w, r, authURL, http.StatusSeeOther)
	})

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// 驗證 state
		/*if r.URL.Query().Get("state") != state {
			http.Error(w, "狀態不符", http.StatusForbidden)
			return
		}*/

		// 獲取授權碼
		code := r.URL.Query().Get("code")
		//fmt.Println("token: ", code)
		if code == "" {
			http.Error(w, "未獲取到授權碼", http.StatusBadRequest)
			return
		}

		// 防止重複處理請求
		if processedRequests[code] {
			http.Error(w, "請求已處理", http.StatusBadRequest)
			return
		}
		processedRequests[code] = true
		defer func() { delete(processedRequests, code) }()

		// 使用授權碼交換 Token
		token, err := exchangeCodeForToken(code)
		if err != nil {
			http.Error(w, "無法獲取 Token", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		// 調用 Spotify API
		// 1. get user information
		err = getCurrentUserInfo(token.AccessToken)
		if err != nil {
			log.Println("取得UserInfo失敗: ", err)
			http.Error(w, "無法取得UserInfo", http.StatusInternalServerError)
			return
		}
		// 2. serach artist and top tracks
		err = searchArtist(ARTISTNAME, token.AccessToken)
		if err != nil {
			log.Println("搜索歌手失敗:", err)
			http.Error(w, "無法搜索歌手", http.StatusInternalServerError)
			return
		}
		// 3. create playlist and add tracks
		exists, playlistID, err := playlistExists(token.AccessToken, "我的收藏")
		if err != nil {
			log.Println("檢查播放清單失敗:", err)
			http.Error(w, "檢查播放清單失敗", http.StatusInternalServerError)
			return
		}

		if exists {
			fmt.Println("播放清單已存在，ID:", playlistID)
			playlistdata.ID = playlistID
			playlistdata.Name = "我的收藏"
		} else {
			playlistpointer, err = createPlaylist(userdata.UserID, "我的收藏", "From SpotGoInsight")
			if err != nil {
				log.Println("新增播放清單失敗: ", err)
				http.Error(w, "無法新增播放清單", http.StatusInternalServerError)
				return
			}
			playlistdata.ID = playlistpointer.ID
			playlistdata.Name = playlistpointer.Name
		}

		// 新增 Tracks 到播放清單
		err = addTracksToPlaylist(playlistdata.ID, playlistdata.TrackURIs)
		if err != nil {
			log.Println("新增歌曲到播放清單失敗: ", err)
			http.Error(w, "無法新增歌曲到播放清單", http.StatusInternalServerError)
			return
		}
		// 4. search track
		err = searchTrack(TRACKNAME, token.AccessToken, &trackdata)
		if err != nil {
			log.Println("搜索歌曲失敗:", err)
			http.Error(w, "無法搜尋歌曲", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/userinfo", http.StatusSeeOther)
		/*fmt.Fprintf(w, "Spotify API 呼叫成功！用戶資訊：%v", userInfo)
		fmt.Fprintf(w, "Singer ID API 呼叫成功！歌手 %s 的ID：%v", ARTISTNAME, SingerID)*/
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

	currentAccessToken = tokenResponse.AccessToken
	currentRefreshToken = tokenResponse.RefreshToken
	tokenExpiresAt = time.Now().Unix() + int64(tokenResponse.ExpiresIn) // 計算過期時間

	return &tokenResponse, nil
}

// refresh token
func refreshAccessToken(refreshToken string) (string, int64, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", 0, err
	}

	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("Token 刷新失敗，狀態碼: %d", resp.StatusCode)
	}

	var response TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", 0, err
	}

	newExpiresAt := time.Now().Unix() + int64(response.ExpiresIn)
	return response.AccessToken, newExpiresAt, nil
}

// ensure token valid
func ensureValidAccessToken() error {
	// 如果 Access Token 已過期，刷新它
	if time.Now().Unix() >= tokenExpiresAt {
		fmt.Println("Access Token 過期，正在刷新...")
		newToken, newExpiresAt, err := refreshAccessToken(currentRefreshToken)
		if err != nil {
			return fmt.Errorf("刷新 Token 失敗: %v", err)
		}

		// 更新全局變數
		currentAccessToken = newToken
		tokenExpiresAt = newExpiresAt
		fmt.Println("Access Token 已刷新")
	}
	return nil
}

// API: get user information
func getCurrentUserInfo(accessToken string) error {
	// 確保 Token 有效
	if err := ensureValidAccessToken(); err != nil {
		return err
	}

	// 繼續使用有效 Token 調用 API
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me", nil)
	if err != nil {
		return err
	}

	// 設置授權標頭，使用 Bearer Token 認證
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// 創建一個 HTTP 客戶端用於發送請求
	client := &http.Client{}
	// 發送請求並接收回應
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	// 確保在函式結束時關閉回應的主體，避免資源泄漏
	defer resp.Body.Close()

	// 檢查回應的 HTTP 狀態碼
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API 請求失敗，狀態碼: %d", resp.StatusCode)
	}

	// 定義一個通用的 map 用於存放解析後的用戶資訊
	var userInfo map[string]interface{}
	// 使用 JSON 解碼器將回應主體中的 JSON 資料解碼到 userInfo map 中
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return err
	}

	//arrange userinfo
	externalUrls := userInfo["external_urls"].(map[string]interface{})
	images := userInfo["images"].([]interface{})

	userdata = User2{
		Name:       userInfo["display_name"].(string),
		SpotifyURL: externalUrls["spotify"].(string),
		ImageURL:   images[0].(map[string]interface{})["url"].(string),
		UserID:     userInfo["id"].(string),
	}

	return nil
}

// API: search artist
func searchArtist(ARTISTNAME string, accessToken string) error {
	//fmt.Println("---search---")

	if err := ensureValidAccessToken(); err != nil {
		return err
	}

	baseSearchURL := "https://api.spotify.com/v1/search"
	params := url.Values{}
	//params.Set("q", ARTISTNAME)
	params.Set("q", fmt.Sprintf(`"%s"`, ARTISTNAME))
	params.Set("type", "artist")
	params.Set("limit", "1")

	searchURL := fmt.Sprintf("%s?%s", baseSearchURL, params.Encode())

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("搜索 API 請求失敗，錯誤碼: %d", resp.StatusCode)
	}

	var searchResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return err
	}

	artists := searchResult["artists"].(map[string]interface{})
	items := artists["items"].([]interface{})
	if len(items) == 0 {
		return fmt.Errorf("未找到名為 '%s' 的歌手", ARTISTNAME)
	}

	// 取得第一個歌手
	artist := items[0].(map[string]interface{})
	artistID := artist["id"].(string)
	artistName := artist["name"].(string)

	var artistImageURL string
	if images, ok := artist["images"].([]interface{}); ok && len(images) > 0 {
		image := images[0].(map[string]interface{})
		artistImageURL = image["url"].(string)
	} else {
		artistImageURL = "未提供圖片"
	}

	topTracksURL := fmt.Sprintf("https://api.spotify.com/v1/artists/%s/top-tracks?market=US", artistID)

	req, err = http.NewRequest("GET", topTracksURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Top Tracks API 請求失敗，錯誤碼: %d", resp.StatusCode)
	}

	var topTracksResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&topTracksResult); err != nil {
		return err
	}

	tracks := topTracksResult["tracks"].([]interface{})
	var topTracks []Track
	for _, track := range tracks {
		trackInfo := track.(map[string]interface{})
		trackName := trackInfo["name"].(string)
		trackImageURL := trackInfo["album"].(map[string]interface{})["images"].([]interface{})[0].(map[string]interface{})["url"].(string)
		previewURL, _ := trackInfo["preview_url"].(string)

		topTracks = append(topTracks, Track{
			Name:       trackName,
			ImageURL:   trackImageURL,
			PreviewURL: previewURL,
		})
		fmt.Printf(" %s: %s\n", trackName, previewURL)
	}

	singerdata.SingerID = artistID
	singerdata.Name = artistName
	singerdata.TopTracks = topTracks
	singerdata.ImageURL = artistImageURL

	return nil
}

func createPlaylist(userID string, playlistName, playlistDescription string) (*Playlist, error) {
	//fmt.Println("===start create playlist===")
	// 確保 Access Token 有效
	if err := ensureValidAccessToken(); err != nil {
		return nil, err
	}

	// 建立請求資料
	data := map[string]interface{}{
		"name":        playlistName,        // 播放清單名稱
		"description": playlistDescription, // 播放清單描述
		"public":      false,               // 設置為私人播放清單(true: public playlist)
	}
	// 將資料轉換為 JSON 格式
	body, _ := json.Marshal(data)

	// 構建 POST 請求 URL
	url := fmt.Sprintf("https://api.spotify.com/v1/users/%s/playlists", userID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body)) // 傳送 JSON 資料
	if err != nil {
		return nil, err
	}

	// 設置請求標頭
	req.Header.Set("Authorization", "Bearer "+currentAccessToken) // 使用 Bearer Token 認證
	req.Header.Set("Content-Type", "application/json")            // 指定內容類型為 JSON

	// 創建 HTTP 客戶端並發送請求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // 確保回應主體在函數結束前關閉

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create playlist, status: %d", resp.StatusCode)
	}

	// 解析回應資料到 Playlist 結構
	var playlist Playlist
	if err := json.NewDecoder(resp.Body).Decode(&playlist); err != nil {
		return nil, err
	}

	log.Println("已新增播放清單")
	return &playlist, nil
}

// API: add Tracks To Playlist
func addTracksToPlaylist(playlistID string, favoriteTrackURIs []string) error {
	//fmt.Println("===start add tracks to playlist===")
	// 確保 Access Token 有效
	if err := ensureValidAccessToken(); err != nil {
		return err
	}

	// 建立請求資料
	data := map[string]interface{}{
		"uris": favoriteTrackURIs, // 包含 Tracks URI 的陣列
	}
	// 將資料轉換為 JSON 格式
	body, _ := json.Marshal(data)

	// 構建 POST 請求 URL
	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlistID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body)) // 傳送 JSON 資料
	if err != nil {
		return err
	}

	// 設置請求標頭
	req.Header.Set("Authorization", "Bearer "+currentAccessToken) // 使用 Bearer Token 認證
	req.Header.Set("Content-Type", "application/json")            // 指定內容類型為 JSON

	// 創建 HTTP 客戶端並發送請求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() // 確保回應主體在函數結束前關閉

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to add tracks, status: %d", resp.StatusCode)
	}

	return nil
}

// check if playlist exists
func playlistExists(accessToken, playlistName string) (bool, string, error) {
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me/playlists", nil)
	if err != nil {
		return false, "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("failed to fetch playlists, status: %d", resp.StatusCode)
	}

	var playlists struct {
		Items []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&playlists); err != nil {
		return false, "", err
	}

	for _, playlist := range playlists.Items {
		if playlist.Name == playlistName {
			return true, playlist.ID, nil
		}
	}

	return false, "", nil
}

func searchTrack(trackName string, accessToken string, inputTracks *Track) error {
	if err := ensureValidAccessToken(); err != nil {
		return err
	}

	baseSearchURL := "https://api.spotify.com/v1/search"
	params := url.Values{}
	//params.Set("q", trackName)
	params.Set("q", fmt.Sprintf(`"%s"`, trackName))
	params.Set("type", "track")
	params.Set("limit", "1")

	searchURL := fmt.Sprintf("%s?%s", baseSearchURL, params.Encode())

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("搜尋歌曲ID API 请求失败，状态码: %d", resp.StatusCode)
	}

	var searchResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return err
	}

	track_result := searchResult["tracks"].(map[string]interface{})
	items := track_result["items"].([]interface{})
	if len(items) == 0 {
		return fmt.Errorf("未找到名为 '%s' 的歌曲", ARTISTNAME)
	}

	track := items[0].(map[string]interface{})
	trackID := track["id"].(string)
	inputTracks.ID = trackID
	inputTracks.Name = track["name"].(string)
	inputTracks.PreviewURL, _ = track["preview_url"].(string)

	// Extract album image URL and other details
	album := track["album"].(map[string]interface{})
	inputTracks.Album.Name = album["name"].(string)
	inputTracks.Album.ID = album["id"].(string)
	inputTracks.Album.Release_date = album["release_date"].(string)

	images := album["images"].([]interface{})
	image := images[0].(map[string]interface{})
	inputTracks.ImageURL, _ = image["url"].(string)

	artists := track["artists"].([]interface{})
	artist := artists[0].(map[string]interface{})
	inputTracks.Singer, _ = artist["name"].(string)

	externalURL := track["external_urls"].(map[string]interface{})
	inputTracks.URL, _ = externalURL["spotify"].(string)

	/*fmt.Println("track name: ", inputTracks.Name)
	fmt.Println("ID: ", inputTracks.ID)
	fmt.Println("URL: ", inputTracks.URL)
	fmt.Println("ImageURL: ", inputTracks.ImageURL)
	fmt.Println("PreviewURL: ", inputTracks.PreviewURL)
	fmt.Println("Singer: ", inputTracks.Singer)
	fmt.Println("year: ", inputTracks.Album.Release_date)*/

	// Log or store the track image and other details
	//fmt.Printf("Track: %s\nImage: %s\n", trackName, imageURL)
	//fmt.Printf("\nSearch result:  TestgetTracks trackname: %s   trackURL: %s  trackID: %s  trackPreview: %s\n", trackdata.Name, trackdata.URL, trackdata.ID, trackdata.PreviewURL)
	//fmt.Printf(" Album name: %s  Album release date: %s  the singer is %s\n", inputTracks.Album.Name, inputTracks.Album.Release_date, sing_name) //測試專輯的使用
	return nil
}

/*func main() {
	startServer()
}*/
