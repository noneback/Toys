package raft

import (
	"fmt"
	"log"
	"os"
)

// Debugging
const (
	Debug      = true
	MaxLogSize = 100
)

func init() {
	f, err := os.OpenFile("./test/log.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModeAppend)
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
