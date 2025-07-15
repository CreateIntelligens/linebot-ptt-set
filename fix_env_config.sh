#!/bin/bash

echo "🔧 修復 .env 配置文件"
echo "===================="

# 備份原始文件
if [ -f ".env" ]; then
    cp .env .env.backup.$(date +%Y%m%d_%H%M%S)
    echo "✅ 已備份原始 .env 文件"
fi

# 創建修復後的 .env 文件
cat > .env << 'EOF'
# MongoDB settings for docker-compose
MONGO_ROOT_USERNAME=admin
MONGO_ROOT_PASSWORD=password123
MONGO_INITDB_DATABASE=ptt

# MongoDB connection for application
MongoDBURI=mongodb://admin:password123@mongodb:27017/ptt?authSource=admin
MongoDBSSL=false

# Line Bot Configuration
# 注意：這些是測試值，請替換為實際的Line Bot憑證
# 獲取方式：https://developers.line.biz/en/services/messaging-api/
ChannelSecret=test_channel_secret_for_development
ChannelAccessToken=test_channel_access_token_for_development

# Application settings
PORT=5050

# Crawler settings
CRAWLER_PAGE_OFFSET=10
CRAWLER_RUN_IMMEDIATELY=true

# Debug settings (可選)
DEBUG=true
LOG_LEVEL=info
EOF

echo "✅ .env 文件已更新"

# 檢查配置
echo ""
echo "📋 當前配置："
echo "=============="
cat .env

echo ""
echo "⚠️  重要提醒："
echo "1. Line Bot 憑證目前使用測試值"
echo "2. 如需完整功能，請到 https://developers.line.biz/ 申請真實憑證"
echo "3. MongoDB 配置已修復，應該可以正常連接"
echo ""
echo "🚀 下一步："
echo "1. 重啟服務: docker-compose down && docker-compose up -d"
echo "2. 檢查連接: curl http://localhost:5050/health"
echo "3. 查看日誌: tail -f logs/pttset.log"
echo "4. 訪問管理面板: http://localhost:5050/admin"