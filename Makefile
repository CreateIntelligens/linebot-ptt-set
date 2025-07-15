APP=linebot-ptt-set
NAMESPACE := mong0520

build:
	docker build -t ${NAMESPACE}/${APP} .
	docker build -t ${NAMESPACE}/${APP}-crawler ./crawler

dev:
	@echo "Creating data directories..."
	@mkdir -p data/mongo/data data/crawler
	@echo "Starting services..."
	docker-compose up -d

dev-logs:
	docker-compose logs -f

down:
	docker-compose down

clean:
	docker-compose down -v
	docker system prune -f

crawler-logs:
	docker-compose logs -f crawler

crawler-run:
	docker-compose exec crawler /app/run_crawler.sh

crawler-test:
	docker-compose exec crawler /app/test_crawler.sh

push:
	@docker tag ${APP} ${NAMESPACE}/${APP}
	@docker push ${NAMESPACE}/${APP}
	@docker tag ${NAMESPACE}/${APP}-crawler ${NAMESPACE}/${APP}-crawler
	@docker push ${NAMESPACE}/${APP}-crawler
	@heroku container:push web

release:
	@heroku container:release web

status:
	docker-compose ps
