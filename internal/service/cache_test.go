package service_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"terminalog/internal/model"
	"terminalog/internal/service"
)

func TestArticleCache_InvalidationRejectsStaleWrite(t *testing.T) {
	cache := service.NewArticleCache(time.Minute)
	staleGeneration := cache.Generation()

	cache.Invalidate()

	stored := cache.SetDirectoryList(staleGeneration, "root", []model.Article{{Path: "old.md"}})
	assert.False(t, stored)
	_, found := cache.GetDirectoryList("root")
	assert.False(t, found)
}

func TestArticleCache_TTLIsPerEntry(t *testing.T) {
	cache := service.NewArticleCache(20 * time.Millisecond)
	generation := cache.Generation()
	cache.SetDirectoryList(generation, "old", []model.Article{{Path: "old.md"}})

	time.Sleep(25 * time.Millisecond)
	cache.SetDirectoryList(generation, "new", []model.Article{{Path: "new.md"}})

	_, oldFound := cache.GetDirectoryList("old")
	_, newFound := cache.GetDirectoryList("new")
	assert.False(t, oldFound)
	assert.True(t, newFound)
}
