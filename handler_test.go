package appsync_test

import (
	"context"
	"encoding/json"

	"testing"

	"github.com/bearchit/appsync-handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestNoMatchedResolver(t *testing.T) {
	h := appsync.NewHandler()
	h.AddResolver("Query.posts", func(ctx context.Context) error {
		return nil
	})

	payload := struct {
		Resolve   string `json:"resolve"`
		Arguments struct {
			UserID string `json:"userID"`
		}
	}{
		Resolve: "Query.unknown",
		Arguments: struct {
			UserID string `json:"userID"`
		}{
			UserID: "1",
		},
	}

	payloadJSON, err := json.Marshal(&payload)
	require.NoError(t, err)

	_, err = h.Handle(context.Background(), payloadJSON)
	assert.Equal(t, "no matched resolver: Query.unknown", err.Error())
}
