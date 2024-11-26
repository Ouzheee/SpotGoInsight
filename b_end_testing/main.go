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

var (
	clientID     = "592fa46f290e4f1aa8b5768bbb802177"
	clientSecret = "4ddd10a13f2a4c00af97c1916b21a8c2"
	redirectURI  = "http://localhost:8086/callback"
	scope        = "user-read-private user-read-email user-top-read playlist-modify-public playlist-modify-private"
	ARTISTNAME   = "King gnu"
	TRACKNAME    = "Supernova"
)

var (
	currentAccessToken  string
	currentRefreshToken string
	tokenExpiresAt      int64
	processedRequests   = make(map[string]bool)
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

type Track struct {
	Name       string
	ID         string
	URL        string
	ImageURL   string
	PreviewURL string
}

type Playlist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TemplateData struct {
	UserData     User
	SearchTrack  Track
	PlaylistData Playlist
}

var userdata User
var playlistdata Playlist

var trackURIs = []string{
	"spotify:track:24ntZeyCrVePmN3nUYhfLx",
	"spotify:track:1pCcNaCodPssCc8Aq68gPS",
	"spotify:track:7kJBYHytiARJlRygfg5VCn",
}

// 連接html
func getUserInfo(userInfo map[string]interface{}) {
	externalUrls := userInfo["external_urls"].(map[string]interface{})
	images := userInfo["images"].([]interface{})

	userdata = User{
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

	data := TemplateData{
		UserData:     userdata,
		SearchTrack:  Track{Name: TRACKNAME},
		PlaylistData: playlistdata,
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
	params.Set("scope", scope)

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

func startServer() {
	http.HandleFunc("/userinfo", handler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		authURL := generateAuthURL()
		http.Redirect(w, r, authURL, http.StatusSeeOther)
	})

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "未獲取到授權碼", http.StatusBadRequest)
			return
		}

		if processedRequests[code] {
			http.Error(w, "請求已處理", http.StatusBadRequest)
			return
		}
		processedRequests[code] = true
		defer func() { delete(processedRequests, code) }()

		token, err := exchangeCodeForToken(code)
		if err != nil {
			http.Error(w, "無法獲取 Token", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		_, err = getCurrentUserInfo(token.AccessToken)
		if err != nil {
			log.Println("取得UserInfo失敗: ", err)
			http.Error(w, "無法取得UserInfo", http.StatusInternalServerError)
			return
		}

		err = searchArtist(ARTISTNAME, token.AccessToken)
		if err != nil {
			log.Println("搜索歌手失敗:", err)
			http.Error(w, "無法搜索歌手", http.StatusInternalServerError)
			return
		}

		exists, playlistID, err := playlistExists(token.AccessToken, "我的收藏")
		if err != nil {
			log.Println("檢查播放清單失敗:", err)
			http.Error(w, "檢查播放清單失敗", http.StatusInternalServerError)
			return
		}

		if exists {
			playlistdata.ID = playlistID
			playlistdata.Name = "我的收藏"
		} else {
			playlistPointer, err := createPlaylist(userdata.UserID, "我的收藏", "From SpotGoInsight")
			if err != nil {
				log.Println("新增播放清單失敗: ", err)
				http.Error(w, "無法新增播放清單", http.StatusInternalServerError)
				return
			}
			playlistdata.ID = playlistPointer.ID
			playlistdata.Name = playlistPointer.Name
		}

		err = addTracksToPlaylist(playlistdata.ID, trackURIs)
		if err != nil {
			log.Println("新增歌曲到播放清單失敗: ", err)
			http.Error(w, "無法新增歌曲到播放清單", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/userinfo", http.StatusSeeOther)
	})

	fmt.Println("伺服器啟動於 http://localhost:8086")
	log.Fatal(http.ListenAndServe(":8086", nil))
}

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

	currentAccessToken = tokenResponse.AccessToken
	currentRefreshToken = tokenResponse.RefreshToken
	tokenExpiresAt = time.Now().Unix() + int64(tokenResponse.ExpiresIn)

	return &tokenResponse, nil
}

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

	externalUrls := userInfo["external_urls"].(map[string]interface{})
	images := userInfo["images"].([]interface{})

	userdata = User{
		Name:       userInfo["display_name"].(string),
		SpotifyURL: externalUrls["spotify"].(string),
		ImageURL:   images[0].(map[string]interface{})["url"].(string),
		UserID:     userInfo["id"].(string),
	}

	return userInfo, nil
}

func searchArtist(ARTISTNAME, accessToken string) error {
	baseSearchURL := "https://api.spotify.com/v1/search"
	params := url.Values{}
	params.Set("q", ARTISTNAME)
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
		return fmt.Errorf("API 請求失敗，狀態碼: %d", resp.StatusCode)
	}

	var searchResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&searchResult)
	if err != nil {
		return err
	}

	artists := searchResult["artists"].(map[string]interface{})
	items := artists["items"].([]interface{})
	if len(items) == 0 {
		return fmt.Errorf("未找到歌手")
	}

	artist := items[0].(map[string]interface{})
	artistID := artist["id"].(string)
	topTracks, err := getTopTracksForArtist(artistID, accessToken)
	if err != nil {
		return err
	}

	if len(topTracks) > 0 {
		track := topTracks[0]
		SearchTrack.Name = track.Name
		SearchTrack.ID = track.ID
		SearchTrack.URL = track.ExternalURLs["spotify"].(string)
		SearchTrack.ImageURL = track.ImageURL
		SearchTrack.PreviewURL = track.PreviewURL
	}

	return nil
}

