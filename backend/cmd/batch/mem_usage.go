package main

import (
	"runtime"

	"github.com/inhies/go-bytesize"
	"github.com/rs/zerolog/log"
)

func findMemUsage() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return m
}

func printMemUsage(mem runtime.MemStats) {
	log.Info().
		Str("Alloc", bytesize.New(float64(mem.Alloc)).String()).
		Str("Total Alloc", bytesize.New(float64(mem.TotalAlloc)).String()).
		Str("Sys", bytesize.New(float64(mem.Sys)).String()).
		Uint32("GC Count", mem.NumGC).
		Msg("memory usage")
}

func printMemDiff(oldMem, newMem runtime.MemStats) {
	log.Info().
		Str("Alloc", bytesize.New(float64(newMem.Alloc-oldMem.Alloc)).String()).
		Str("Total Alloc", bytesize.New(float64(newMem.TotalAlloc-oldMem.TotalAlloc)).String()).
		Str("Sys", bytesize.New(float64(newMem.Sys-oldMem.Sys)).String()).
		Uint32("GC Count", newMem.NumGC-oldMem.NumGC).
		Msg("memory diff")

}
