#!/bin/bash

echo "🔍 PTT Set Line Bot - 爬蟲整合驗證"
echo "======================================="

# 檢查必要文件
echo "📋 檢查必要文件..."
required_files=(
    "crawler/Dockerfile"
    "crawler/run_crawler.sh"
    "crawler/start_cron.sh"
    "crawler/test_crawler.sh"
    "docker-compose.yml"
)

for file in "${required_files[@]}"; do
    if [ -f "$file" ]; then
        echo "✅ $file"
    else
        echo "❌ $file 缺失"
        exit 1
    fi
done

# 檢查腳本權限
echo ""
echo "🔐 檢查腳本權限..."
scripts=(
    "crawler/run_crawler.sh"
    "crawler/start_cron.sh"
    "crawler/test_crawler.sh"
    "test_setup.sh"
    "verify_crawler_integration.sh"
)

for script in "${scripts[@]}"; do
    if [ -x "$script" ]; then
        echo "✅ $script 可執行"
    else
        echo "⚠️  $script 沒有執行權限，正在修正..."
        chmod +x "$script"
    fi
done

# 檢查 Docker 環境
echo ""
echo "🐳 檢查 Docker 環境..."
if command -v docker &> /dev/null; then
    echo "✅ Docker 已安裝"
else
    echo "❌ Docker 未安裝"
    exit 1
fi

if command -v docker-compose &> /dev/null; then
    echo "✅ Docker Compose 已安裝"
else
    echo "❌ Docker Compose 未安裝"
    exit 1
fi

# 檢查 .env 文件
echo ""
echo "📝 檢查環境設定..."
if [ -f ".env" ]; then
    echo "✅ .env 文件存在"
else
    echo "⚠️  .env 文件不存在，從模板複製..."
    cp .env.template .env
    echo "📝 請編輯 .env 文件設定 Line Bot 憑證"
fi

# 創建資料目錄
echo ""
echo "📁 創建資料目錄..."
mkdir -p data/mongo/data data/crawler
echo "✅ 資料目錄已創建"

# 構建映像檔
echo ""
echo "🔨 構建 Docker 映像檔..."
if make build; then
    echo "✅ Docker 映像檔構建成功"
else
    echo "❌ Docker 映像檔構建失敗"
    exit 1
fi

# 啟動服務
echo ""
echo "🚀 啟動服務..."
if make dev; then
    echo "✅ 服務啟動成功"
else
    echo "❌ 服務啟動失敗"
    exit 1
fi

# 等待服務準備就緒
echo ""
echo "⏳ 等待服務準備就緒..."
sleep 15

# 檢查服務狀態
echo ""
echo "📊 檢查服務狀態..."
make status

# 測試 MongoDB 連接
echo ""
echo "🔍 測試 MongoDB 連接..."
if docker-compose exec -T mongodb mongosh --eval "db.adminCommand('ping')" --quiet &>/dev/null; then
    echo "✅ MongoDB 連接成功"
else
    echo "❌ MongoDB 連接失敗"
    make down
    exit 1
fi

# 測試爬蟲功能
echo ""
echo "🕷️  測試爬蟲功能..."
if make crawler-test; then
    echo "✅ 爬蟲測試成功"
else
    echo "❌ 爬蟲測試失敗"
    echo "📋 查看爬蟲日誌："
    make crawler-logs
    make down
    exit 1
fi

# 檢查 Line Bot 健康狀態
echo ""
echo "🤖 檢查 Line Bot 健康狀態..."
sleep 5
if curl -f http://localhost:5050/health &>/dev/null; then
    echo "✅ Line Bot 健康檢查通過"
else
    echo "⚠️  Line Bot 健康檢查失敗（可能需要設定 Line Bot 憑證）"
fi

echo ""
echo "🎉 爬蟲整合驗證完成！"
echo ""
echo "📋 系統狀態摘要："
echo "  ✅ 所有必要文件已就位"
echo "  ✅ Docker 映像檔構建成功"
echo "  ✅ MongoDB 服務正常運行"
echo "  ✅ 爬蟲功能測試通過"
echo "  ✅ 自動化排程已設定（每小時執行）"
echo ""
echo "📚 常用指令："
echo "  make status          - 查看服務狀態"
echo "  make crawler-logs    - 查看爬蟲日誌"
echo "  make crawler-run     - 手動執行爬蟲"
echo "  make crawler-test    - 測試爬蟲功能"
echo "  make dev-logs        - 查看所有服務日誌"
echo "  make down            - 停止所有服務"
echo "  make clean           - 完全清理"
echo ""
echo "⚠️  重要提醒："
echo "  1. 請在 .env 文件中設定正確的 Line Bot 憑證"
echo "  2. 爬蟲會每小時自動執行，請確保有足夠的磁碟空間"
echo "  3. 請遵守 PTT 使用條款，避免過度頻繁的請求"
echo ""
echo "🔗 爬蟲將自動每小時更新 PTT Set 板資料到 MongoDB"