package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Album struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Tracks     int    `json:"tracks"`
	Year       int    `json:"year"`
	IsFavorite bool   `json:"is_favorite"`
	AudioURL   string `json:"audio_url"`
	ImageURL   string `json:"image_url"`
}

type Track struct {
	Name       string `json:"name"`
	ID         string `json:"id"`
	SongURL    string `json:"song_url`
	ImageURL   string `json:"image_url"`
	PreviewURL string `json:"preview_url"`
	Album      Album  `json:"album"`
}

type User struct {
	ID           int    `json:"id"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	UserName     string `json:"user_name"`
}

type Singer struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Genre      string `json:"genre"`
	IsFavorite bool   `json:"is_favorite"`
	AudioURL   string `json:"audio_url"`
	ImageURL   string `json:"image_url"`
}

var maxAlbumID = 1
var maxSingerID = 1

var albumList = []Album{
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

var singerList = []Singer{
	{
		ID:         1,
		Name:       "Dua Lipa",
		Genre:      "Pop",
		IsFavorite: false,
		AudioURL:   "https://p.scdn.co/mp3-preview/82e442871e6afd7efa4410ca735b3b13644f5184",
		ImageURL:   "https://i.scdn.co/image/ab67616d00001e02ff9ca10b55ce82ae553c8228",
	},
}

var users = []User{
	{
		ID:           1,
		ClientID:     "jason",
		ClientSecret: "it's a secret",
		UserName:     "jasonnning",
	},
}

var maxUserID = 0

var ex_user = User{
	ID:           1,
	ClientID:     "592fa46f290e4f1aa8b5768bbb802177",
	ClientSecret: "4ddd10a13f2a4c00af97c1916b21a8c2",
	UserName:     "jasonnning",
}

var favoriteAlbums []Album
var favoriteSingers []Singer

func getCurrentUser(c *gin.Context) *User {
	// 從 cookie 中獲取 user_id
	userID, err := c.Cookie("id")
	if err != nil {
		return &ex_user // 如果 cookie 不存在，表示未登入
	}

	// 將 userID 轉換為整數
	id, _ := strconv.Atoi(userID)

	// 根據 userID 查找用戶
	for _, user := range users {
		if user.ID == id {
			return &user // 返回對應的用戶
		}
	}
	return &ex_user
}

func main() {
	r := gin.Default()

	// 加載模板文件
	r.LoadHTMLGlob("templates/*")

	// Main Menu page
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "menu.html", nil)
	})

	// Album Mode
	r.GET("/album", func(c *gin.Context) {
		c.HTML(http.StatusOK, "album.html", gin.H{
			"albumlist": albumList,
			"favorites": favoriteAlbums,
		})
	})

	// Singer Mode
	r.GET("/singer", func(c *gin.Context) {
		c.HTML(http.StatusOK, "singer.html", gin.H{
			"singers":   singerList,
			"favorites": favoriteSingers,
		})
	})

	// Add Album
	r.POST("/add/album", func(c *gin.Context) {
		name := c.PostForm("name")
		tracks, err := strconv.Atoi(c.PostForm("tracks"))
		year, err2 := strconv.Atoi(c.PostForm("year"))
		if err != nil || err2 != nil || name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
			return
		}

		maxAlbumID++
		newAlbum := Album{ID: maxAlbumID, Name: name, Tracks: tracks, Year: year}
		albumList = append(albumList, newAlbum)

		c.Redirect(http.StatusFound, "/album")
	})

	// Add Singer
	r.POST("/add/singer", func(c *gin.Context) {
		name := c.PostForm("name")
		genre := c.PostForm("genre")
		if name == "" || genre == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
			return
		}

	 	maxSingerID++
		newSinger := Singer{ID: maxSingerID, Name: name, Genre: genre}
		singerList = append(singerList, newSinger)

		c.Redirect(http.StatusFound, "/singer")
	})

	// Add Album to Favorite
	r.GET("/favorite/album/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID format"})
			return
		}

		for i, album := range albumList {
			if album.ID == id {
				albumList[i].IsFavorite = true // 加入最愛
				favoriteAlbums = append(favoriteAlbums, albumList[i])
				break
			}
		}
	})

	// Remove Album from Favorite
	r.GET("/favorite/album/remove/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID format"})
			return
		}

		for i, album := range favoriteAlbums {
			if album.ID == id {
				// 從最愛中移除
				favoriteAlbums = append(favoriteAlbums[:i], favoriteAlbums[i+1:]...)
				for j, albumInList := range albumList {
					if albumInList.ID == id {
						albumList[j].IsFavorite = false
						break
					}
				}
				break
			}
		}
		c.Redirect(http.StatusFound, "/album")
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

	// Remove Album (Delete Album)
	r.GET("/album/delete/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID format"})
			return
		}

		for i, album := range albumList {
			if album.ID == id {
				// 從專輯列表中刪除
				albumList = append(albumList[:i], albumList[i+1:]...)
				break
			}
		}
		c.Redirect(http.StatusFound, "/album")
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

	r.GET("/favorite", func(c *gin.Context) {
		c.HTML(http.StatusOK, "favorite.html", gin.H{
			"favorites": favoriteAlbums,
		})
	})

	/*r.GET("/user", func(c *gin.Context) {
		c.HTML(http.StatusOK, "user.html", nil)
	})*/
	r.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		clientID := c.PostForm("client_id")
		clientSecret := c.PostForm("client_secret")

		// 是否是註冊新用戶
		isRegister := clientID != "" && clientSecret != ""

		if isRegister {
			// 檢查用戶是否已存在
			for _, user := range users {
				if user.UserName == username {
					c.JSON(http.StatusConflict, gin.H{"message": "User already exists"})
					return
				}
			}
			// 註冊新用戶
			maxUserID++
			newUser := User{
				ID:           maxUserID,
				ClientID:     clientID,
				ClientSecret: clientSecret,
				UserName:     username,
			}
			users = append(users, newUser)

			// 設置當前用戶在 context 中
			c.Set("user_id", newUser.ID)

			// 重定向到歌手頁面
			c.Redirect(http.StatusFound, "/singer")
			return
		}

		// 登入驗證
		for _, user := range users {
			if user.UserName == username {
				// 設置當前用戶在 context 中
				c.Set("user_id", user.ID)

				// 重定向到歌手頁面
				c.Redirect(http.StatusFound, "/singer")
				return
			}
		}

		// 用戶不存在
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found"})
	})

	r.POST("/register", func(c *gin.Context) {
		clientID := c.PostForm("client_id")
		clientSecret := c.PostForm("client_secret")
		userName := c.PostForm("username")

		// 驗證輸入是否有效
		if clientID == "" || clientSecret == "" || userName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "All fields are required"})
			return
		}

		// 新增到使用者清單
		maxUserID++
		newUser := User{
			ID:           maxUserID,
			ClientID:     clientID,
			ClientSecret: clientSecret,
			UserName:     userName,
		}
		users = append(users, newUser)

		c.Redirect(http.StatusFound, "/user")

	})

	// Register Page
	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})

	r.GET("/user", func(c *gin.Context) {
		// 確認用戶是否已經登入
		currentUser := getCurrentUser(c)

		if currentUser == nil {
			// 如果沒有登入，則提示或重定向到登入頁面
			c.JSON(http.StatusUnauthorized, gin.H{"message": "No user logged in"})
			return
		}

		// 如果用戶已登錄，顯示 user.html 並傳遞用戶資料
		c.HTML(http.StatusOK, "user.html", gin.H{
			"user_id":       currentUser.ID,
			"user_name":     currentUser.UserName,
			"client_id":     currentUser.ClientID,
			"client_secret": currentUser.ClientSecret,
		})
	})

	// Start the server
	r.Run(":8080")
}
