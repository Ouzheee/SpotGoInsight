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
}

type Singer struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Genre      string `json:"genre"`
	IsFavorite bool   `json:"is_favorite"`
}

var maxAlbumID = 1
var maxSingerID = 1

var albumList = []Album{
	{ID: 1, Name: "Future Nostalgia", Tracks: 11, Year: 2020, IsFavorite: false},
}

var singerList = []Singer{
	{ID: 1, Name: "Dua Lipa", Genre: "Pop", IsFavorite: false},
}

var favoriteAlbums []Album
var favoriteSingers []Singer

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
		c.Redirect(http.StatusFound, "/album")
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

	// Start the server
	r.Run(":8080")
}
