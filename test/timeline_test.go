package test

import (
	"testing"

	"bergo/utils"
)

func TestTimelineLoadStore(t *testing.T) {
	// Create a test timeline
	timeline := &utils.Timeline{
		MaxId:     0,
		SessionId: "test_timeline_session",
		Items:     make([]*utils.TimelineItem, 0),
		Branch:    "test_branch",
	}

	// Add some test items
	timeline.AddUserInput(&utils.Query{
		UserInput: "test user input",
		Mode:      "test_mode",
	})

	timeline.AddLLMResponse("test response content", "test reasoning", "test rendered", nil, "")

	// Test Store method
	timeline.Store()

	// Test Load method
	loadedTimeline := &utils.Timeline{
		SessionId: "test_timeline_session",
		Branch:    "test_branch",
	}

	loadedTimeline.Load()

	// Verify the loaded data
	if loadedTimeline.MaxId != timeline.MaxId {
		t.Errorf("MaxId mismatch: expected %d, got %d", timeline.MaxId, loadedTimeline.MaxId)
	}

	if loadedTimeline.SessionId != timeline.SessionId {
		t.Errorf("SessionId mismatch: expected %s, got %s", timeline.SessionId, loadedTimeline.SessionId)
	}

	if loadedTimeline.Branch != timeline.Branch {
		t.Errorf("Branch mismatch: expected %s, got %s", timeline.Branch, loadedTimeline.Branch)
	}

	if len(loadedTimeline.Items) != len(timeline.Items) {
		t.Errorf("Items count mismatch: expected %d, got %d", len(timeline.Items), len(loadedTimeline.Items))
	}

	// Verify first item (UserInput)
	if len(loadedTimeline.Items) > 0 {
		firstItem := loadedTimeline.Items[0]
		if firstItem.Type != utils.TL_UserInput {
			t.Errorf("First item type mismatch: expected %s, got %s", utils.TL_UserInput, firstItem.Type)
		}

		// Verify Query data
		if query, ok := firstItem.Data.(*utils.Query); ok {
			if query.UserInput != "test user input" {
				t.Errorf("Query UserInput mismatch: expected 'test user input', got '%s'", query.UserInput)
			}
			if query.Mode != "test_mode" {
				t.Errorf("Query Mode mismatch: expected 'test_mode', got '%s'", query.Mode)
			}
		} else {
			t.Error("First item data is not a Query")
		}
	}

	// Verify second item (LLMResponse)
	if len(loadedTimeline.Items) > 1 {
		secondItem := loadedTimeline.Items[1]
		if secondItem.Type != utils.TL_LLMResponse {
			t.Errorf("Second item type mismatch: expected %s, got %s", utils.TL_LLMResponse, secondItem.Type)
		}

		// Verify LLMResponse data
		if response, ok := secondItem.Data.(*utils.LLMResponseItem); ok {
			if response.Content != "test response content" {
				t.Errorf("Response Content mismatch: expected 'test response content', got '%s'", response.Content)
			}
			if response.ReasoningContent != "test reasoning" {
				t.Errorf("Response ReasoningContent mismatch: expected 'test reasoning', got '%s'", response.ReasoningContent)
			}
			if response.RenderedContent != "test rendered" {
				t.Errorf("Response RenderedContent mismatch: expected 'test rendered', got '%s'", response.RenderedContent)
			}
		} else {
			t.Error("Second item data is not an LLMResponseItem")
		}
	}

	t.Log("Timeline Load and Store test completed successfully")
}
