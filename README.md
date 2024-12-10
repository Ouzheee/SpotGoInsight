# SpotGoInsight
GOOOOOO


### 11/19進度
#### 前端
  網頁分成兩個模式 signer跟albumn mode，可以輸入歌手或專輯的名字，之後連結後端及API功能後可以顯示歌手的資料、專輯歌曲等等

  也可以點擊右側星星將歌曲、歌手加入favorite，再點一次星星可以移出favorite

  問題：移出favorite時，favorite欄會確實刪除，但星星不會變回空心的
  

#### 後端
  了解怎麼抓取token以及用token來access其他API並抓取資料的方式了
  ![image](https://github.com/user-attachments/assets/2bb47606-2467-4143-86b6-8dbdaf002f2c)
  
  最簡單的架構就是
  
  ┌-server.go架網站
  
  ├-autho.go獲得token
  
  ├-app.js 用token取得歌手的ID,專輯的ID等等
  
  └-index.html 把js的結果畫在網頁上
  
  目標：
    1.完成autho.go
    2.用code及token抓到ID
    3.看能不能用go lang 取代 js，以Gin等方式去實現js做到的功能

---
### 11/20進度
#### 前端
  前端畫面已經完成90%左右，我的收藏、播放、切換上下首都已完成
  可以播放音樂和顯示圖片了，現在需要等待輸入歌曲後後端能夠回傳音樂和圖片的url
  預計新增一項我的收藏，能夠在那播放所有最愛歌曲
  
  介面美化

#### 後端 
  可以在運行程式後連結到授權頁面並獲取token及使用者資訊。且可以透過帶入的歌手名字取得歌手ID及top tracks了
  
  在其他電腦運行程式時會無法獲取token，需解決不同電腦要重新申請client ID的問題

---
### 11/21進度

#### 前端
   大致上弄完了所以今天來幫忙弄後端
#### 後端
   client ID的問題沒有解決，切換到不同電腦、登入不同帳號就會找不到token
   找裡來說，所有使用者都可以使用同一個開發者提供的client ID去申請token，但目前看下來是不行的
   該問題亟待解決
   新增了preview URL、draw the toptracks on the html檔
   
---
### 11/22進度

#### 前端
登入、輸入client ID的網頁模塊好了，但是輸入的資訊(client Id等等沒辦法回傳，按下button之後資料就不見了，待解決

#### 後端

可以擷取專輯的ID、Name、藝人；歌曲屬於的專輯、唱的歌手等等。也可以在登入的Spotif帳號建立想要的歌曲組成的播放清單了。\n
Client Id 決定先讓使用者自己想辦法生出來之後書來，再使用我們的網頁。未來有空再解決。
問題：不知道為什麼，hadler("/callback")會執行兩次，在terminal端出都會跑2次結果。待解決

---
### 11/26進度

開始合併前後端了，成功合併到獲得歌手的資料的部分。
前端9成完成了，使用者資料弄完了
後端的URL還沒連結完，但已經串街完歌手資料了，歌曲後面的資料再加油

---
### 11/27進度

完成了前後端的串接，搜尋歌曲的網頁可以顯示歌曲的名字、圖片、連結
![image](https://github.com/user-attachments/assets/37fc1c44-9d2d-477c-a9f9-b29ea4439beb)

但是當我們測試試聽功能時，發現原本可以正常使用的試聽連結無法使用了，當我們重新翻閱API的使用頁面時，才發現spotify更新一部份的web API功能，30秒視聽的API被移除了
原本這是對我們來說很重要的功能，花了很多時間，結果做到一半被刪除了，下次討論想辦法解決
![image](https://github.com/user-attachments/assets/03ce3ba3-599f-44c7-af49-19401f10296f)

---

### 11/28進度

在發現試聽功能背移除後，我們分頭往兩個方向發展，試著解決試聽API被移除，無法在網頁播放音訊的問題
方案一、使用Spotify的嵌入功能

Spotify的嵌入功能就如同一個
方案二、使用網路爬蟲，搜索youtube的影片並將音訊嵌入到網頁中


我們實現的爬蟲可以在youtube搜尋輸入的文字，並回傳搜尋頁面的第一個結果的連結到網頁中

最後我們採用了方案一，並將嵌入畫面和原本的搜尋歌曲、搜尋歌手結合
![image](https://github.com/user-attachments/assets/60b2ddbb-d8ab-4205-9e7e-9225ad12ffcf)![image](https://github.com/user-attachments/assets/02cfda69-fad8-4b74-9b13-3adb32d84880)
但是個人spotify資訊的顯示還有問題，我的最愛、新增到播放清單因為改使用嵌入功能出現了問題，下次解決

另外，搜尋歌曲有時會出現和預期相差很多的歌曲，刪除歌曲也會從最左邊(歌曲陣列的第0個)開始刪除，待解決

---
### 12/3進度

#### 新增功能
* 新增匯出播放清單功能，可以在加入我的收藏後，前往專輯頁面，點擊畫面中央的匯出按鈕就會把收藏的匯出致使用者的spotify帳號![image](https://github.com/user-attachments/assets/bc0bf171-2999-4400-aeb4-66f9bbe452ea)
並且新增的播放清單也會嵌入到畫面上
![image](https://github.com/user-attachments/assets/c0e6464e-1123-4bc9-8f59-9a23bb6452a3)

* 使用者資料修復完成
  原本無法顯示照片、資料的問題解決，原因是user變數的設定錯誤跟重複混用，沒有傳送到HTML上
* 播放清單可以在框框中命名，原本只能叫 我的收藏 變得可以自由命名
* 些許介面美化，應該是最終版本，不會再更改了
#### 遇到的問題
* 版本合併過程中，對於VS code 合併功能不熟悉，不小心就會合併錯誤，要重新修正code
* Access token 常常在不同帳號間使用就爆錯，連輸入client ID 的畫面都跑不了
 或者新增下列程式後，原本可以運行的程式又突然抓不到token了
* 把原本加到收藏的歌曲刪除後，新建的播放清單還是會有那首歌，試著用一樣的方法刪除TrackURIs中的指定歌曲，但會變成無法新建播放清單，待解決
```
// Remove Singer (Delete Singer)
	var this_song string
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
				this_song = song.Name
				songList = append(songList[:i], songList[i+1:]...)
				break
			}
		}

		for i, song := range favoriteSongs {
			if song.Name == this_song {
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
```

#### 剩下的問題、想加的東西
* Access token
      除了陳庭毅的帳號電腦外，大家都可以跑了，陳庭毅的待解決
* 輸入框有點問題
* 刪除我的加到我的最愛的歌曲，可以從搜尋頁面及我的最愛中刪除，但沒辦法刪除播放清單中的歌

