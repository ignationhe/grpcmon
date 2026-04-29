package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTagStore_GetMissing(t *testing.T) {
	s := NewTagStore()
	assert.Nil(t, s.Get("localhost:50051"))
}

func TestTagStore_SetAndGet(t *testing.T) {
	s := NewTagStore()
	s.Set("localhost:50051", map[string]string{"env": "prod", "region": "us-east"})

	tags := s.Get("localhost:50051")
	assert.Equal(t, "prod", tags["env"])
	assert.Equal(t, "us-east", tags["region"])
}

func TestTagStore_GetReturnsCopy(t *testing.T) {
	s := NewTagStore()
	s.Set("localhost:50051", map[string]string{"env": "prod"})

	tags := s.Get("localhost:50051")
	tags["env"] = "mutated"

	// Original should be unchanged.
	original := s.Get("localhost:50051")
	assert.Equal(t, "prod", original["env"])
}

func TestTagStore_SetOverwrites(t *testing.T) {
	s := NewTagStore()
	s.Set("localhost:50051", map[string]string{"env": "prod"})
	s.Set("localhost:50051", map[string]string{"env": "staging", "team": "platform"})

	tags := s.Get("localhost:50051")
	assert.Equal(t, "staging", tags["env"])
	assert.Equal(t, "platform", tags["team"])
	assert.Len(t, tags, 2)
}

func TestTagStore_Delete(t *testing.T) {
	s := NewTagStore()
	s.Set("localhost:50051", map[string]string{"env": "prod"})
	s.Delete("localhost:50051")
	assert.Nil(t, s.Get("localhost:50051"))
}

func TestTagStore_All(t *testing.T) {
	s := NewTagStore()
	s.Set("a:1", map[string]string{"x": "1"})
	s.Set("b:2", map[string]string{"y": "2"})

	all := s.All()
	assert.Len(t, all, 2)
	assert.Equal(t, "1", all["a:1"]["x"])
	assert.Equal(t, "2", all["b:2"]["y"])
}

func TestTagStore_AllReturnsCopy(t *testing.T) {
	s := NewTagStore()
	s.Set("a:1", map[string]string{"env": "prod"})

	all := s.All()
	all["a:1"]["env"] = "mutated"

	original := s.Get("a:1")
	assert.Equal(t, "prod", original["env"])
}
