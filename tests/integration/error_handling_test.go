//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"
)

// ER5: Unicode content handling
func TestError_UnicodeContent(t *testing.T) {
	// Navigate to a local page with Unicode
	navigate(t, fixtureURL(t, "/unicode"))

	// Test snapshot
	code, snapBody := httpGet(t, "/snapshot?tabId="+currentTabID)
	if code != 200 {
		t.Errorf("snapshot failed with %d", code)
	}

	// Verify response is valid JSON
	var snapData map[string]any
	if err := json.Unmarshal(snapBody, &snapData); err != nil {
		t.Errorf("snapshot response is not valid JSON: %v", err)
	}

	// Test text
	code, textBody := httpGet(t, "/text?tabId="+currentTabID)
	if code != 200 {
		t.Errorf("text failed with %d", code)
	}

	// Verify response is valid JSON
	var textData map[string]any
	if err := json.Unmarshal(textBody, &textData); err != nil {
		t.Errorf("text response is not valid JSON: %v", err)
	}
}

// ER6: Empty page handling
func TestError_EmptyPage(t *testing.T) {
	// Navigate to about:blank (empty page)
	navigate(t, "about:blank")

	// Test snapshot
	code, snapBody := httpGet(t, "/snapshot?tabId="+currentTabID)
	if code != 200 {
		t.Errorf("snapshot on empty page failed with %d", code)
	}

	// Verify response is valid JSON
	var snapData map[string]any
	if err := json.Unmarshal(snapBody, &snapData); err != nil {
		t.Errorf("snapshot response is not valid JSON: %v", err)
	}

	// Test text
	code, textBody := httpGet(t, "/text?tabId="+currentTabID)
	if code != 200 {
		t.Errorf("text on empty page failed with %d", code)
	}

	// Verify response is valid JSON
	var textData map[string]any
	if err := json.Unmarshal(textBody, &textData); err != nil {
		t.Errorf("text response is not valid JSON: %v", err)
	}
}

// ER3: Binary page (PDF) handling
func TestError_BinaryPage(t *testing.T) {
	// Navigate to a local PDF
	navigate(t, fixtureURL(t, "/binary.pdf"))

	// Verify navigation completes
	t.Logf("navigation to PDF completed successfully")

	// Test snapshot on PDF content
	code, snapBody := httpGet(t, "/snapshot?tabId="+currentTabID)
	if code == 200 {
		var snapData map[string]any
		if err := json.Unmarshal(snapBody, &snapData); err != nil {
			t.Errorf("snapshot response is not valid JSON: %v", err)
		}
	} else if code >= 400 && code < 500 {
		t.Logf("snapshot returned %d (acceptable for PDF)", code)
	} else {
		t.Errorf("snapshot failed with unexpected code %d", code)
	}

	// Test text extraction on PDF
	code, textBody := httpGet(t, "/text?tabId="+currentTabID)
	if code == 200 {
		var textData map[string]any
		if err := json.Unmarshal(textBody, &textData); err != nil {
			t.Errorf("text response is not valid JSON: %v", err)
		}
	} else if code >= 400 && code < 500 {
		t.Logf("text returned %d (acceptable for PDF)", code)
	} else {
		t.Errorf("text failed with unexpected code %d", code)
	}
}

// ER4: Rapid navigation stress test
func TestError_RapidNavigate(t *testing.T) {
	urls := []string{
		fixtureURL(t, "/page1"),
		fixtureURL(t, "/page2"),
		fixtureURL(t, "/slow"),
		fixtureURL(t, "/page1"),
		fixtureURL(t, "/page2"),
	}

	// Use a WaitGroup to ensure all navigations complete
	var wg sync.WaitGroup
	errors := make([]error, len(urls))
	var mu sync.Mutex

	startTime := time.Now()

	// Rapidly navigate to all URLs
	for i, url := range urls {
		wg.Add(1)
		go func(idx int, u string) {
			defer wg.Done()
			// Use local navigate logic to avoid tabId closure in loop
			code, _ := httpPost(t, "/navigate", map[string]string{"url": u})
			if code != 200 {
				mu.Lock()
				errors[idx] = fmt.Errorf("navigate to %s failed with code %d", u, code)
				mu.Unlock()
			}
		}(i, url)
	}

	// Wait for all navigations to complete
	wg.Wait()
	elapsed := time.Since(startTime)

	// Check that navigations completed quickly
	if elapsed > 5*time.Second {
		t.Logf("rapid navigation took %v", elapsed)
	}

	// Verify no critical errors occurred
	for _, err := range errors {
		if err != nil {
			t.Errorf("%v", err)
		}
	}

	// Verify server didn't crash
	time.Sleep(500 * time.Millisecond)

	code, snapBody := httpGet(t, "/snapshot")
	if code == 200 {
		var snapData map[string]any
		if err := json.Unmarshal(snapBody, &snapData); err != nil {
			t.Errorf("snapshot response is not valid JSON: %v", err)
		}
	} else {
		t.Errorf("snapshot failed with code %d", code)
	}

	t.Logf("rapid navigation test completed: 5 navigations in %v", elapsed)
}
