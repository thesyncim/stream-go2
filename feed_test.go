package stream_test

import (
	"sort"
	"testing"
	"time"

	stream "github.com/reifcode/stream-go2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeedID(t *testing.T) {
	client := newClient(t)

	flat := client.FlatFeed("flat", "123")
	assert.Equal(t, "flat:123", flat.ID())

	aggregated := client.AggregatedFeed("aggregated", "456")
	assert.Equal(t, "aggregated:456", aggregated.ID())
}

func TestAddActivities(t *testing.T) {
	client := newClient(t)
	flat := newFlatFeed(client)
	bobActivity := stream.Activity{Actor: "bob", Verb: "like", Object: "ice-cream"}
	aliceActivity := stream.Activity{Actor: "alice", Verb: "dislike", Object: "ice-cream"}
	resp, err := flat.AddActivities(bobActivity, aliceActivity)
	require.NoError(t, err)
	assert.Len(t, resp.Activities, 2)
}

func TestUpdateActivities(t *testing.T) {
	client := newClient(t)
	flat := newFlatFeed(client)
	bobActivity := stream.Activity{Actor: "bob", Verb: "like", Object: "ice-cream", ForeignID: "bob:123", Time: getTime(time.Now()), Extra: map[string]interface{}{"influence": 42}}
	_, err := flat.AddActivities(bobActivity)
	require.NoError(t, err)

	bobActivity.Extra = map[string]interface{}{"influence": 42}
	err = flat.UpdateActivities(bobActivity)
	require.NoError(t, err)

	resp, err := flat.GetActivities()
	require.NoError(t, err)
	assert.Len(t, resp.Results, 1)
	assert.NotEmpty(t, resp.Results[0].Extra)
}

func TestRemoveActivities(t *testing.T) {
	client := newClient(t)
	flat := newFlatFeed(client)
	activities := []stream.Activity{
		stream.Activity{
			Actor:  "john",
			Verb:   "like",
			Object: "something",
		},
		stream.Activity{
			Actor:     "bob",
			Verb:      "like",
			Object:    "something",
			ForeignID: "bob:123",
		},
	}
	added, err := flat.AddActivities(activities...)
	require.NoError(t, err)
	activities = added.Activities

	err = flat.RemoveActivityByID(activities[0].ID)
	assert.NoError(t, err)
	resp, err := flat.GetActivities()
	assert.Len(t, resp.Results, 1)
	assert.Equal(t, activities[1].ID, resp.Results[0].ID)

	err = flat.RemoveActivityByForeignID("bob:123")
	assert.NoError(t, err)
	resp, err = flat.GetActivities()
	assert.Len(t, resp.Results, 0)
}

func TestUpdateToTargets(t *testing.T) {
	client := newClient(t)
	flat := newFlatFeed(client)
	f1, f2, f3 := newFlatFeedWithUserID(client, "f1"), newFlatFeedWithUserID(client, "f2"), newFlatFeedWithUserID(client, "f3")
	activity := stream.Activity{Time: getTime(time.Now()), ForeignID: "bob:123", Actor: "bob", Verb: "like", Object: "ice-cream", To: []string{f1.ID()}, Extra: map[string]interface{}{"popularity": 9000}}
	sort.Strings(activity.To)
	_, err := flat.AddActivity(activity)
	require.NoError(t, err)

	resp, err := flat.GetActivities()
	require.NoError(t, err)
	assert.Len(t, resp.Results, 1)
	assert.Len(t, resp.Results[0].To, 1)
	assert.Equal(t, f1.ID(), resp.Results[0].To[0])

	err = flat.UpdateToTargets(activity, []stream.Feed{f2}, nil)
	require.NoError(t, err)
	resp, err = flat.GetActivities()
	require.NoError(t, err)
	assert.Len(t, resp.Results, 1)
	require.Len(t, resp.Results[0].To, 2)
	assert.Equal(t, f1.ID(), resp.Results[0].To[0])
	assert.Equal(t, f2.ID(), resp.Results[0].To[1])

	err = flat.ReplaceToTargets(activity, []stream.Feed{f3})
	require.NoError(t, err)
	resp, err = flat.GetActivities()
	require.NoError(t, err)
	assert.Len(t, resp.Results, 1)
	assert.Len(t, resp.Results[0].To, 1)
	assert.Equal(t, f3.ID(), resp.Results[0].To[0])

	err = flat.UpdateToTargets(activity, nil, []stream.Feed{f3})
	require.NoError(t, err)
	resp, err = flat.GetActivities()
	require.NoError(t, err)
	assert.Len(t, resp.Results, 1)
	assert.Len(t, resp.Results[0].To, 0)
}

func TestToken(t *testing.T) {
	client, err := stream.NewClient("key", "super secret")
	require.NoError(t, err)
	flat := newFlatFeedWithUserID(client, "sample")
	testCases := []struct {
		readOnly bool
		expected string
	}{
		{
			readOnly: false,
			expected: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY3Rpb24iOiJ3cml0ZSIsImZlZWRfaWQiOiJmbGF0c2FtcGxlIiwicmVzb3VyY2UiOiJmZWVkIn0._7eLZ3-_6dmOoCKp8MvSoKCp0PA-gAerKnr8tuwut2M",
		},
		{
			readOnly: true,
			expected: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY3Rpb24iOiJyZWFkIiwiZmVlZF9pZCI6ImZsYXRzYW1wbGUiLCJyZXNvdXJjZSI6ImZlZWQifQ.Ab6NX3dAGbBiXkQrEIWg9Z-WRm1R4710ont2y0OONiE",
		},
	}
	for _, tc := range testCases {
		token := flat.Token(tc.readOnly)
		assert.Equal(t, tc.expected, token)
	}
}
