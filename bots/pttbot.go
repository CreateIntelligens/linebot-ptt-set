package bots

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/mong0520/linebot-ptt-set/controllers"
	"github.com/mong0520/linebot-ptt-set/models"
	"github.com/mong0520/linebot-ptt-set/utils"
	"go.mongodb.org/mongo-driver/bson"
)

var bot *linebot.Client
var meta *models.Model
var maxCountOfCarousel = 10
var defaultImage = "https://i.imgur.com/WAnWk7K.png"
var defaultThumbnail = "https://i.imgur.com/StcRAPB.png"
var oneDayInSec = 60 * 60 * 24
var oneWeekInSec = oneDayInSec * 7
var oneMonthInSec = oneDayInSec * 30
var oneYearInSec = oneMonthInSec * 365

const (
	DefaultTitle string = "🔥 SET 看看"

	ActionQuery       string = "一般查詢"
	ActionNewest      string = "🎊 最新文章"
	ActionDailyHot    string = "📈 本日熱門"
	ActionMonthlyHot  string = "🔥 近期熱門"
	ActionYearHot     string = "🏆 年度熱門"
	ActionRandom      string = "🎲 隨機十連抽"
	ActionAddFavorite string = "加入最愛"
	ActionClick       string = "👉 點我打開"
	ActionHelp        string = "SET 選單"
	ActionAllImage    string = "👁️ 預覽圖片"
	ActonShowFav      string = "❤️ 我的最愛"
	ActonRunCC        string = "/cc"
	ModeHTTP          string = "http"
	ModeHTTPS         string = "https"
	AltText           string = "內容只在手機上"
)

