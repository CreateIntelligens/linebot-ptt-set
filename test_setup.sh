#!/bin/bash

echo "🚀 Testing PTT SET Line Bot with Crawler Setup"
echo "=================================================="

# 檢查必要文件
echo "📋 Checking required files..."
required_files=(
    "docker-compose.yml"
    "Dockerfile"
    "crawler/Dockerfile"
    "crawler/run_crawler.sh"
    "crawler/start_cron.sh"
    ".env.template"
)

for file in "${required_files[@]}"; do
    if [ -f "$file" ]; then
        echo "✅ $file exists"
    else
        echo "❌ $file missing"
        exit 1
    fi
done

# 檢查 .env 文件
if [ ! -f ".env" ]; then
    echo "⚠️  .env file not found, copying from template..."
    cp .env.template .env
    echo "📝 Please edit .env file with your Line Bot credentials"
fi

# 創建必要目錄
echo "📁 Creating data directories..."
mkdir -p data/mongo/data data/crawler

# 檢查 Docker 和 Docker Compose
echo "🐳 Checking Docker..."
if ! command -v docker &> /dev/null; then
    echo "❌ Docker not found. Please install Docker first."
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose not found. Please install Docker Compose first."
    exit 1
fi

echo "✅ Docker and Docker Compose are available"

# 構建映像
echo "🔨 Building Docker images..."
make build

if [ $? -ne 0 ]; then
    echo "❌ Failed to build Docker images"
    exit 1
fi

echo "✅ Docker images built successfully"

# 啟動服務
echo "🚀 Starting services..."
make dev

if [ $? -ne 0 ]; then
    echo "❌ Failed to start services"
    exit 1
fi

# 等待服務啟動
echo "⏳ Waiting for services to start..."
sleep 10

# 檢查服務狀態
echo "📊 Checking service status..."
make status

# 檢查 MongoDB 連接
echo "🔍 Testing MongoDB connection..."
timeout 30 bash -c 'until docker-compose exec -T mongodb mongo --eval "db.adminCommand(\"ismaster\")" &>/dev/null; do sleep 1; done'

if [ $? -eq 0 ]; then
    echo "✅ MongoDB is running and accessible"
else
    echo "❌ MongoDB connection failed"
    make down
    exit 1
fi

# 檢查爬蟲服務
echo "🕷️  Checking crawler service..."
if docker-compose ps crawler | grep -q "Up"; then
    echo "✅ Crawler service is running"
else
    echo "❌ Crawler service is not running"
    make down
    exit 1
fi

# 檢查應用程式服務
echo "🤖 Checking Line Bot application..."
if docker-compose ps app | grep -q "Up"; then
    echo "✅ Line Bot application is running"
else
    echo "❌ Line Bot application is not running"
    make down
    exit 1
fi

# 測試健康檢查端點
echo "🏥 Testing health endpoint..."
sleep 5
if curl -f http://localhost:5050/health &>/dev/null; then
    echo "✅ Health endpoint is responding"
else
    echo "⚠️  Health endpoint not responding (this is normal if Line Bot credentials are not set)"
fi

echo ""
echo "🎉 Setup test completed!"
echo ""
echo "📋 Next steps:"
echo "1. Edit .env file with your Line Bot credentials"
echo "2. Restart services: make down && make dev"
echo "3. Monitor crawler logs: make crawler-logs"
echo "4. Check service status: make status"
echo ""
echo "📚 Useful commands:"
echo "- View all logs: make dev-logs"
echo "- View crawler logs: make crawler-logs"
echo "- Manual crawler run: make crawler-run"
echo "- Stop services: make down"
echo "- Clean everything: make clean"
echo ""
echo "🔗 The crawler will automatically run every hour to fetch new PTT SET posts."