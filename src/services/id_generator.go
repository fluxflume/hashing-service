package services

import "sync/atomic"

type IdGenerator interface {
	Get() uint64
}

type LocalIdGenerator struct {
	currentId uint64
}

func (receiver *LocalIdGenerator) Get() uint64 {
	// The expectation is for numbers to start from 0.
	// Since this initially increments from 0 to 1, subtract 1 to reset the lower bound.
	return atomic.AddUint64(&receiver.currentId, 1) - 1
}

func NewLocalIdGenerator() IdGenerator {
	return &LocalIdGenerator{}
}
