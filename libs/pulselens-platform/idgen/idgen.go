package idgen

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	customEpochMillis = int64(1767225600000)
	nodeBits          = int64(10)
	sequenceBits      = int64(12)
	maxNodeID         = int64((1 << nodeBits) - 1)
	maxSequence       = int64((1 << sequenceBits) - 1)
)

type generator struct {
	mu        sync.Mutex
	nodeID    int64
	lastStamp int64
	sequence  int64
}

var global = &generator{}

func Configure(nodeID int64) {
	global.mu.Lock()
	defer global.mu.Unlock()

	if nodeID < 0 {
		nodeID = 0
	}
	global.nodeID = nodeID & maxNodeID
}

func NodeIDFromServiceName(serviceName string) int64 {
	var hash uint32 = 2166136261
	for _, ch := range strings.ToLower(serviceName) {
		hash ^= uint32(ch)
		hash *= 16777619
	}
	return int64(hash % uint32(maxNodeID+1))
}

func New(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, strconv.FormatInt(global.next(), 36))
}

func (g *generator) next() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UTC().UnixMilli()
	if now < customEpochMillis {
		now = customEpochMillis
	}

	if now == g.lastStamp {
		g.sequence = (g.sequence + 1) & maxSequence
		if g.sequence == 0 {
			now = waitNextMillis(g.lastStamp)
		}
	} else {
		g.sequence = 0
	}

	g.lastStamp = now
	timePart := (now - customEpochMillis) << (nodeBits + sequenceBits)
	nodePart := g.nodeID << sequenceBits
	return timePart | nodePart | g.sequence
}

func waitNextMillis(previous int64) int64 {
	current := time.Now().UTC().UnixMilli()
	for current <= previous {
		time.Sleep(time.Millisecond)
		current = time.Now().UTC().UnixMilli()
	}
	return current
}
