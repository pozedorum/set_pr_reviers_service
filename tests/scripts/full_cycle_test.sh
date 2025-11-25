#!/bin/bash

echo "🧪 Starting comprehensive test scenario..."

# 1. Create frontend team
echo "1. Creating frontend team..."
curl -s -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "frontend",
    "members": [
      {"user_id": "fe1", "username": "Frontend Alice", "is_active": true},
      {"user_id": "fe2", "username": "Frontend Bob", "is_active": true},
      {"user_id": "fe3", "username": "Frontend Charlie", "is_active": true}
    ]
  }' | jq .

# 2. Create PR from frontend team
echo "2. Creating PR..."
PR_RESPONSE=$(curl -s -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-frontend-001",
    "pull_request_name": "Add responsive design",
    "author_id": "fe1"
  }')
echo $PR_RESPONSE | jq .

# 3. Get assigned reviewers
REVIEWER=$(echo $PR_RESPONSE | jq -r '.pr.assigned_reviewers[0]')
echo "3. Assigned reviewer: $REVIEWER"

# 4. Get user reviews
echo "4. Getting user reviews..."
curl -s "http://localhost:8080/users/getReview?user_id=$REVIEWER" | jq .

# 5. Reassign reviewer
echo "5. Reassigning reviewer..."
curl -s -X POST http://localhost:8080/pullRequest/reassign \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-frontend-001",
    "old_reviewer_id": "'$REVIEWER'"
  }' | jq .

# 6. Merge PR
echo "6. Merging PR..."
curl -s -X POST http://localhost:8080/pullRequest/merge \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-frontend-001"
  }' | jq .

echo "✅ Test scenario completed!"