func InitLineBot(m *models.Model, runMode string, sslCertPath string, sslPKeyPath string) {

	var err error
	meta = m
	secret := os.Getenv("ChannelSecret")
	token := os.Getenv("ChannelAccessToken")
	bot, err = linebot.New(secret, token)
	if err != nil {
		log.Println(err)
	}
	//log.Println("Bot:", bot, " err:", err)
	http.HandleFunc("/callback", callbackHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/stats", statsHandler)
	http.HandleFunc("/api/articles", articlesHandler)
	http.HandleFunc("/", dashboardHandler)
	port := os.Getenv("PORT")
	//port := "8080"
	addr := fmt.Sprintf(":%s", port)
	m.Log.Printf("Run Mode = %s\n", runMode)
	if strings.ToLower(runMode) == ModeHTTPS {
		m.Log.Printf("Secure listen on %s with \n", addr)
		err := http.ListenAndServeTLS(addr, sslCertPath, sslPKeyPath, nil)
		if err != nil {
			m.Log.Panic(err)
		}
	} else {
		m.Log.Printf("Listen on %s\n", addr)
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			m.Log.Panic(err)
		}
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(200)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	// 只處理根路徑，避免攔截API路由
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	html := `
<!DOCTYPE html>
<html lang="zh-TW">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PTT SET Bot 數據面板</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; border-radius: 10px; margin-bottom: 30px; text-align: center; }
        .header h1 { font-size: 2.5em; margin-bottom: 10px; }
        .header p { font-size: 1.2em; opacity: 0.9; }
        .dashboard { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .card { background: white; border-radius: 10px; padding: 20px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); transition: transform 0.2s; }
        .card:hover { transform: translateY(-2px); }
        .card h3 { color: #333; margin-bottom: 15px; font-size: 1.3em; border-bottom: 2px solid #eee; padding-bottom: 10px; }
        .stat { display: flex; justify-content: space-between; align-items: center; margin: 10px 0; padding: 10px; background: #f8f9fa; border-radius: 5px; }
        .stat-label { font-weight: 500; color: #666; }
        .stat-value { font-weight: bold; color: #333; font-size: 1.2em; }
        .loading { text-align: center; padding: 20px; color: #666; }
        .btn { padding: 10px 20px; border: none; border-radius: 5px; cursor: pointer; font-size: 14px; transition: all 0.2s; background: #007bff; color: white; }
        .btn:hover { opacity: 0.8; transform: translateY(-1px); }
        .table { width: 100%; border-collapse: collapse; margin-top: 15px; }
        .table th, .table td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        .table th { background: #f8f9fa; font-weight: 600; }
        .table tr:hover { background: #f5f5f5; }
        .article-title { max-width: 300px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .push-count { color: #28a745; font-weight: bold; }
        .boo-count { color: #dc3545; font-weight: bold; }
        .status-indicator { display: inline-block; width: 12px; height: 12px; border-radius: 50%; margin-right: 8px; background: #28a745; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔥 PTT SET Bot 數據面板</h1>
            <p>監控爬蟲狀態 • 查看數據統計 • 系統管理</p>
        </div>

        <div class="dashboard">
            <!-- 系統狀態 -->
            <div class="card">
                <h3>⚙️ 系統狀態</h3>
                <div class="stat">
                    <span class="stat-label">
                        <span class="status-indicator"></span>
                        MongoDB連接
                    </span>
                    <span class="stat-value">正常</span>
                </div>
                <div class="stat">
                    <span class="stat-label">應用程式狀態</span>
                    <span class="stat-value">運行中</span>
                </div>
                <div class="stat">
                    <span class="stat-label">端口</span>
                    <span class="stat-value">5050</span>
                </div>
            </div>

            <!-- 數據庫統計 -->
            <div class="card">
                <h3>📊 數據庫統計</h3>
                <div id="db-stats" class="loading">載入中...</div>
                <button class="btn" onclick="refreshStats()">刷新統計</button>
            </div>

            <!-- 快速操作 -->
            <div class="card">
                <h3>🚀 快速操作</h3>
                <div class="stat">
                    <span class="stat-label">健康檢查</span>
                    <button class="btn" onclick="checkHealth()">檢查</button>
                </div>
                <div class="stat">
                    <span class="stat-label">查看日誌</span>
                    <button class="btn" onclick="viewLogs()">查看</button>
                </div>
                <div class="stat">
                    <span class="stat-label">爬蟲狀態</span>
                    <button class="btn" onclick="checkCrawler()">檢查</button>
                </div>
            </div>
        </div>

        <!-- 最新文章 -->
        <div class="card">
            <h3>📰 最新文章</h3>
            <div id="articles-list" class="loading">載入中...</div>
            <button class="btn" onclick="refreshArticles()">刷新列表</button>
        </div>
    </div>

    <script>
        // 載入統計數據
        function loadStats() {
            fetch('/api/stats')
                .then(response => response.json())
                .then(data => {
                    updateStats(data);
                })
                .catch(error => {
                    console.error('Error loading stats:', error);
                    document.getElementById('db-stats').innerHTML = '<div class="stat"><span class="stat-label">錯誤</span><span class="stat-value">無法載入數據</span></div>';
                });
        }

        // 更新統計顯示
        function updateStats(stats) {
            const statsHtml = ` + "`" + `
                <div class="stat">
                    <span class="stat-label">總文章數</span>
                    <span class="stat-value">${stats.total_articles || 0}</span>
                </div>
                <div class="stat">
                    <span class="stat-label">今日新增</span>
                    <span class="stat-value">${stats.today_articles || 0}</span>
                </div>
                <div class="stat">
                    <span class="stat-label">本週新增</span>
                    <span class="stat-value">${stats.week_articles || 0}</span>
                </div>
                <div class="stat">
                    <span class="stat-label">最後更新</span>
                    <span class="stat-value">${formatTime(stats.last_update)}</span>
                </div>
            ` + "`" + `;
            document.getElementById('db-stats').innerHTML = statsHtml;
        }

        // 載入文章列表
        function loadArticles() {
            fetch('/api/articles?limit=15')
                .then(response => response.json())
                .then(data => {
                    updateArticlesList(data);
                })
                .catch(error => {
                    console.error('Error loading articles:', error);
                    document.getElementById('articles-list').innerHTML = '<div class="loading">無法載入文章數據</div>';
                });
        }

        // 更新文章列表顯示
        function updateArticlesList(articles) {
            if (!articles || articles.length === 0) {
                document.getElementById('articles-list').innerHTML = '<div class="loading">暫無文章數據</div>';
                return;
            }

            let html = '<table class="table"><thead><tr><th>標題</th><th>作者</th><th>日期</th><th>推/噓</th><th>圖片</th><th>連結</th></tr></thead><tbody>';
            
            articles.forEach(article => {
                const title = article.article_title || '無標題';
                const author = article.author || '未知';
                const date = article.date || '未知';
                const pushCount = article.message_count ? article.message_count.push || 0 : 0;
                const booCount = article.message_count ? article.message_count.boo || 0 : 0;
                const imageCount = article.image_links ? article.image_links.length : 0;
                const url = article.url || '#';
                
                html += ` + "`" + `<tr>
                    <td class="article-title" title="${title}">${title.length > 40 ? title.substring(0, 40) + '...' : title}</td>
                    <td>${author}</td>
                    <td>${date}</td>
                    <td><span class="push-count">${pushCount}</span>/<span class="boo-count">${booCount}</span></td>
                    <td>${imageCount}</td>
                    <td><a href="${url}" target="_blank" class="btn" style="padding: 5px 10px; font-size: 12px;">查看</a></td>
                </tr>` + "`" + `;
            });
            
            html += '</tbody></table>';
            document.getElementById('articles-list').innerHTML = html;
        }

        // 刷新統計
        function refreshStats() {
            document.getElementById('db-stats').innerHTML = '<div class="loading">載入中...</div>';
            loadStats();
        }

        // 刷新文章
        function refreshArticles() {
            document.getElementById('articles-list').innerHTML = '<div class="loading">載入中...</div>';
            loadArticles();
        }

        // 健康檢查
        function checkHealth() {
            fetch('/health')
                .then(response => {
                    if (response.ok) {
                        alert('✅ 系統健康狀態正常');
                    } else {
                        alert('❌ 系統健康檢查失敗');
                    }
                })
                .catch(error => {
                    alert('❌ 無法連接到系統');
                });
        }

        // 查看日誌
        function viewLogs() {
            alert('📝 日誌位置:\n- 文件: ./logs/pttset.log\n- Docker: docker-compose logs app');
        }

        // 檢查爬蟲
        function checkCrawler() {
            alert('🕷️ 爬蟲狀態:\n- 容器: docker-compose ps crawler\n- 日誌: docker-compose logs crawler');
        }

        // 格式化時間
        function formatTime(timeStr) {
            if (!timeStr) return '未知';
            const date = new Date(timeStr);
            return date.toLocaleString('zh-TW');
        }

        // 頁面載入時執行
        document.addEventListener('DOMContentLoaded', function() {
            loadStats();
            loadArticles();
            // 每30秒自動刷新
            setInterval(loadStats, 30000);
        });
    </script>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// 獲取數據庫統計
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	stats := map[string]interface{}{
		"total_articles": 0,
		"today_articles": 0,
		"week_articles":  0,
		"last_update":    time.Now().Format(time.RFC3339),
	}
	
	// 嘗試獲取實際統計數據
	if meta.Collection != nil {
		if count, err := meta.Collection.CountDocuments(ctx, bson.M{}); err == nil {
			stats["total_articles"] = count
		}
		
		// 今日文章數
		today := time.Now().Truncate(24 * time.Hour)
		todayTimestamp := int(today.Unix())
		if count, err := meta.Collection.CountDocuments(ctx, bson.M{"timestamp": bson.M{"$gte": todayTimestamp}}); err == nil {
			stats["today_articles"] = count
		}
		
		// 本週文章數
		weekAgo := time.Now().AddDate(0, 0, -7)
		weekTimestamp := int(weekAgo.Unix())
		if count, err := meta.Collection.CountDocuments(ctx, bson.M{"timestamp": bson.M{"$gte": weekTimestamp}}); err == nil {
			stats["week_articles"] = count
		}
	}
	
	json.NewEncoder(w).Encode(stats)
}

func articlesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// 獲取查詢參數
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	// 獲取最新文章
	var articles []models.ArticleDocument
	if meta.Collection != nil {
		articles, _ = controllers.Get(meta.Collection, 0, limit)
	}
	
	json.NewEncoder(w).Encode(articles)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	meta.Log.Println("enter callback hander")
	events, err := bot.ParseRequest(r)

	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			userDisplayName := getUserNameById(event.Source.UserID)
			meta.Log.Printf("Receieve Event Type = %s from User [%s](%s), or Room [%s] or Group [%s]\n",
				event.Type, userDisplayName, event.Source.UserID, event.Source.RoomID, event.Source.GroupID)

			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				meta.Log.Println("Text = ", message.Text)
				textHander(event, message.Text)
			default:
				meta.Log.Println("Unimplemented handler for event type ", event.Type)
			}
		} else if event.Type == linebot.EventTypePostback {
			meta.Log.Println("got a postback event")
			meta.Log.Println(event.Postback.Data)
			postbackHandler(event)

		} else {
			meta.Log.Printf("got a %s event\n", event.Type)
		}
	}
}

func actionHandler(event *linebot.Event, action string, values url.Values) {
	switch action {
	case ActionNewest:
		actionNewest(event, values)
	case ActionAllImage:
		actionAllImage(event, values)
	case ActionQuery, ActionRandom:
		actionGeneral(event, action, values)
	case ActionAddFavorite:
		actinoAddFavorite(event, action, values)
	case ActonShowFav:
		actionShowFavorite(event, action, values)
	default:
		meta.Log.Println("Unimplement action handler", action)
	}
}

func actinoAddFavorite(event *linebot.Event, action string, values url.Values) {
	toggleMessage := ""
	userId := values.Get("user_id")
	newFavoriteArticle := values.Get("article_id")
	userFavorite := &controllers.UserFavorite{
		UserId:    userId,
		Favorites: []string{newFavoriteArticle},
	}
	latestFavArticles := []string{}
	if record, err := userFavorite.Get(meta); err != nil {
		meta.Log.Println("User data is not created, create a new one")
		userFavorite.Add(meta)
		latestFavArticles = append(latestFavArticles, newFavoriteArticle)
	} else {
		meta.Log.Println("Record found, update it", record)
		oldRecords := record.Favorites
		if exist, idx := utils.InArray(newFavoriteArticle, oldRecords); exist == true {
			meta.Log.Println(newFavoriteArticle, "已存在，移除")
			oldRecords = utils.RemoveStringItem(oldRecords, idx)
			toggleMessage = "已從最愛中移除"
		} else {
			oldRecords = append(oldRecords, newFavoriteArticle)
			toggleMessage = "已新增至最愛"
		}
		latestFavArticles = oldRecords
		userFavorite.Favorites = oldRecords
		userFavorite.Update(meta)
	}
	sendTextMessage(event, toggleMessage)
}

func actionShowFavorite(event *linebot.Event, action string, values url.Values) {
	columnCount := 9
	userId := values.Get("user_id")
	userFavorite := &controllers.UserFavorite{
		UserId:    userId,
		Favorites: []string{},
	}

	if currentPage, err := strconv.Atoi(values.Get("page")); err != nil {
		meta.Log.Println("Unable to parse parameters", values)
	} else {
		userData, _ := userFavorite.Get(meta)

		// reverse slice
		for i := len(userData.Favorites)/2 - 1; i >= 0; i-- {
			opp := len(userData.Favorites) - 1 - i
			userData.Favorites[i], userData.Favorites[opp] = userData.Favorites[opp], userData.Favorites[i]
		}

		startIdx := currentPage * columnCount
		endIdx := startIdx + columnCount
		lastPage := false
		if endIdx > len(userData.Favorites)-1 || startIdx > endIdx {
			endIdx = len(userData.Favorites)
			lastPage = true
		}

		fmt.Println("Start Index", startIdx)
		fmt.Println("End Index", endIdx)
		fmt.Println("Total Length", len(userData.Favorites))

		favDocuments := []models.ArticleDocument{}
		favs := userData.Favorites[startIdx:endIdx]
		fmt.Println(favs)

		for i := startIdx; i < endIdx; i++ {
			favArticleId := userData.Favorites[i]
			query := bson.M{"article_id": favArticleId}
			tmpRecord, _ := controllers.GetOne(meta.Collection, query)
			favDocuments = append(favDocuments, *tmpRecord)
		}

		// append next page column
		previousPage := currentPage - 1
		if previousPage < 0 {
			previousPage = 0
		}
		nextPage := currentPage + 1
		previousData := fmt.Sprintf("action=%s&page=%d&user_id=%s", ActonShowFav, previousPage, userId)
		nextData := fmt.Sprintf("action=%s&page=%d&user_id=%s", ActonShowFav, nextPage, userId)
		previousText := fmt.Sprintf("上一頁 %d", previousPage)
		nextText := fmt.Sprintf("下一頁 %d", nextPage)
		if lastPage == true {
			nextData = "--"
			nextText = "--"
		}

		tmpColumn := linebot.NewCarouselColumn(
			defaultThumbnail,
			DefaultTitle,
			"繼續看？",
			linebot.NewMessageAction(ActionHelp, ActionHelp),
			linebot.NewPostbackAction(previousText, previousData, "", ""),
			linebot.NewPostbackAction(nextText, nextData, "", ""),
		)

		template := getCarouseTemplate(event.Source.UserID, favDocuments)
		template.Columns = append(template.Columns, tmpColumn)
		sendCarouselMessage(event, template, "最愛照片已送達")
	}
}

func actionGeneral(event *linebot.Event, action string, values url.Values) {
	meta.Log.Println("Enter actionGeneral, action = ", action)
	meta.Log.Println("Enter actionGeneral, values = ", values)
	records := []models.ArticleDocument{}
	label := ""
	switch action {
	case ActionQuery:
		//meta.Log.Println(values.Get("period"))
		tsOffset, _ := strconv.Atoi(values.Get("period"))
		meta.Log.Println("timestampe off set = ", tsOffset)
		records, _ = controllers.GetMostLike(meta.Collection, maxCountOfCarousel, tsOffset)
		label = "已幫您查詢到一些文章~"
	case ActionRandom:
		records, _ = controllers.GetRandom(meta.Collection, maxCountOfCarousel, "")
		label = "隨機文章已送到囉"
	default:
		return
	}
	template := getCarouseTemplate(event.Source.UserID, records)
	if template != nil {
		sendCarouselMessage(event, template, label)
	}

}

func actionAllImage(event *linebot.Event, values url.Values) {
	if articleId := values.Get("article_id"); articleId != "" {
		query := bson.M{"article_id": articleId}
		result, _ := controllers.GetOne(meta.Collection, query)
		template := getImgCarousTemplate(result, values)
		sendImgCarouseMessage(event, template)
	} else {
		meta.Log.Println("Unable to get article id", values)
	}
}

func actionNewest(event *linebot.Event, values url.Values) {
	columnCount := 9
	if currentPage, err := strconv.Atoi(values.Get("page")); err != nil {
		meta.Log.Println("Unable to parse parameters", values)
	} else {
		records, _ := controllers.Get(meta.Collection, currentPage, columnCount)
		for idx, record := range records {
			meta.Log.Printf("ID: %d, Date: %s, Title: %s", idx, record.Date, record.ArticleTitle)
		}
		template := getCarouseTemplate(event.Source.UserID, records)

		if template == nil {
			meta.Log.Println("Unable to get template", values)
			return
		}

		// append next page column
		previousPage := currentPage - 1
		if previousPage < 0 {
			previousPage = 0
		}
		nextPage := currentPage + 1
		previousData := fmt.Sprintf("action=%s&page=%d", ActionNewest, previousPage)
		nextData := fmt.Sprintf("action=%s&page=%d", ActionNewest, nextPage)
		previousText := fmt.Sprintf("上一頁 %d", previousPage)
		nextText := fmt.Sprintf("下一頁 %d", nextPage)
		tmpColumn := linebot.NewCarouselColumn(
			defaultThumbnail,
			DefaultTitle,
			"繼續看？",
			linebot.NewMessageAction(ActionHelp, ActionHelp),
			linebot.NewPostbackAction(previousText, previousData, "", ""),
			linebot.NewPostbackAction(nextText, nextData, "", ""),
		)
		template.Columns = append(template.Columns, tmpColumn)

		sendCarouselMessage(event, template, "熱騰騰的最新文章送到了!")
	}
}

func getCarouseTemplate(userId string, records []models.ArticleDocument) (template *linebot.CarouselTemplate) {
	if len(records) == 0 {
		return nil
	}

	columnList := []*linebot.CarouselColumn{}
	userFavorite := &controllers.UserFavorite{
		UserId:    userId,
		Favorites: []string{},
	}
	userData, _ := userFavorite.Get(meta)
	favLabel := ""

	for _, result := range records {
		if exist, _ := utils.InArray(result.ArticleID, userData.Favorites); exist == true {
			favLabel = "❤️ 移除最愛"
		} else {
			favLabel = "💛 加入最愛"
		}
		thumnailUrl := defaultImage
		imgUrlCounts := len(result.ImageLinks)
		lable := fmt.Sprintf("%s (%d)", ActionAllImage, imgUrlCounts)
		title := result.ArticleTitle
		postBackData := fmt.Sprintf("action=%s&article_id=%s&page=0", ActionAllImage, result.ArticleID)
		text := fmt.Sprintf("%d 😍\t%d 😡", result.MessageCount.Push, result.MessageCount.Boo)

		if imgUrlCounts > 0 {
			thumnailUrl = result.ImageLinks[0]
		}

		// Title's hard limit by Line
		if len(title) >= 40 {
			title = title[0:38]
		}
		//meta.Log.Println("===============", idx)
		//meta.Log.Println("Thumbnail Url = ", thumnailUrl)
		//meta.Log.Println("Title = ", title)
		//meta.Log.Println("Text = ", text)
		//meta.Log.Println("URL = ", result.URL)
		//meta.Log.Println("===============", idx)
		//dataRandom := fmt.Sprintf("action=%s", ActionRandom)
		dataAddFavorite := fmt.Sprintf("action=%s&user_id=%s&article_id=%s",
			ActionAddFavorite, userId, result.ArticleID)
		tmpColumn := linebot.NewCarouselColumn(
			thumnailUrl,
			title,
			text,
			linebot.NewURIAction(ActionClick, result.URL),
			linebot.NewPostbackAction(lable, postBackData, "", ""),
			//linebot.NewPostbackAction(ActionRandom, dataRandom, "", ""),
			linebot.NewPostbackAction(favLabel, dataAddFavorite, "", ""),
		)
		columnList = append(columnList, tmpColumn)
	}
	template = linebot.NewCarouselTemplate(columnList...)
	return template
}

func postbackHandler(event *linebot.Event) {
	m, _ := url.ParseQuery(event.Postback.Data)
	action := m.Get("action")
	meta.Log.Println("Action = ", action)
	actionHandler(event, action, m)
}

func getUserNameById(userId string) (userDisplayName string) {
	res, err := bot.GetProfile(userId).Do()
	if err != nil {
		userDisplayName = "Unknown"
	} else {
		userDisplayName = res.DisplayName
	}
	return userDisplayName
}

func textHander(event *linebot.Event, message string) {
	userFavorite := &controllers.UserFavorite{
		UserId:    event.Source.UserID,
		Favorites: []string{},
	}
	if _, err := userFavorite.Get(meta); err != nil {
		meta.Log.Println("User data is not created, create a new one")
		userFavorite.Add(meta)
	}
	switch message {
	case ActionHelp:
		template := getMenuButtonTemplateV2(event, DefaultTitle)
		sendCarouselMessage(event, template, "我能為您做什麼？")
	case ActionRandom:
		records, _ := controllers.GetRandom(meta.Collection, maxCountOfCarousel, "")
		template := getCarouseTemplate(event.Source.UserID, records)
		sendCarouselMessage(event, template, "隨機表特已送到囉")
	case ActionNewest:
		values := url.Values{}
		values.Set("period", fmt.Sprintf("%d", oneDayInSec))
		values.Set("page", "0")
		actionNewest(event, values)
	case ActonShowFav:
		values := url.Values{}
		values.Set("user_id", event.Source.UserID)
		values.Set("page", "0")
		actionShowFavorite(event, "", values)
	default:
		if strings.HasPrefix(message, ActonRunCC) {
			commands := strings.Split(message, " ")
			action := commands[1]
			cmd := exec.Command("./run_cc.sh", action)
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				log.Fatal(err)
			}
			defer stdout.Close()
			if err := cmd.Start(); err != nil {
				log.Fatal(err)
			}
			// 读取输出结果
			opBytes, err := ioutil.ReadAll(stdout)
			if err != nil {
				log.Fatal(err)
			}
			log.Println(string(opBytes))
			sendTextMessage(event, string(opBytes))
			return
		}

		if event.Source.UserID != "" && event.Source.GroupID == "" && event.Source.RoomID == "" {
			records, _ := controllers.GetRandom(meta.Collection, maxCountOfCarousel, message)
			if records != nil && len(records) > 0 {
				template := getCarouseTemplate(event.Source.UserID, records)
				sendCarouselMessage(event, template, "隨機表特已送到囉")
			} else {
				template := getMenuButtonTemplateV2(event, DefaultTitle)
				sendCarouselMessage(event, template, "我能為您做什麼？")
			}
		}
	}
}

func getMenuButtonTemplateV2(event *linebot.Event, title string) (template *linebot.CarouselTemplate) {
	columnList := []*linebot.CarouselColumn{}
	dataNewlest := fmt.Sprintf("action=%s&page=0", ActionNewest)
	dataRandom := fmt.Sprintf("action=%s", ActionRandom)
	dataQuery := fmt.Sprintf("action=%s", ActionQuery)
	dataShowFav := fmt.Sprintf("action=%s&user_id=%s&page=0", ActonShowFav, event.Source.UserID)

	menu1 := linebot.NewCarouselColumn(
		defaultThumbnail,
		title,
		"你可以試試看以下選項，或直接輸入關鍵字查詢",
		linebot.NewPostbackAction(ActionNewest, dataNewlest, "", ""),
		linebot.NewPostbackAction(ActionRandom, dataRandom, "", ""),
		linebot.NewPostbackAction(ActonShowFav, dataShowFav, "", ""),
	)
	menu2 := linebot.NewCarouselColumn(
		defaultThumbnail,
		title,
		"你可以試試看以下選項，或直接輸入關鍵字查詢",
		linebot.NewPostbackAction(ActionDailyHot, dataQuery+"&period="+fmt.Sprintf("%d", oneDayInSec), "", ""),
		linebot.NewPostbackAction(ActionMonthlyHot, dataQuery+"&period="+fmt.Sprintf("%d", oneWeekInSec), "", ""),
		linebot.NewPostbackAction(ActionYearHot, dataQuery+"&period="+fmt.Sprintf("%d", oneYearInSec), "", ""),
	)
	columnList = append(columnList, menu1, menu2)
	template = linebot.NewCarouselTemplate(columnList...)
	return template
}

func sendTextMessage(event *linebot.Event, text string) {
	if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(text)).Do(); err != nil {
		log.Println("Send Fail")
	}
}

func getImgCarousTemplate(record *models.ArticleDocument, values url.Values) (template *linebot.ImageCarouselTemplate) {
	urls := record.ImageLinks
	columnList := []*linebot.ImageCarouselColumn{}
	articleID := values.Get("article_id")
	page, _ := strconv.Atoi(values.Get("page"))
	startIdx := page * 9
	endIdx := startIdx + 9
	lastPage := false
	if endIdx >= len(urls)-1 {
		endIdx = len(urls)
		lastPage = true
	}
	urls = urls[startIdx:endIdx]

	for _, url := range urls {
		tmpColumn := linebot.NewImageCarouselColumn(
			url,
			linebot.NewURIAction(ActionClick, url),
		)
		columnList = append(columnList, tmpColumn)
	}
	if lastPage == false {
		postBackData := fmt.Sprintf("action=%s&article_id=%s&page=%d", ActionAllImage, articleID, page+1)
		tmpColumn := linebot.NewImageCarouselColumn(
			defaultImage,
			linebot.NewPostbackAction("下一頁", postBackData, "", ""),
		)
		columnList = append(columnList, tmpColumn)
	}

	template = linebot.NewImageCarouselTemplate(columnList...)
	return template
}

func sendCarouselMessage(event *linebot.Event, template *linebot.CarouselTemplate, altText string) {
	if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTemplateMessage(altText, template)).Do(); err != nil {
		meta.Log.Println(err)
	}
}

func sendImgCarouseMessage(event *linebot.Event, template *linebot.ImageCarouselTemplate) {
	if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTemplateMessage("預覽圖片已送達", template)).Do(); err != nil {
		meta.Log.Println(err)
	}
}
