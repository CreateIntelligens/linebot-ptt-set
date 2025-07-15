package bots

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/mong0520/linebot-ptt-set/admin"
	"github.com/mong0520/linebot-ptt-set/controllers"
	"github.com/mong0520/linebot-ptt-set/models"
	"github.com/mong0520/linebot-ptt-set/utils"
	"gopkg.in/mgo.v2/bson"
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
	
	// 初始化管理面板
	admin.InitAdmin(meta)
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
