#!/bin/bash

echo "🧪 Testing PTT Crawler Setup"
echo "============================="

# 設定測試環境變數
export MONGODB_HOST=${MONGODB_HOST:-mongodb}
export MONGODB_PORT=${MONGODB_PORT:-27017}
export MONGODB_DB=${MONGODB_DB:-ptt}
export MONGODB_USERNAME=${MONGODB_USERNAME:-admin}
export MONGODB_PASSWORD=${MONGODB_PASSWORD:-password123}
export PAGE_OFFSET=${PAGE_OFFSET:-1}

echo "📋 Configuration:"
echo "  MongoDB Host: $MONGODB_HOST:$MONGODB_PORT"
echo "  Database: $MONGODB_DB"
echo "  Page Offset: $PAGE_OFFSET"
echo ""

# 測試 MongoDB 連接
echo "🔍 Testing MongoDB connection..."
if mongosh --host $MONGODB_HOST:$MONGODB_PORT --username $MONGODB_USERNAME --password $MONGODB_PASSWORD --authenticationDatabase admin --eval "db.adminCommand('ping')" --quiet 2>/dev/null; then
    echo "✅ MongoDB connection successful"
else
    echo "❌ MongoDB connection failed"
    exit 1
fi

# 測試爬蟲腳本
echo "🕷️  Testing PTT crawler..."
if python3 -m PttWebCrawler.crawler -b SET -o 1; then
    echo "✅ Crawler executed successfully"
    
    # 檢查生成的文件
    if [ -f "SET.json" ]; then
        echo "✅ SET.json generated"
        
        # 檢查文件大小
        file_size=$(stat -c%s "SET.json")
        if [ $file_size -gt 100 ]; then
            echo "✅ SET.json has content ($file_size bytes)"
        else
            echo "⚠️  SET.json seems empty or too small"
        fi
        
        # 測試 JSON 格式
        if python3 -c "import json; json.load(open('SET.json'))" 2>/dev/null; then
            echo "✅ SET.json is valid JSON"
        else
            echo "❌ SET.json is not valid JSON"
        fi
    else
        echo "❌ SET.json not generated"
        exit 1
    fi
else
    echo "❌ Crawler execution failed"
    exit 1
fi

# 測試資料匯入
echo "📥 Testing data import..."
if mongoimport --host $MONGODB_HOST:$MONGODB_PORT \
               --username $MONGODB_USERNAME \
               --password $MONGODB_PASSWORD \
               --authenticationDatabase admin \
               --db $MONGODB_DB \
               --collection set \
               --type json \
               --file SET.json \
               --jsonArray \
               --mode merge \
               --upsertFields article_id; then
    echo "✅ Data import successful"
    
    # 檢查匯入的資料數量
    count=$(mongosh --host $MONGODB_HOST:$MONGODB_PORT --username $MONGODB_USERNAME --password $MONGODB_PASSWORD --authenticationDatabase admin $MONGODB_DB --eval "db.set.countDocuments()" --quiet 2>/dev/null | tail -1)
    echo "📊 Documents in database: $count"
    
    # 測試索引創建
    echo "🔍 Creating unique index..."
    if mongosh --host $MONGODB_HOST:$MONGODB_PORT --username $MONGODB_USERNAME --password $MONGODB_PASSWORD --authenticationDatabase admin $MONGODB_DB --eval 'db.set.createIndex( { "article_id": 1 }, { unique: true } )' --quiet 2>/dev/null; then
        echo "✅ Index created successfully"
    else
        echo "⚠️  Index creation failed (may already exist)"
    fi
else
    echo "❌ Data import failed"
    exit 1
fi

echo ""
echo "🎉 All tests passed! Crawler setup is working correctly."
echo ""
echo "📋 Next steps:"
echo "1. Build the crawler image: docker build -t crawler ./crawler"
echo "2. Start the full stack: make dev"
echo "3. Monitor crawler logs: make crawler-logs"