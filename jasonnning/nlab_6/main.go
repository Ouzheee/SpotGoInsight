package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Song2 struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Tracks     int    `json:"tracks"`
	Year       int    `json:"year"`
	IsFavorite bool   `json:"is_favorite"`
	AudioURL   string `json:"audio_url"`
	ImageURL   string `json:"image_url"`
}

type User2 struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type Singer2 struct {
	ID         int    `json:"id"`
	SingerID   string `json:"singerid"`
	Name       string `json:"name"`
	Genre      string `json:"genre"`
	IsFavorite bool   `json:"is_favorite"`
	AudioURL   string `json:"audio_url"`
	ImageURL   string `json:"image_url"`
}

var maxSongID = 1
var maxSingerID = 1

var songList = []Song2{
	{
		ID:         1,
		Name:       "Future Nostalgia",
		Tracks:     11,
		Year:       2020,
		IsFavorite: false,
		AudioURL:   "https://p.scdn.co/mp3-preview/82e442871e6afd7efa4410ca735b3b13644f5184",
		ImageURL:   "https://i.scdn.co/image/ab67616d00001e02ff9ca10b55ce82ae553c8228",
	},

	{
		ID:         2,
		Name:       "陳庭毅真的很強",
		Tracks:     11,
		Year:       2020,
		IsFavorite: false,
		AudioURL:   "https://p.scdn.co/mp3-preview/104ad0ea32356b9f3b2e95a8610f504c90b0026b?cid=8897482848704f2a8f8d7c79726a70d4",
		ImageURL:   "https://i.scdn.co/image/ab67616d00001e02ff9ca10b55ce82ae553c8228",
	},
}

var singerList = []Singer2{
	{
		ID:         1,
		Name:       "Dua Lipa",
		Genre:      "Pop",
		IsFavorite: false,
		AudioURL:   "https://p.scdn.co/mp3-preview/82e442871e6afd7efa4410ca735b3b13644f5184",
		ImageURL:   "https://i.scdn.co/image/ab67616d00001e02ff9ca10b55ce82ae553c8228",
	},
}

var users = []User2{}

var (
	currentUser *User2 // 用於存儲當前登入的用戶
)

var favoriteSongs []Song2
var favoriteSingers []Singer2

