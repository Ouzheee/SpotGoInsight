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
