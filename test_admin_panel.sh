#!/bin/bash

echo "🔧 PTT SET Bot - 管理面板測試"
echo "================================"

# 檢查服務是否運行
echo "📋 檢查服務狀態..."
if ! docker-compose ps | grep -q "Up"; then
    echo "❌ 服務未運行，請先執行 'make dev'"
    exit 1
fi

echo "✅ 服務正在運行"

# 等待服務完全啟動
echo "⏳ 等待服務完全啟動..."
sleep 5

# 測試基本健康檢查
echo "🏥 測試健康檢查端點..."
if curl -f http://localhost:5050/health >/dev/null 2>&1; then
    echo "✅ 健康檢查端點正常"
else
    echo "❌ 健康檢查端點失敗"
    exit 1
fi

# 測試管理面板主頁
echo "🎛️ 測試管理面板主頁..."
if curl -f http://localhost:5050/admin >/dev/null 2>&1; then
    echo "✅ 管理面板主頁可訪問"
else
    echo "❌ 管理面板主頁無法訪問"
    exit 1
fi

# 測試API端點
echo "🔌 測試管理面板API..."

# 測試狀態API
if curl -f http://localhost:5050/admin/api/status >/dev/null 2>&1; then
    echo "✅ 狀態API正常"
else
    echo "❌ 狀態API失敗"
fi

# 測試文章API
if curl -f http://localhost:5050/admin/api/articles >/dev/null 2>&1; then
    echo "✅ 文章API正常"
else
    echo "❌ 文章API失敗"
fi

# 測試日誌API
if curl -f http://localhost:5050/admin/api/crawler/logs >/dev/null 2>&1; then
    echo "✅ 日誌API正常"
else
    echo "❌ 日誌API失敗"
fi

# 檢查MongoDB連接
echo "🗄️ 檢查MongoDB連接..."
if docker-compose exec -T mongodb mongosh --eval "db.adminCommand('ping')" --quiet >/dev/null 2>&1; then
    echo "✅ MongoDB連接正常"
    
    # 檢查數據
    echo "📊 檢查數據庫數據..."
    count=$(docker-compose exec -T mongodb mongosh ptt --username admin --password password123 --authenticationDatabase admin --eval "db.set.countDocuments()" --quiet 2>/dev/null | tail -1)
    if [[ "$count" =~ ^[0-9]+$ ]]; then
        echo "✅ 數據庫包含 $count 篇文章"
    else
        echo "⚠️ 無法獲取文章數量，可能需要先運行爬蟲"
    fi
else
    echo "❌ MongoDB連接失敗"
fi

# 檢查爬蟲容器
echo "🕷️ 檢查爬蟲狀態..."
if docker-compose ps crawler | grep -q "Up"; then
    echo "✅ 爬蟲容器正在運行"
    
    # 檢查爬蟲日誌
    echo "📝 最近的爬蟲日誌:"
    docker-compose logs --tail=5 crawler | sed 's/^/   /'
else
    echo "❌ 爬蟲容器未運行"
fi

echo ""
echo "🎉 測試完成！"
echo ""
echo "📱 管理面板訪問方式:"
echo "   🌐 網頁: http://localhost:5050/admin"
echo "   📊 狀態API: http://localhost:5050/admin/api/status"
echo "   📰 文章API: http://localhost:5050/admin/api/articles"
echo "   📝 日誌API: http://localhost:5050/admin/api/crawler/logs"
echo ""
echo "🛠️ 常用指令:"
echo "   make admin          # 打開管理面板"
echo "   make crawler-logs   # 查看爬蟲日誌"
echo "   make status         # 查看服務狀態"
echo "   make test-admin     # 測試管理面板"
echo ""
echo "💡 提示: 如果看到錯誤，請檢查:"
echo "   1. 所有服務是否正常運行 (make status)"
echo "   2. 端口5050是否被占用"
echo "   3. MongoDB是否正確啟動"
echo "   4. 爬蟲是否已執行並有數據"