func main() {
	r := gin.Default()

	// 加載模板文件
	r.LoadHTMLGlob("templates/*")

	fmt.Println("伺服器啟動於 http://localhost:8080")
	fmt.Println("cliend id: 592fa46f290e4f1aa8b5768bbb802177")
	fmt.Println("cliend secret: 4ddd10a13f2a4c00af97c1916b21a8c2")

	/*================================進入網頁URL指令==================================*/
	// Main Menu page
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "menu.html", nil)
	})

	// Song Mode
	r.GET("/song", func(c *gin.Context) {
		c.HTML(http.StatusOK, "song.html", gin.H{
			"songlist":  songList,
			"favorites": favoriteSongs,
		})
	})

	// Singer Mode
	r.GET("/singer", func(c *gin.Context) {
		c.HTML(http.StatusOK, "singer.html", gin.H{
			"singers":   singerList,
			"favorites": favoriteSingers,
		})
	})

	// 用戶頁面
	r.GET("/user", func(c *gin.Context) {
		if currentUser == nil {
			// 如果沒有登入，重定向到主頁或登入頁
			c.Redirect(http.StatusFound, "/")
			return
		}

		// 傳遞 currentUser 的資料給模板
		c.HTML(http.StatusOK, "user.html", gin.H{
			"user": map[string]string{
				"ClientID":     currentUser.ClientID,
				"ClientSecret": currentUser.ClientSecret,
			},
		})
	})

	r.GET("/favorite", func(c *gin.Context) {
		c.HTML(http.StatusOK, "favorite.html", gin.H{
			"favorites": favoriteSongs,
		})
	})

	r.GET("/logout", func(c *gin.Context) {
		// 清空 currentUser
		currentUser = nil
		c.Redirect(http.StatusFound, "/")
	})

	/*==================================================================================*/

	/*================================URL指令執行動作==================================*/
	// Add Song

	r.POST("/add/song", func(c *gin.Context) {
		name := c.PostForm("name")
		tracks, err := strconv.Atoi(c.PostForm("tracks"))
		year, err2 := strconv.Atoi(c.PostForm("year"))
		if err != nil || err2 != nil || name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
			return
		}
		/*




		 */
		maxSongID++
		newSong := Song2{ID: maxSongID, Name: name, Tracks: tracks, Year: year}
		songList = append(songList, newSong)

		c.Redirect(http.StatusFound, "/song")
	})

	// Add Singer
	r.POST("/add/singer", func(c *gin.Context) {
		name := c.PostForm("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
			return
		}
		fmt.Println("enter name: ", name)

		ARTISTNAME = name
		fmt.Println("token======", token.AccessToken)
		err := searchArtist(ARTISTNAME, token.AccessToken)
		if err != nil {
			log.Println("搜索歌手失敗:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "無法搜索歌手"})
			return
		}

		maxSingerID++
		newSinger := Singer2{
			SingerID: singerdata.SingerID,
			Name:     singerdata.Name,
		}
		fmt.Println("singerID: ", singerdata.SingerID)
		singerList = append(singerList, newSinger)

		c.Redirect(http.StatusFound, "/singer")
	})

	// Add Song to Favorite
	r.GET("/favorite/song/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID format"})
			return
		}

		for i, song := range songList {
			if song.ID == id {
				songList[i].IsFavorite = true // 加入最愛
				favoriteSongs = append(favoriteSongs, songList[i])
				break
			}
		}
		c.Redirect(http.StatusFound, "/song")
	})

	// Remove Song from Favorite
	r.GET("/favorite/song/remove/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID format"})
			return
		}

		for i, song := range favoriteSongs {
			if song.ID == id {
				// 從最愛中移除
				favoriteSongs = append(favoriteSongs[:i], favoriteSongs[i+1:]...)
				for j, songInList := range songList {
					if songInList.ID == id {
						songList[j].IsFavorite = false
						break
					}
				}
				break
			}
		}
		c.Redirect(http.StatusFound, "/song")
	})

	// Add Singer to Favorite
	r.GET("/favorite/singer/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID format"})
			return
		}

		for i, singer := range singerList {
			if singer.ID == id {
				singerList[i].IsFavorite = true
				favoriteSingers = append(favoriteSingers, singerList[i])
				break
			}
		}
		c.Redirect(http.StatusFound, "/singer")
	})

	// Remove Singer from Favorite
	r.GET("/favorite/singer/remove/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID format"})
			return
		}

		for i, singer := range favoriteSingers {
			if singer.ID == id {
				favoriteSingers = append(favoriteSingers[:i], favoriteSingers[i+1:]...)
				for j, singerInList := range singerList {
					if singerInList.ID == id {
						singerList[j].IsFavorite = false
						break
					}
				}
				break
			}
		}
		c.Redirect(http.StatusFound, "/singer")
	})

	// Remove Song (Delete Song)
	r.GET("/song/delete/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID format"})
			return
		}

		for i, song := range songList {
			if song.ID == id {
				// 從專輯列表中刪除
				songList = append(songList[:i], songList[i+1:]...)
				break
			}
		}
		c.Redirect(http.StatusFound, "/song")
	})

	// Remove Singer (Delete Singer)
	r.GET("/singer/delete/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID format"})
			return
		}

		for i, singer := range singerList {
			if singer.ID == id {
				// 從歌手列表中刪除
				singerList = append(singerList[:i], singerList[i+1:]...)
				break
			}
		}
		c.Redirect(http.StatusFound, "/singer")
	})

	//login輸入操作
	r.POST("/login", func(c *gin.Context) {
		clientID = c.PostForm("client_id")
		clientSecret = c.PostForm("client_secret")
		
		// 直接創建一個新的 User
		newUser := User2{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		}

		// 設定 currentUser 為新創建的用戶
		currentUser = &newUser

		//授權畫面
		authURL := generateAuthURL()
		c.Redirect(http.StatusFound, authURL)
	})

	r.GET("/callback", func(c *gin.Context) {
		// 獲取授權碼
		code := c.Query("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "未獲取到授權碼"})
			return
		}

		// 防止重複處理請求
		if processedRequests[code] {
			c.JSON(http.StatusBadRequest, gin.H{"message": "請求已處理"})
			return
		}
		processedRequests[code] = true
		defer func() { delete(processedRequests, code) }()

		// 使用授權碼交換 Token
		tokentmp, err := exchangeCodeForToken(code)
		if err != nil {
			log.Println("無法獲取 Token:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "無法獲取 Token"})
			return
		}
		token = tokentmp

		// 返回成功消息
		/*c.JSON(http.StatusOK, gin.H{
			"message":       "成功獲取授權碼與 Token",
			"access_token":  token.AccessToken,
			"refresh_token": token.RefreshToken,
		})*/

		// 登入成功，跳轉到用戶頁面
		c.Redirect(http.StatusFound, "/user")
	})

	/*=========================================================================*/

	// Start the server
	r.Run(":8080")

}
