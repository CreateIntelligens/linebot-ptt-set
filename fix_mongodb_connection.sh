#!/bin/bash

echo "🔧 PTT SET Bot - MongoDB連接修復腳本"
echo "=================================="

# 創建必要的目錄
echo "📁 創建必要目錄..."
mkdir -p logs data/mongo/data data/crawler
echo "✅ 目錄創建完成"

# 檢查.env文件
echo "📋 檢查環境配置..."
if [ ! -f ".env" ]; then
    echo "⚠️  .env文件不存在，從模板創建..."
    cp .env.template .env
    echo "✅ .env文件已創建，請編輯其中的Line Bot配置"
else
    echo "✅ .env文件存在"
fi

# 檢查MongoDB配置
echo "🗄️  檢查MongoDB配置..."
if grep -q "MongoDBURI=mongodb://admin:password123@mongodb:27017/ptt?authSource=admin" .env; then
    echo "✅ MongoDB URI配置正確"
else
    echo "⚠️  MongoDB URI可能需要調整"
    echo "建議的配置："
    echo "MongoDBURI=mongodb://admin:password123@mongodb:27017/ptt?authSource=admin"
    echo "MongoDBSSL=false"
fi

# 停止現有服務
echo "🛑 停止現有服務..."
docker-compose down

# 清理MongoDB數據（可選）
read -p "是否要清理MongoDB數據重新開始？(y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🗑️  清理MongoDB數據..."
    sudo rm -rf data/mongo/data/*
    echo "✅ MongoDB數據已清理"
fi

# 啟動MongoDB服務
echo "🚀 啟動MongoDB服務..."
docker-compose up -d mongodb

# 等待MongoDB啟動
echo "⏳ 等待MongoDB啟動..."
timeout=60
counter=0
while [ $counter -lt $timeout ]; do
    if docker-compose exec -T mongodb mongosh --eval "db.adminCommand('ping')" --quiet >/dev/null 2>&1; then
        echo "✅ MongoDB已啟動"
        break
    fi
    echo "等待MongoDB啟動... ($counter/$timeout)"
    sleep 2
    counter=$((counter + 2))
done

if [ $counter -ge $timeout ]; then
    echo "❌ MongoDB啟動超時"
    echo "請檢查Docker和MongoDB配置"
    exit 1
fi

# 測試MongoDB連接
echo "🔍 測試MongoDB連接..."
if docker-compose exec -T mongodb mongosh --username admin --password password123 --authenticationDatabase admin --eval "db.adminCommand('ping')" --quiet >/dev/null 2>&1; then
    echo "✅ MongoDB認證連接成功"
else
    echo "❌ MongoDB認證連接失敗"
    echo "請檢查用戶名和密碼配置"
fi

# 啟動爬蟲服務
echo "🕷️  啟動爬蟲服務..."
docker-compose up -d crawler

# 等待爬蟲服務啟動
echo "⏳ 等待爬蟲服務啟動..."
sleep 10

# 啟動應用程式
echo "🚀 啟動應用程式..."
docker-compose up -d app

# 等待應用程式啟動
echo "⏳ 等待應用程式啟動..."
sleep 15

# 檢查服務狀態
echo "📊 檢查服務狀態..."
docker-compose ps

# 測試應用程式健康檢查
echo "🏥 測試應用程式健康檢查..."
if curl -f http://localhost:5050/health >/dev/null 2>&1; then
    echo "✅ 應用程式健康檢查通過"
else
    echo "❌ 應用程式健康檢查失敗"
    echo "查看應用程式日誌："
    docker-compose logs --tail=20 app
fi

echo ""
echo "🎉 修復完成！"
echo ""
echo "📝 日誌位置："
echo "   - 應用程式日誌: ./logs/pttset.log"
echo "   - Docker日誌: docker-compose logs app"
echo "   - MongoDB日誌: docker-compose logs mongodb"
echo "   - 爬蟲日誌: docker-compose logs crawler"
echo ""
echo "🔧 常用指令："
echo "   docker-compose logs app     # 查看應用程式日誌"
echo "   docker-compose logs mongodb # 查看MongoDB日誌"
echo "   docker-compose ps           # 查看服務狀態"
echo "   make admin                  # 打開管理面板"
echo ""
echo "🌐 訪問地址："
echo "   健康檢查: http://localhost:5050/health"
echo "   管理面板: http://localhost:5050/admin"