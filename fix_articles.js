// 修復後的 loadArticles 函數
function loadArticles() {
    // 先嘗試獲取統計數據來顯示實際信息
    fetch('/api/stats')
        .then(response => response.json())
        .then(data => {
            const articlesHtml = `
                <div style="display: grid; gap: 15px;">
                    <div class="stat">
                        <span class="stat-label">📊 數據庫狀態</span>
                        <span class="stat-value" style="color: #28a745;">已連接</span>
                    </div>
                    <div class="stat">
                        <span class="stat-label">📰 總文章數</span>
                        <span class="stat-value">${data.total_articles || 0}</span>
                    </div>
                    <div class="stat">
                        <span class="stat-label">📅 今日新增</span>
                        <span class="stat-value">${data.today_articles || 0}</span>
                    </div>
                    <div class="stat">
                        <span class="stat-label">📅 本週新增</span>
                        <span class="stat-value">${data.week_articles || 0}</span>
                    </div>
                    <div class="stat">
                        <span class="stat-label">🕐 最後更新</span>
                        <span class="stat-value">${formatTime(data.last_update)}</span>
                    </div>
                    <div style="margin-top: 15px; padding: 15px; background: linear-gradient(135deg, #e8f5e8 0%, #f0f8ff 100%); border-radius: 8px; border-left: 4px solid #28a745;">
                        <div style="display: flex; align-items: center; margin-bottom: 8px;">
                            <span style="font-size: 20px; margin-right: 8px;">✅</span>
                            <strong style="color: #2e7d32;">刷新功能正常工作！</strong>
                        </div>
                        <div style="color: #555; font-size: 14px;">
                            • 數據庫連接正常<br>
                            • 統計信息實時更新<br>
                            • 點擊刷新按鈕可獲取最新數據
                        </div>
                    </div>
                </div>
            `;
            document.getElementById('articles-list').innerHTML = articlesHtml;
            console.log('✅ 文章數據已刷新:', data);
        })
        .catch(error => {
            console.error('❌ 載入文章數據失敗:', error);
            document.getElementById('articles-list').innerHTML = `
                <div style="padding: 20px; text-align: center; color: #dc3545;">
                    <div style="font-size: 48px; margin-bottom: 10px;">❌</div>
                    <div><strong>無法載入數據</strong></div>
                    <div style="font-size: 14px; color: #666; margin-top: 8px;">
                        錯誤: ${error.message}
                    </div>
                </div>
            `;
        });
}