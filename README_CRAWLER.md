# 🕷️ PTT 爬蟲整合完成！

## ✨ 新功能概述

已成功將 PTT Set 板爬蟲整合到 docker-compose 中，現在系統會：

- 🔄 **每小時自動爬取** PTT Set 板最新資料
- 📦 **自動匯入 MongoDB** 並建立唯一索引
- 🚀 **容器化部署** 所有服務（Line Bot + 爬蟲 + MongoDB）
- 📊 **完整日誌記錄** 便於監控和除錯

## 🚀 快速開始

### 1. 測試整個設置
```bash
./test_setup.sh
```

### 2. 啟動所有服務
```bash
make dev
```

### 3. 監控爬蟲運行
```bash
make crawler-logs
```

## 📋 新增的文件

```
crawler/
├── Dockerfile          # 爬蟲容器配置
├── run_crawler.sh      # 爬蟲執行腳本
└── start_cron.sh       # Cron 啟動腳本

CRAWLER_SETUP.md        # 詳細設置說明
test_setup.sh          # 自動化測試腳本
README_CRAWLER.md       # 本文件
```

## 🔧 主要改動

### docker-compose.yml
- ✅ 新增 `crawler` 服務
- ✅ 設定服務依賴關係
- ✅ 配置環境變數和資料卷

### Makefile
- ✅ 新增爬蟲相關指令
- ✅ 自動創建資料目錄
- ✅ 增強的日誌查看功能

### .env.template
- ✅ 新增爬蟲配置選項

## ⚙️ 爬蟲配置

| 環境變數 | 預設值 | 說明 |
|---------|--------|------|
| `PAGE_OFFSET` | 10 | 爬取頁數 |
| `RUN_IMMEDIATELY` | true | 啟動時立即執行 |
| `MONGODB_HOST` | mongodb | MongoDB 主機 |
| `MONGODB_DB` | ptt | 資料庫名稱 |

## 📊 監控指令

```bash
# 查看所有服務狀態
make status

# 查看爬蟲日誌
make crawler-logs

# 手動執行爬蟲
make crawler-run

# 查看所有服務日誌
make dev-logs

# 停止所有服務
make down

# 完全清理（包含資料）
make clean
```

## 🕐 執行時程

- **自動執行**: 每小時的第 0 分鐘
- **首次執行**: 容器啟動時立即執行一次
- **重試機制**: 失敗時會記錄錯誤並等待下次執行

## 📈 資料流程

```
PTT Set 板 → Python 爬蟲 → Set.json → MongoDB → Line Bot
     ↑              ↑             ↑           ↑         ↑
   每小時更新    Docker容器    自動備份    唯一索引   即時查詢
```

## 🛠️ 故障排除

### 常見問題

1. **爬蟲無法啟動**
   ```bash
   # 檢查容器狀態
   docker-compose ps
   
   # 查看錯誤日誌
   make crawler-logs
   ```

2. **MongoDB 連接失敗**
   ```bash
   # 重啟 MongoDB
   docker-compose restart mongodb
   
   # 檢查網路連接
   docker-compose exec crawler ping mongodb
   ```

3. **資料匯入失敗**
   ```bash
   # 手動執行爬蟲查看詳細錯誤
   make crawler-run
   ```

## 🔒 注意事項

1. **遵守使用條款**: 請確保爬蟲使用符合 PTT 的服務條款
2. **資源監控**: 定期檢查磁碟空間和記憶體使用量
3. **備份策略**: 建議定期備份 MongoDB 資料
4. **日誌管理**: 定期清理日誌文件

## 🎯 下一步

1. 設定 Line Bot 憑證到 `.env` 文件
2. 執行 `make dev` 啟動所有服務
3. 使用 `make crawler-logs` 監控爬蟲運行
4. 測試 Line Bot 功能

---

🎉 **恭喜！您的 PTT Set Line Bot 現在具備自動資料更新功能！**