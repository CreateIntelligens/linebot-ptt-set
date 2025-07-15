FROM python:3.9-slim

# 安裝必要的系統套件
RUN apt-get update && apt-get install -y \
    git \
    wget \
    gnupg \
    cron \
    && rm -rf /var/lib/apt/lists/*

# 安裝 MongoDB tools
RUN wget -qO - https://www.mongodb.org/static/pgp/server-6.0.asc | apt-key add - \
    && echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/ubuntu focal/mongodb-org/6.0 multiverse" | tee /etc/apt/sources.list.d/mongodb-org-6.0.list \
    && apt-get update \
    && apt-get install -y mongodb-database-tools mongodb-mongosh \
    && rm -rf /var/lib/apt/lists/*

# 設定工作目錄
WORKDIR /app

# Clone PTT Web Crawler
RUN git clone https://github.com/mong0520/ptt-web-crawler.git .

# 安裝 Python 依賴
RUN pip install -r requirements.txt

# 複製爬蟲腳本
COPY run_crawler.sh /app/run_crawler.sh
COPY start_cron.sh /app/start_cron.sh

# 給腳本執行權限
RUN chmod +x /app/run_crawler.sh /app/start_cron.sh

# 創建 cron job
RUN echo "0 * * * * /app/run_crawler.sh >> /var/log/crawler.log 2>&1" | crontab -

# 創建日誌文件
RUN touch /var/log/crawler.log

CMD ["/app/start_cron.sh"]