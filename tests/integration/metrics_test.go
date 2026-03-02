//go:build integration

package integration

import (
	"encoding/json"
	"testing"
)

// M1: Get metrics endpoint returns aggregated memory
func TestMetrics_Basic(t *testing.T) {
	// Navigate to a page first to ensure we have a tab with content
	navigate(t, "https://example.com")
	defer closeCurrentTab(t)

	code, body := httpGet(t, "/metrics")
	if code != 200 {
		t.Fatalf("expected 200, got %d: %s", code, string(body))
	}

	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("invalid json: %v", err)
	}

	// Should have metrics field
	if _, ok := m["metrics"]; !ok {
		t.Error("expected metrics field in response")
	}

	// Should have memory field with aggregated stats
	mem, ok := m["memory"].(map[string]any)
	if !ok {
		t.Fatal("expected memory field in response")
	}

	// Check expected memory fields exist
	fields := []string{"jsHeapUsedMB", "jsHeapTotalMB", "documents", "nodes"}
	for _, f := range fields {
		if _, ok := mem[f]; !ok {
			t.Errorf("expected %s in memory response", f)
		}
	}
}

// M2: Per-tab metrics
func TestMetrics_PerTab(t *testing.T) {
	navigate(t, "https://example.com")
	defer closeCurrentTab(t)

	code, body := httpGet(t, "/tabs/"+currentTabID+"/metrics")
	if code != 200 {
		t.Fatalf("expected 200, got %d: %s", code, string(body))
	}

	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("invalid json: %v", err)
	}

	// Should have jsHeapUsedMB > 0 for a loaded page
	heap, ok := m["jsHeapUsedMB"].(float64)
	if !ok {
		t.Fatal("expected jsHeapUsedMB in response")
	}
	if heap <= 0 {
		t.Errorf("expected jsHeapUsedMB > 0, got %f", heap)
	}
}

// M3: Invalid tab ID returns error
func TestMetrics_InvalidTab(t *testing.T) {
	code, _ := httpGet(t, "/tabs/invalid_tab_id/metrics")
	if code != 500 && code != 404 {
		t.Errorf("expected 500 or 404 for invalid tab, got %d", code)
	}
}

// M4: Memory increases with DOM nodes
func TestMetrics_MemoryGrowth(t *testing.T) {
	navigate(t, "https://example.com")
	defer closeCurrentTab(t)

	// Get initial metrics
	_, body1 := httpGet(t, "/tabs/"+currentTabID+"/metrics")
	var m1 map[string]any
	json.Unmarshal(body1, &m1)
	initialNodes := m1["nodes"].(float64)

	// Inject some DOM nodes
	httpPost(t, "/evaluate", map[string]any{
		"tabId":      currentTabID,
		"expression": "for(let i=0;i<100;i++){document.body.appendChild(document.createElement('div'))}",
	})

	// Get metrics again
	_, body2 := httpGet(t, "/tabs/"+currentTabID+"/metrics")
	var m2 map[string]any
	json.Unmarshal(body2, &m2)
	finalNodes := m2["nodes"].(float64)

	if finalNodes <= initialNodes {
		t.Errorf("expected nodes to increase: initial=%f, final=%f", initialNodes, finalNodes)
	}
}
