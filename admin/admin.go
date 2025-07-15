package admin

import (
"encoding/json"
"net/http"
"os/exec"
"strconv"
"strings"
"time"

"github.com/mong0520/linebot-ptt-set/controllers"
"github.com/mong0520/linebot-ptt-set/models"
"gopkg.in/mgo.v2/bson"
)

var meta *models.Model

type AdminData struct {
	CrawlerStatus   CrawlerStatus         `json:"crawler_status"`
	DatabaseStats   DatabaseStats         `json:"database_stats"`
	RecentArticles  []models.ArticleDocument `json:"recent_articles"`
	SystemInfo      SystemInfo            `json:"system_info"`
}

type CrawlerStatus struct {
	IsRunning     bool      `json:"is_running"`
	LastRunTime   time.Time `json:"last_run_time"`
	NextRunTime   time.Time `json:"next_run_time"`
	LastRunStatus string    `json:"last_run_status"`
	LogTail       []string  `json:"log_tail"`
}

type DatabaseStats struct {
	TotalArticles    int64     `json:"total_articles"`
	TodayArticles    int64     `json:"today_articles"`
	WeekArticles     int64     `json:"week_articles"`
	LastUpdateTime   time.Time `json:"last_update_time"`
	TopAuthors       []AuthorStat `json:"top_authors"`
	PopularArticles  []models.ArticleDocument `json:"popular_articles"`
}

type AuthorStat struct {
	Author string `json:"author"`
	Count  int    `json:"count"`
}

type SystemInfo struct {
	Uptime        string `json:"uptime"`
	GoVersion     string `json:"go_version"`
	MongoStatus   string `json:"mongo_status"`
	DiskUsage     string `json:"disk_usage"`
}

