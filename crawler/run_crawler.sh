#!/bin/bash

# 設定環境變數和 PATH（cron 環境需要）
export PATH="/usr/local/bin:/usr/bin:/bin:$PATH"
export MONGODB_HOST=${MONGODB_HOST:-mongo}
export MONGODB_PORT=${MONGODB_PORT:-27017}
export MONGODB_DB=${MONGODB_DB:-ptt}
export PAGE_OFFSET=${PAGE_OFFSET:-10}

# 等待 MongoDB 啟動
echo "$(date): Waiting for MongoDB to be ready..."
until mongosh --host $MONGODB_HOST:$MONGODB_PORT --eval "db.adminCommand('ping')" --quiet 2>/dev/null; do
    echo "$(date): MongoDB is unavailable - sleeping"
    sleep 5
done
echo "$(date): MongoDB is ready!"

cd /app

# 備份舊的 JSON 文件
if [ -f "SET.json" ]; then
    mv SET.json SET.json.old
    echo "$(date): Backed up existing SET.json"
fi

# 執行爬蟲
echo "$(date): Starting PTT SET crawler with offset $PAGE_OFFSET..."
python3 -m PttWebCrawler.crawler -b SET -o $PAGE_OFFSET

# 檢查是否成功產生 JSON 文件
if [ ! -f "SET.json" ]; then
    echo "$(date): ERROR - SET.json not generated!"
    exit 1
fi

# 檢查 JSON 文件是否為空
if [ ! -s "SET.json" ]; then
    echo "$(date): ERROR - SET.json is empty!"
    exit 1
fi

# 匯入資料到 MongoDB
echo "$(date): Importing data to MongoDB..."
mongoimport --host $MONGODB_HOST:$MONGODB_PORT \
           --db $MONGODB_DB \
           --collection set \
           --type json \
           --file SET.json \
           --jsonArray \
           --mode merge \
           --upsertFields article_id

if [ $? -eq 0 ]; then
    echo "$(date): Successfully imported data to MongoDB"
    
    # 創建唯一索引（如果不存在）
    mongosh --host $MONGODB_HOST:$MONGODB_PORT \
            --username $MONGODB_USERNAME \
            --password $MONGODB_PASSWORD \
            --authenticationDatabase admin \
            $MONGODB_DB \
            --eval 'db.set.createIndex( { "article_id": 1 }, { unique: true } )' \
            --quiet 2>/dev/null || true
    
    # 清理舊的備份文件（保留最近3個）
    ls -t SET.json.old* 2>/dev/null | tail -n +4 | xargs rm -f 2>/dev/null || true
    
else
    echo "$(date): ERROR - Failed to import data to MongoDB"
    exit 1
fi

echo "$(date): Crawler job completed successfully"
