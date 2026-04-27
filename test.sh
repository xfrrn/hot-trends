#!/bin/bash

echo "🚀 Testing Hot Trends Service..."
echo ""

# Start the server in background
echo "Starting server..."
cd "$(dirname "$0")"
./bin/server &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo ""
echo "📊 Running API tests..."
echo ""

# Test 1: Health check
echo "1. Testing health endpoint..."
curl -s http://localhost:6000/health | jq '.'
echo ""

# Test 2: List platforms
echo "2. Testing platforms list..."
curl -s http://localhost:6000/api/v1/platforms | jq '.platforms[] | {id, name, type}'
echo ""

# Test 3: Get Weibo trends
echo "3. Testing Weibo trends (limit 3)..."
curl -s "http://localhost:6000/api/v1/trends/weibo?limit=3" | jq '{platform, items: .items[0:3] | map({title, hot_value}), cached}'
echo ""

# Test 4: Batch request
echo "4. Testing batch request (Weibo + Zhihu)..."
curl -s -X POST http://localhost:6000/api/v1/trends/batch \
  -H "Content-Type: application/json" \
  -d '{"platforms":["weibo","zhihu"],"limit":2}' | jq '{total_platforms, successful, failed, results: .results | map({platform, item_count: .items | length, cached})}'
echo ""

# Cleanup
echo "Stopping server..."
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null

echo ""
echo "✅ Tests completed!"
