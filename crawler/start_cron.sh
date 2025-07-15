#!/bin/bash

echo "Starting cron daemon..."

# 啟動 cron 服務
service cron start

# 立即執行一次爬蟲（可選）
if [ "$RUN_IMMEDIATELY" = "true" ]; then
    echo "Running crawler immediately..."
    /app/run_crawler.sh
fi

# 保持容器運行並顯示日誌
echo "Cron started. Crawler will run every hour."
echo "Logs will be shown below:"
tail -f /var/log/crawler.log