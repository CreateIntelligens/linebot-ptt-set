# 🎛️ PTT SET Bot 管理面板

## 概述

管理面板提供了一個Web界面來監控PTT SET爬蟲的運行狀態、查看MongoDB數據統計，以及管理整個系統。

## 功能特色

### 📊 實時監控
- **爬蟲狀態**: 查看爬蟲是否運行、上次執行時間、下次執行時間
- **數據庫統計**: 總文章數、今日新增、本週新增文章數量
- **系統資訊**: 運行時間、Go版本、MongoDB狀態、磁碟使用量

### 📝 日誌查看
- **爬蟲日誌**: 實時查看爬蟲執行日誌
- **自動刷新**: 每30秒自動更新狀態信息
- **錯誤追蹤**: 快速識別爬蟲執行中的問題

### 📰 數據管理
- **最新文章**: 查看最近爬取的文章列表
- **熱門文章**: 按推文數排序的熱門內容
- **文章詳情**: 標題、作者、日期、推噓數、圖片數量

### 🔧 系統控制
- **手動執行爬蟲**: 一鍵觸發爬蟲立即執行
- **日誌刷新**: 手動刷新各種日誌和數據
- **狀態監控**: 實時監控各服務運行狀態

## 快速開始

### 1. 啟動服務
```bash
# 啟動所有服務（包含管理面板）
make dev

# 檢查服務狀態
make status
```

### 2. 訪問管理面板
```bash
# 自動打開管理面板
make admin

# 或手動訪問
# 瀏覽器打開: http://localhost:5050/admin
```

### 3. 測試管理面板
```bash
# 運行管理面板測試
./test_admin_panel.sh

# 或使用Makefile
make test-admin
```

## API 端點

### 管理面板主頁
- **URL**: `http://localhost:5050/admin`
- **方法**: GET
- **描述**: 管理面板主界面

### 狀態API
- **URL**: `http://localhost:5050/admin/api/status`
- **方法**: GET
- **描述**: 獲取系統狀態、爬蟲狀態、數據庫統計
- **回應**: JSON格式的狀態數據

### 文章API
- **URL**: `http://localhost:5050/admin/api/articles?limit=20`
- **方法**: GET
- **參數**: 
  - `limit`: 返回文章數量（預設20）
- **描述**: 獲取最新文章列表

### 爬蟲控制API
- **URL**: `http://localhost:5050/admin/api/crawler/run`
- **方法**: POST
- **描述**: 手動觸發爬蟲執行

### 日誌API
- **URL**: `http://localhost:5050/admin/api/crawler/logs`
- **方法**: GET
- **描述**: 獲取爬蟲執行日誌

## 使用指南

### 監控爬蟲狀態
1. 打開管理面板
2. 查看「爬蟲狀態」卡片
3. 綠色指示器表示運行中，紅色表示已停止
4. 查看上次執行時間和下次執行時間

### 查看數據統計
1. 在「數據庫統計」卡片中查看：
   - 總文章數
   - 今日新增文章數
   - 本週新增文章數
   - 最後更新時間

### 手動執行爬蟲
1. 點擊「立即執行爬蟲」按鈕
2. 確認執行
3. 系統會顯示執行結果
4. 2秒後自動刷新狀態

### 查看日誌
1. 滾動到「爬蟲日誌」區域
2. 查看最近50行日誌
3. 點擊「刷新日誌」獲取最新日誌
4. 日誌會自動滾動到底部

### 瀏覽文章數據
1. 在「最新文章」表格中查看最近文章
2. 在「熱門文章」表格中查看高推文文章
3. 點擊「刷新列表」更新數據

## 故障排除

### 管理面板無法訪問
```bash
# 檢查服務狀態
make status

# 檢查應用程式日誌
docker-compose logs app

# 檢查端口是否被占用
netstat -tulpn | grep 5050
```

### API返回錯誤
```bash
# 檢查MongoDB連接
docker-compose exec mongodb mongosh --eval "db.adminCommand('ping')"

# 檢查應用程式日誌
make admin-logs
```

### 爬蟲狀態顯示異常
```bash
# 檢查爬蟲容器
docker-compose ps crawler

# 查看爬蟲日誌
make crawler-logs

# 手動執行爬蟲測試
make crawler-run
```

### 數據統計不正確
```bash
# 檢查MongoDB數據
docker-compose exec mongodb mongosh ptt --eval "db.set.countDocuments()"

# 重啟應用程式
docker-compose restart app
```

## 安全注意事項

1. **訪問控制**: 管理面板目前沒有身份驗證，請確保只在安全環境中使用
2. **網路安全**: 不要將管理面板暴露到公網
3. **權限管理**: 確保Docker有足夠權限執行爬蟲命令
4. **資源監控**: 定期檢查系統資源使用情況

## 自定義配置

### 修改刷新間隔
編輯 `admin/admin.go` 中的JavaScript部分：
```javascript
// 每30秒自動刷新狀態（預設）
setInterval(loadStatus, 30000);

// 改為每60秒刷新
setInterval(loadStatus, 60000);
```

### 調整日誌顯示行數
修改 `apiCrawlerLogsHandler` 函數：
```go
// 顯示最近50行（預設）
cmd := exec.Command("docker-compose", "logs", "--tail=50", "crawler")

// 改為顯示最近100行
cmd := exec.Command("docker-compose", "logs", "--tail=100", "crawler")
```

### 自定義文章顯示數量
在API調用中修改limit參數：
```javascript
// 預設顯示20篇文章
fetch('/admin/api/articles?limit=20')

// 改為顯示50篇文章
fetch('/admin/api/articles?limit=50')
```

## 開發指南

### 添加新的監控指標
1. 在 `AdminData` 結構體中添加新字段
2. 在相應的獲取函數中實現邏輯
3. 在HTML模板中添加顯示元素
4. 在JavaScript中添加更新邏輯

### 添加新的API端點
1. 在 `InitAdmin` 函數中註冊新路由
2. 實現對應的處理函數
3. 添加必要的錯誤處理
4. 更新前端JavaScript調用

### 自定義樣式
管理面板使用內嵌CSS，可以直接修改 `adminHandler` 函數中的樣式部分。

---

🎉 **管理面板讓PTT SET Bot的運維變得更加簡單高效！**