func InitAdmin(m *models.Model) {
	meta = m
	http.HandleFunc("/admin", adminHandler)
	http.HandleFunc("/admin/api/status", apiStatusHandler)
	http.HandleFunc("/admin/api/articles", apiArticlesHandler)
	http.HandleFunc("/admin/api/crawler/run", apiCrawlerRunHandler)
	http.HandleFunc("/admin/api/crawler/logs", apiCrawlerLogsHandler)
	http.HandleFunc("/admin/static/", staticHandler)
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html lang="zh-TW">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PTT SET Bot 管理面板</title>
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
        .stat-value { font-weight: bold; color: #333; }
        .status-indicator { display: inline-block; width: 12px; height: 12px; border-radius: 50%; margin-right: 8px; }
        .status-running { background: #28a745; }
        .status-stopped { background: #dc3545; }
        .status-warning { background: #ffc107; }
        .btn { padding: 10px 20px; border: none; border-radius: 5px; cursor: pointer; font-size: 14px; transition: all 0.2s; }
        .btn-primary { background: #007bff; color: white; }
        .btn-success { background: #28a745; color: white; }
        .btn-warning { background: #ffc107; color: #212529; }
        .btn:hover { opacity: 0.8; transform: translateY(-1px); }
        .log-container { background: #1e1e1e; color: #f8f8f2; padding: 15px; border-radius: 5px; font-family: 'Courier New', monospace; font-size: 12px; max-height: 300px; overflow-y: auto; }
        .table { width: 100%; border-collapse: collapse; margin-top: 15px; }
        .table th, .table td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        .table th { background: #f8f9fa; font-weight: 600; }
        .table tr:hover { background: #f5f5f5; }
        .loading { text-align: center; padding: 20px; color: #666; }
        .refresh-btn { float: right; margin-bottom: 10px; }
        .article-title { max-width: 300px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .push-count { color: #28a745; font-weight: bold; }
        .boo-count { color: #dc3545; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔥 PTT SET Bot 管理面板</h1>
            <p>監控爬蟲狀態 • 查看數據統計 • 系統管理</p>
        </div>

        <div class="dashboard">
            <!-- 爬蟲狀態 -->
            <div class="card">
                <h3>🕷️ 爬蟲狀態</h3>
                <div id="crawler-status" class="loading">載入中...</div>
                <button class="btn btn-success" onclick="runCrawler()">立即執行爬蟲</button>
            </div>

            <!-- 數據庫統計 -->
            <div class="card">
                <h3>📊 數據庫統計</h3>
                <div id="db-stats" class="loading">載入中...</div>
            </div>

            <!-- 系統資訊 -->
            <div class="card">
                <h3>⚙️ 系統資訊</h3>
                <div id="system-info" class="loading">載入中...</div>
            </div>
        </div>

        <!-- 爬蟲日誌 -->
        <div class="card">
            <h3>📝 爬蟲日誌</h3>
            <button class="btn btn-primary refresh-btn" onclick="refreshLogs()">刷新日誌</button>
            <div id="crawler-logs" class="log-container">載入中...</div>
        </div>

        <!-- 最新文章 -->
        <div class="card">
            <h3>📰 最新文章</h3>
            <button class="btn btn-primary refresh-btn" onclick="refreshArticles()">刷新列表</button>
            <div id="articles-list" class="loading">載入中...</div>
        </div>

        <!-- 熱門文章 -->
        <div class="card">
            <h3>🔥 熱門文章</h3>
            <div id="popular-articles" class="loading">載入中...</div>
        </div>
    </div>

    <script>
        // 載入所有數據
        function loadAllData() {
            loadStatus();
            loadArticles();
            loadLogs();
        }

        // 載入狀態數據
        function loadStatus() {
            fetch('/admin/api/status')
                .then(response => response.json())
                .then(data => {
                    updateCrawlerStatus(data.crawler_status);
                    updateDatabaseStats(data.database_stats);
                    updateSystemInfo(data.system_info);
                    updatePopularArticles(data.database_stats.popular_articles);
                })
                .catch(error => {
                    console.error('Error loading status:', error);
                });
        }

        // 載入文章列表
        function loadArticles() {
            fetch('/admin/api/articles?limit=20')
                .then(response => response.json())
                .then(data => {
                    updateArticlesList(data);
                })
                .catch(error => {
                    console.error('Error loading articles:', error);
                });
        }

        // 載入日誌
        function loadLogs() {
            fetch('/admin/api/crawler/logs')
                .then(response => response.json())
                .then(data => {
                    updateLogs(data.logs);
                })
                .catch(error => {
                    console.error('Error loading logs:', error);
                });
        }

        // 更新爬蟲狀態
        function updateCrawlerStatus(status) {
            const statusHtml = ` + "`" + `
                <div class="stat">
                    <span class="stat-label">
                        <span class="status-indicator ${status.is_running ? 'status-running' : 'status-stopped'}"></span>
                        運行狀態
                    </span>
                    <span class="stat-value">${status.is_running ? '運行中' : '已停止'}</span>
                </div>
                <div class="stat">
                    <span class="stat-label">上次執行</span>
                    <span class="stat-value">${formatTime(status.last_run_time)}</span>
                </div>
                <div class="stat">
                    <span class="stat-label">下次執行</span>
                    <span class="stat-value">${formatTime(status.next_run_time)}</span>
                </div>
                <div class="stat">
                    <span class="stat-label">執行狀態</span>
                    <span class="stat-value">${status.last_run_status}</span>
                </div>
            ` + "`" + `;
            document.getElementById('crawler-status').innerHTML = statusHtml;
        }

        // 更新數據庫統計
        function updateDatabaseStats(stats) {
            const statsHtml = ` + "`" + `
                <div class="stat">
                    <span class="stat-label">總文章數</span>
                    <span class="stat-value">${stats.total_articles.toLocaleString()}</span>
                </div>
                <div class="stat">
                    <span class="stat-label">今日新增</span>
                    <span class="stat-value">${stats.today_articles}</span>
                </div>
                <div class="stat">
                    <span class="stat-label">本週新增</span>
                    <span class="stat-value">${stats.week_articles}</span>
                </div>
                <div class="stat">
                    <span class="stat-label">最後更新</span>
                    <span class="stat-value">${formatTime(stats.last_update_time)}</span>
                </div>
            ` + "`" + `;
            document.getElementById('db-stats').innerHTML = statsHtml;
        }

        // 更新系統資訊
        function updateSystemInfo(info) {
            const infoHtml = ` + "`" + `
                <div class="stat">
                    <span class="stat-label">運行時間</span>
                    <span class="stat-value">${info.uptime}</span>
                </div>
                <div class="stat">
                    <span class="stat-label">Go 版本</span>
                    <span class="stat-value">${info.go_version}</span>
                </div>
                <div class="stat">
                    <span class="stat-label">MongoDB 狀態</span>
                    <span class="stat-value">${info.mongo_status}</span>
                </div>
                <div class="stat">
                    <span class="stat-label">磁碟使用</span>
                    <span class="stat-value">${info.disk_usage}</span>
                </div>
            ` + "`" + `;
            document.getElementById('system-info').innerHTML = infoHtml;
        }

        // 更新文章列表
        function updateArticlesList(articles) {
            let html = '<table class="table"><thead><tr><th>標題</th><th>作者</th><th>日期</th><th>推/噓</th><th>圖片</th></tr></thead><tbody>';
            articles.forEach(article => {
                html += ` + "`" + `<tr>
                    <td class="article-title" title="${article.article_title}">${article.article_title}</td>
                    <td>${article.author}</td>
                    <td>${article.date}</td>
                    <td><span class="push-count">${article.message_count.push}</span>/<span class="boo-count">${article.message_count.boo}</span></td>
                    <td>${article.image_links ? article.image_links.length : 0}</td>
                </tr>` + "`" + `;
            });
            html += '</tbody></table>';
            document.getElementById('articles-list').innerHTML = html;
        }

        // 更新熱門文章
        function updatePopularArticles(articles) {
            let html = '<table class="table"><thead><tr><th>標題</th><th>作者</th><th>推數</th><th>圖片</th></tr></thead><tbody>';
            articles.forEach(article => {
                html += ` + "`" + `<tr>
                    <td class="article-title" title="${article.article_title}">${article.article_title}</td>
                    <td>${article.author}</td>
                    <td><span class="push-count">${article.message_count.push}</span></td>
                    <td>${article.image_links ? article.image_links.length : 0}</td>
                </tr>` + "`" + `;
            });
            html += '</tbody></table>';
            document.getElementById('popular-articles').innerHTML = html;
        }

        // 更新日誌
        function updateLogs(logs) {
            const logsHtml = logs.join('\n');
            document.getElementById('crawler-logs').textContent = logsHtml;
            // 自動滾動到底部
            const logContainer = document.getElementById('crawler-logs');
            logContainer.scrollTop = logContainer.scrollHeight;
        }

        // 執行爬蟲
        function runCrawler() {
            if (confirm('確定要立即執行爬蟲嗎？')) {
                fetch('/admin/api/crawler/run', { method: 'POST' })
                    .then(response => response.json())
                    .then(data => {
                        alert(data.message);
                        setTimeout(loadAllData, 2000); // 2秒後刷新數據
                    })
                    .catch(error => {
                        alert('執行失敗: ' + error);
                    });
            }
        }

        // 刷新日誌
        function refreshLogs() {
            loadLogs();
        }

        // 刷新文章
        function refreshArticles() {
            loadArticles();
        }

        // 格式化時間
        function formatTime(timeStr) {
            if (!timeStr || timeStr === '0001-01-01T00:00:00Z') return '未知';
            const date = new Date(timeStr);
            return date.toLocaleString('zh-TW');
        }

        // 頁面載入時執行
        document.addEventListener('DOMContentLoaded', function() {
            loadAllData();
            // 每30秒自動刷新狀態
            setInterval(loadStatus, 30000);
        });
    </script>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(tmpl))
}

func apiStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	data := AdminData{
		CrawlerStatus: getCrawlerStatus(),
		DatabaseStats: getDatabaseStats(),
		SystemInfo:    getSystemInfo(),
	}
	
	json.NewEncoder(w).Encode(data)
}

func apiArticlesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	
	articles, _ := controllers.Get(meta.Collection, 0, limit)
	json.NewEncoder(w).Encode(articles)
}

func apiCrawlerRunHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// 執行爬蟲
	cmd := exec.Command("docker-compose", "exec", "-T", "crawler", "/app/run_crawler.sh")
	err := cmd.Start()
	
	response := map[string]string{}
	if err != nil {
		response["status"] = "error"
		response["message"] = "爬蟲執行失敗: " + err.Error()
	} else {
		response["status"] = "success"
		response["message"] = "爬蟲已開始執行"
	}
	
	json.NewEncoder(w).Encode(response)
}

func apiCrawlerLogsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// 獲取爬蟲日誌
	cmd := exec.Command("docker-compose", "logs", "--tail=50", "crawler")
	output, err := cmd.Output()
	
	logs := []string{}
	if err == nil {
		logLines := strings.Split(string(output), "\n")
		for _, line := range logLines {
			if strings.TrimSpace(line) != "" {
				logs = append(logs, line)
			}
		}
	} else {
		logs = append(logs, "無法獲取日誌: "+err.Error())
	}
	
	response := map[string]interface{}{
		"logs": logs,
	}
	
	json.NewEncoder(w).Encode(response)
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	// 簡單的靜態文件處理
	http.ServeFile(w, r, r.URL.Path[1:])
}

func getCrawlerStatus() CrawlerStatus {
	status := CrawlerStatus{
		IsRunning:     false,
		LastRunTime:   time.Time{},
		NextRunTime:   getNextHour(),
		LastRunStatus: "未知",
		LogTail:       []string{},
	}
	
	// 檢查爬蟲容器是否運行
	cmd := exec.Command("docker-compose", "ps", "-q", "crawler")
	output, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(output)) != "" {
		status.IsRunning = true
	}
	
	// 獲取最後執行時間（從日誌中解析）
	cmd = exec.Command("docker-compose", "logs", "--tail=10", "crawler")
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Crawler job completed successfully") {
				status.LastRunStatus = "成功"
				// 這裡可以解析時間戳
				break
			} else if strings.Contains(line, "ERROR") {
				status.LastRunStatus = "失敗"
			}
		}
	}
	
	return status
}

func getDatabaseStats() DatabaseStats {
	stats := DatabaseStats{
		TotalArticles:   0,
		TodayArticles:   0,
		WeekArticles:    0,
		LastUpdateTime:  time.Time{},
		TopAuthors:      []AuthorStat{},
		PopularArticles: []models.ArticleDocument{},
	}
	
	ctx := context.Background()
	
	// 總文章數
	if count, err := meta.Collection.CountDocuments(ctx, bson.M{}); err == nil {
		stats.TotalArticles = count
	}
	
	// 今日文章數
	today := time.Now().Truncate(24 * time.Hour)
	todayTimestamp := int(today.Unix())
	if count, err := meta.Collection.CountDocuments(ctx, bson.M{"timestamp": bson.M{"$gte": todayTimestamp}}); err == nil {
		stats.TodayArticles = count
	}
	
	// 本週文章數
	weekAgo := time.Now().AddDate(0, 0, -7)
	weekTimestamp := int(weekAgo.Unix())
	if count, err := meta.Collection.CountDocuments(ctx, bson.M{"timestamp": bson.M{"$gte": weekTimestamp}}); err == nil {
		stats.WeekArticles = count
	}
	
	// 最後更新時間（最新文章的時間）
	var latestArticle models.ArticleDocument
	opts := options.FindOne().SetSort(bson.D{{"timestamp", -1}})
	if err := meta.Collection.FindOne(ctx, bson.M{}, opts).Decode(&latestArticle); err == nil {
		stats.LastUpdateTime = time.Unix(int64(latestArticle.Timestamp), 0)
	}
	
	// 熱門文章
	if articles, err := controllers.GetMostLike(meta.Collection, 10, 0); err == nil {
		stats.PopularArticles = articles
	}
	
	return stats
}

func getSystemInfo() SystemInfo {
	info := SystemInfo{
		Uptime:      "未知",
		GoVersion:   "未知",
		MongoStatus: "未知",
		DiskUsage:   "未知",
	}
	
	// Go 版本
	if output, err := exec.Command("go", "version").Output(); err == nil {
		info.GoVersion = strings.TrimSpace(string(output))
	}
	
	// MongoDB 狀態
	cmd := exec.Command("docker-compose", "exec", "-T", "mongodb", "mongosh", "--eval", "db.adminCommand('ping')", "--quiet")
	if err := cmd.Run(); err == nil {
		info.MongoStatus = "正常"
	} else {
		info.MongoStatus = "異常"
	}
	
	// 磁碟使用量
	if output, err := exec.Command("df", "-h", ".").Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) >= 5 {
				info.DiskUsage = fields[4] // 使用百分比
			}
		}
	}
	
	return info
}

func getNextHour() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location())
}
