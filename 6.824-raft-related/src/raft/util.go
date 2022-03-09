package raft

import (
	"fmt"
	"log"
	"os"
)

// Debugging
const (
	Debug      = false
	MaxLogSize = 100
)

func init() {
	f, err := os.OpenFile("/home/noneback/workspace/Toys/6.824-raft-related/src/raft/test/log.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModeAppend)
	if err != nil {
		panic(err)
	}
	log.SetOutput(f)
}

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

func Min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func shrinkEntriesArray(entries []LogEntry) []LogEntry {
	// We replace the array if we're using less than half of the space in
	// it. This number is fairly arbitrary, chosen as an attempt to balance
	// memory usage vs number of allocations. It could probably be improved
	// with some focused tuning.
	const lenMultiple = 2
	if len(entries)*lenMultiple < cap(entries) {
		newEntries := make([]LogEntry, len(entries))
		copy(newEntries, entries)
		return newEntries
	}
	return entries
}

func LogInfoToString(entries *[]LogEntry) string {
	return fmt.Sprintf("addr %p, len %v, cap %v, content %+v", &entries, len(*entries), cap(*entries), entries)
}

type LogEntries []LogEntry

func (logs LogEntries) Less(i, j int) bool {
	return logs[i].Index < logs[j].Index
}
func (logs LogEntries) Len() int { return len(logs) }

func (logs LogEntries) Swap(i, j int) {
	logs[i], logs[j] = logs[j], logs[i]
}

func ReverseSortedIndexes(indexes []int) []int {
	for i, j := 0, len(indexes)-1; i < j; i, j = i+1, j-1 {
		indexes[i], indexes[j] = indexes[j], indexes[i]
	}
	return indexes
}