func getTopTracksForArtist(artistID, accessToken string) ([]Track, error) {
	topTracksURL := fmt.Sprintf("https://api.spotify.com/v1/artists/%s/top-tracks?market=US", artistID)
	req, err := http.NewRequest("GET", topTracksURL, nil)
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

	var topTracksResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&topTracksResp)
	if err != nil {
		return nil, err
	}

	var topTracks []Track
	tracks := topTracksResp["tracks"].([]interface{})
	for _, trackInterface := range tracks {
		track := trackInterface.(map[string]interface{})
		topTracks = append(topTracks, Track{
			Name:       track["name"].(string),
			ID:         track["id"].(string),
			URL:        track["external_urls"].(map[string]interface{})["spotify"].(string),
			ImageURL:   track["album"].(map[string]interface{})["images"].([]interface{})[0].(map[string]interface{})["url"].(string),
			PreviewURL: track["preview_url"].(string),
		})
	}

	return topTracks, nil
}

func createPlaylist(userID, name, description string) (Playlist, error) {
	url := fmt.Sprintf("https://api.spotify.com/v1/users/%s/playlists", userID)

	playlist := map[string]interface{}{
		"name":        name,
		"description": description,
		"public":      false,
	}

	playlistData, err := json.Marshal(playlist)
	if err != nil {
		return Playlist{}, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(playlistData))
	if err != nil {
		return Playlist{}, err
	}

	req.Header.Set("Authorization", "Bearer "+currentAccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Playlist{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return Playlist{}, fmt.Errorf("創建播放清單失敗，狀態碼: %d", resp.StatusCode)
	}

	var playlistResponse Playlist
	err = json.NewDecoder(resp.Body).Decode(&playlistResponse)
	if err != nil {
		return Playlist{}, err
	}

	return playlistResponse, nil
}

func playlistExists(accessToken, playlistName string) (bool, string, error) {
	url := fmt.Sprintf("https://api.spotify.com/v1/me/playlists")
	req, err := http.NewRequest("GET", url, nil)
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
		return false, "", fmt.Errorf("API 請求失敗，狀態碼: %d", resp.StatusCode)
	}

	var playlistsResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&playlistsResponse)
	if err != nil {
		return false, "", err
	}

	items := playlistsResponse["items"].([]interface{})
	for _, itemInterface := range items {
		item := itemInterface.(map[string]interface{})
		if item["name"].(string) == playlistName {
			return true, item["id"].(string), nil
		}
	}

	return false, "", nil
}

func addTracksToPlaylist(playlistID string, trackURIs []string) error {
	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlistID)

	trackData := map[string]interface{}{
		"uris": trackURIs,
	}

	trackDataJSON, err := json.Marshal(trackData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(trackDataJSON))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+currentAccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("無法添加歌曲到播放清單，狀態碼: %d", resp.StatusCode)
	}

	return nil
}

func main() {
	startServer()
}
