# PTT 爬蟲整合設定

## 概述

此專案已整合 PTT SET 板爬蟲，會自動每小時爬取最新資料並匯入 MongoDB。

## 新增的服務

### Crawler 服務
- **功能**: 每小時自動爬取 PTT SET 板資料
- **基於**: [ptt-web-crawler](https://github.com/mong0520/ptt-web-crawler)
- **排程**: 使用 cron 每小時執行一次 (0 * * * *)

## 快速開始

### 1. 設定環境變數
```bash
cp .env.template .env
# 編輯 .env 文件，設定 Line Bot 相關參數
```

### 2. 啟動所有服務
```bash
make dev
```

這會啟動：
- MongoDB 資料庫
- PTT 爬蟲 (每小時執行)
- Line Bot 應用程式

### 3. 查看服務狀態
```bash
make status
```

### 4. 查看爬蟲日誌
```bash
make crawler-logs
```

### 5. 手動執行爬蟲
```bash
make crawler-run
```

## 環境變數說明

### 爬蟲相關設定
- `CRAWLER_PAGE_OFFSET`: 爬取頁數偏移量 (預設: 10)
- `CRAWLER_RUN_IMMEDIATELY`: 容器啟動時是否立即執行爬蟲 (預設: true)

## 資料目錄

```
data/
├── mongo/data/     # MongoDB 資料存放目錄
└── crawler/        # 爬蟲相關文件存放目錄
```

## 爬蟲執行流程

1. **每小時觸發**: cron job 每小時的第 0 分鐘執行
2. **等待 MongoDB**: 確保資料庫服務可用
3. **備份舊資料**: 將現有 SET.json 備份
4. **執行爬蟲**: 使用 Python 爬蟲取得最新資料
5. **匯入資料**: 使用 mongoimport 將資料匯入 MongoDB
6. **建立索引**: 確保 article_id 為唯一索引
7. **清理檔案**: 保留最近 3 個備份文件

## 監控和維護

### 查看爬蟲日誌
```bash
# 即時查看日誌
make crawler-logs

# 或直接使用 docker-compose
docker-compose logs -f crawler
```

### 手動執行爬蟲
```bash
# 立即執行一次爬蟲
make crawler-run

# 或直接使用 docker-compose
docker-compose exec crawler /app/run_crawler.sh
```

### 重啟爬蟲服務
```bash
docker-compose restart crawler
```

### 停止所有服務
```bash
make down
```

### 完全清理 (包含資料)
```bash
make clean
```

## 故障排除

### 1. 爬蟲無法連接 MongoDB
- 檢查 MongoDB 服務是否正常運行
- 確認網路連接設定

### 2. 爬蟲執行失敗
- 查看爬蟲日誌: `make crawler-logs`
- 檢查 PTT 網站是否可正常訪問
- 確認爬蟲腳本是否需要更新

### 3. 資料匯入失敗
- 檢查 SET.json 文件是否正確產生
- 確認 MongoDB 有足夠空間
- 檢查資料格式是否正確

## 自訂設定

### 修改爬取頻率
編輯 `crawler/Dockerfile` 中的 cron 設定：
```dockerfile
# 每30分鐘執行一次
RUN echo "*/30 * * * * /app/run_crawler.sh >> /var/log/crawler.log 2>&1" | crontab -

# 每6小時執行一次
RUN echo "0 */6 * * * /app/run_crawler.sh >> /var/log/crawler.log 2>&1" | crontab -
```

### 修改爬取頁數
在 `.env` 文件中調整：
```
CRAWLER_PAGE_OFFSET=20  # 增加爬取頁數
```

## 注意事項

1. **資源使用**: 爬蟲會定期執行，請確保有足夠的系統資源
2. **網路政策**: 請遵守 PTT 的使用條款，避免過於頻繁的請求
3. **資料備份**: 建議定期備份 MongoDB 資料
4. **日誌管理**: 定期清理日誌文件以避免磁碟空間不足