
package appsync_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/bearchit/appsync-handler"
)

func TestHandler_Handle(t *testing.T) {
	type postsInput struct {
		UserID string `json:"userID"`
	}

	type post struct {
		ID string `json:"id"`
	}

	expected := []*post{
		{
			ID: "1",
		},
		{
			ID: "2",
		},
	}

	payload := struct {
		Resolve   string `json:"resolve"`
		Arguments struct {
			UserID string `json:"userID"`
		}
	}{
		Resolve: "Query.posts",
		Arguments: struct {
			UserID string `json:"userID"`
		}{
			UserID: "1",
		},
	}

	h := appsync.NewHandler()
	h.AddResolver("Query.posts", func(ctx context.Context, input *postsInput) ([]*post, error) {
		assert.Equal(t, "1", input.UserID)
		return expected, nil
	})

	payloadJSON, err := json.Marshal(&payload)
	require.NoError(t, err)

	resp, err := h.Handle(context.Background(), payloadJSON)
	require.NoError(t, err)
	assert.Equal(t, expected, resp)
}
