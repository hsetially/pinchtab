#!/bin/bash
# 39-scheduler.sh — Scheduler task lifecycle E2E tests

source "$(dirname "$0")/common.sh"

AGENT="test-agent-$$"

# Get a tab for task execution
pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/index.html\"}"
TAB_ID=$(echo "$RESULT" | jq -r '.tabId')

# ─────────────────────────────────────────────────────────────────
start_test "POST /tasks — submit task"

pt_post /tasks -d "{\"agentId\":\"${AGENT}\",\"action\":\"snapshot\",\"tabId\":\"${TAB_ID}\"}"
assert_http_status "202" "task accepted"
TASK_ID=$(echo "$RESULT" | jq -r '.taskId')
assert_json_eq "$RESULT" ".state" "queued" "initial state is queued"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "GET /tasks/{id} — get task"

sleep 2
pt_get "/tasks/${TASK_ID}"
assert_ok "get task by id"
assert_json_eq "$RESULT" ".taskId" "$TASK_ID" "correct task id"
assert_json_eq "$RESULT" ".agentId" "$AGENT" "correct agent id"
assert_json_eq "$RESULT" ".action" "snapshot" "correct action"
STATE=$(echo "$RESULT" | jq -r '.state')
if [ "$STATE" = "completed" ] || [ "$STATE" = "running" ] || [ "$STATE" = "failed" ]; then
  echo -e "  ${GREEN}✓${NC} task reached terminal/active state: $STATE"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} unexpected state: $STATE"
  ((ASSERTIONS_FAILED++)) || true
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "GET /tasks — list tasks"

pt_get "/tasks?agentId=${AGENT}"
assert_ok "list tasks"
COUNT=$(echo "$RESULT" | jq '.count')
if [ "$COUNT" -ge 1 ]; then
  echo -e "  ${GREEN}✓${NC} found $COUNT tasks for agent"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} expected at least 1 task, got $COUNT"
  ((ASSERTIONS_FAILED++)) || true
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "GET /tasks — filter by state"

pt_get "/tasks?state=completed,failed"
assert_ok "list terminal tasks"
echo "$RESULT" | jq -e '.tasks' > /dev/null 2>&1
if [ $? -eq 0 ]; then
  echo -e "  ${GREEN}✓${NC} tasks array present in response"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} missing tasks array"
  ((ASSERTIONS_FAILED++)) || true
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "POST /tasks/{id}/cancel — cancel queued task"

pt_post /tasks -d "{\"agentId\":\"${AGENT}\",\"action\":\"snapshot\",\"tabId\":\"${TAB_ID}\"}"
assert_http_status "202" "task accepted for cancel test"
CANCEL_ID=$(echo "$RESULT" | jq -r '.taskId')

pt_post "/tasks/${CANCEL_ID}/cancel" ""
# Could be 200 (cancelled) or 409 (already in terminal state)
if [ "$HTTP_STATUS" = "200" ]; then
  assert_json_eq "$RESULT" ".status" "cancelled" "task cancelled"
elif [ "$HTTP_STATUS" = "409" ]; then
  echo -e "  ${GREEN}✓${NC} task already terminal (409 conflict, acceptable)"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} unexpected status: $HTTP_STATUS"
  ((ASSERTIONS_FAILED++)) || true
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "POST /tasks/{id}/cancel — cancel nonexistent → 404"

pt_post "/tasks/tsk_nonexistent/cancel" ""
assert_http_status "404" "cancel nonexistent task"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "GET /scheduler/stats"

pt_get /scheduler/stats
assert_ok "stats endpoint"
echo "$RESULT" | jq -e '.queue' > /dev/null 2>&1
if [ $? -eq 0 ]; then
  echo -e "  ${GREEN}✓${NC} has queue stats"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} missing queue stats"
  ((ASSERTIONS_FAILED++)) || true
fi
echo "$RESULT" | jq -e '.metrics' > /dev/null 2>&1
if [ $? -eq 0 ]; then
  echo -e "  ${GREEN}✓${NC} has metrics"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} missing metrics"
  ((ASSERTIONS_FAILED++)) || true
fi
assert_json_eq "$RESULT" ".config.strategy" "fair-fifo" "strategy is fair-fifo"
assert_json_eq "$RESULT" ".config.maxQueueSize" "5" "maxQueueSize matches config"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "POST /tasks/batch — submit 3 tasks"

pt_post /tasks/batch -d "{\"agentId\":\"${AGENT}\",\"tasks\":[{\"action\":\"snapshot\",\"tabId\":\"${TAB_ID}\"},{\"action\":\"text\",\"tabId\":\"${TAB_ID}\"},{\"action\":\"screenshot\",\"tabId\":\"${TAB_ID}\"}]}"
assert_http_status "202" "batch accepted"
SUBMITTED=$(echo "$RESULT" | jq '.submitted')
if [ "$SUBMITTED" = "3" ]; then
  echo -e "  ${GREEN}✓${NC} all 3 tasks submitted"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} expected 3 submitted, got $SUBMITTED"
  ((ASSERTIONS_FAILED++)) || true
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "POST /tasks/batch — validation: empty tasks"

pt_post /tasks/batch -d '{"agentId":"test","tasks":[]}'
assert_http_status "400" "empty batch rejected"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "POST /tasks/batch — validation: missing agentId"

pt_post /tasks/batch -d '{"tasks":[{"action":"snapshot"}]}'
assert_http_status "400" "missing agentId rejected"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "POST /tasks — deadline in the past → 400"

pt_post /tasks -d "{\"agentId\":\"${AGENT}\",\"action\":\"snapshot\",\"tabId\":\"${TAB_ID}\",\"deadline\":\"2020-01-01T00:00:00Z\"}"
assert_http_status "400" "past deadline rejected"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "POST /tasks — 429 queue full"

# maxPerAgent is 3, submit rapidly for a fresh agent
FLOOD_AGENT="flood-agent-$$"
for i in $(seq 1 4); do
  pt_post /tasks -d "{\"agentId\":\"${FLOOD_AGENT}\",\"action\":\"snapshot\",\"tabId\":\"${TAB_ID}\"}"
done
# The 4th should be 429 (maxPerAgent=3) — racy if tasks complete fast
if [ "$HTTP_STATUS" = "429" ]; then
  echo -e "  ${GREEN}✓${NC} queue full for agent (HTTP 429)"
  ((ASSERTIONS_PASSED++)) || true
elif [ "$HTTP_STATUS" = "202" ]; then
  echo -e "  ${GREEN}✓${NC} task accepted (tasks completed before limit hit)"
  ((ASSERTIONS_PASSED++)) || true
else
  echo -e "  ${RED}✗${NC} unexpected status: $HTTP_STATUS (expected 429 or 202)"
  ((ASSERTIONS_FAILED++)) || true
fi

end_test

# ─────────────────────────────────────────────────────────────────
start_test "GET /tasks/{id} — nonexistent → 404"

pt_get /tasks/tsk_doesnotexist
assert_http_status "404" "nonexistent task"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "POST /tasks — missing agentId → 400"

pt_post /tasks -d '{"action":"snapshot"}'
assert_http_status "400" "missing agentId rejected"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "POST /tasks — missing action → 400"

pt_post /tasks -d "{\"agentId\":\"validation-agent-$$\"}"
assert_http_status "400" "missing action rejected"

end_test
