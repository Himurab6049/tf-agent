package server_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestCancelTask_NotRunning(t *testing.T) {
	env := newTestEnv(t)
	taskID := submitTestTask(t, env)

	// Cancel immediately — task may be queued, not running yet.
	res := env.do("POST", "/v1/tasks/"+taskID+"/cancel", env.adminToken, nil)
	// Either 200 (cancelled) or 409 (not running yet) are valid.
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusConflict {
		t.Errorf("unexpected status %d", res.StatusCode)
	}
}

func TestCancelTask_NotFound(t *testing.T) {
	env := newTestEnv(t)
	res := env.do("POST", "/v1/tasks/nonexistent-id/cancel", env.adminToken, nil)
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("got %d, want 404", res.StatusCode)
	}
}

func TestCancelTask_Forbidden(t *testing.T) {
	env := newTestEnv(t)

	// Submit as admin, cancel as member — should be forbidden.
	taskID := submitTestTask(t, env)
	res := env.do("POST", "/v1/tasks/"+taskID+"/cancel", env.memberToken, nil)
	if res.StatusCode != http.StatusForbidden {
		t.Errorf("got %d, want 403", res.StatusCode)
	}
}

func TestCancelTask_Unauthenticated(t *testing.T) {
	env := newTestEnv(t)
	taskID := submitTestTask(t, env)
	res := env.do("POST", "/v1/tasks/"+taskID+"/cancel", "bad-token", nil)
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", res.StatusCode)
	}
}

// submitTestTask creates a task and returns its ID.
func submitTestTask(t *testing.T, env *testEnv) string {
	t.Helper()
	res := env.do("POST", "/v1/tasks", env.adminToken, map[string]any{
		"input":  map[string]string{"type": "prompt", "text": "create an S3 bucket"},
		"output": map[string]string{"type": "print"},
	})
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		t.Fatalf("submit task: got %d", res.StatusCode)
	}
	var resp struct {
		TaskID string `json:"task_id"`
	}
	_ = json.NewDecoder(res.Body).Decode(&resp)
	if resp.TaskID == "" {
		t.Fatal("empty task_id in response")
	}
	time.Sleep(50 * time.Millisecond)
	return resp.TaskID
}
