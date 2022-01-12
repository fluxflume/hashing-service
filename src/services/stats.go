package services

import (
	"sync"
	"time"
)

type StatsService interface {
	Update(handler func())
	AsMap() map[string]uint64
}

type LocalStats struct {
	total   uint64
	average uint64
	mu      *sync.RWMutex
}

func (stats *LocalStats) Update(handler func()) {
	startTime := time.Now()
	stats.mu.Lock()
	defer stats.mu.Unlock()
	prevTotal := stats.total
	prevAvg := stats.average
	newTotal := prevTotal + 1
	handler()
	timeTaken := uint64(time.Since(startTime).Microseconds())
	// Calculate the new average
	stats.average = (prevTotal*prevAvg + timeTaken) / (newTotal)
	// Assign the new total
	stats.total = newTotal
}

func (stats *LocalStats) AsMap() map[string]uint64 {
	m := make(map[string]uint64)
	stats.mu.RLock()
	defer stats.mu.RUnlock()
	m["total"] = stats.total
	m["average"] = stats.average
	return m
}

func CreateLocalStatsService() StatsService {
	return &LocalStats{mu: &sync.RWMutex{}, total: 0, average: 0}
}
