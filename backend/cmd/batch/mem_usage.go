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
		Str("Alloc", bytesize.New(float64(oldMem.Alloc-newMem.Alloc)).String()).
		Str("Total Alloc", bytesize.New(float64(oldMem.TotalAlloc-newMem.TotalAlloc)).String()).
		Str("Sys", bytesize.New(float64(oldMem.Sys-newMem.Sys)).String()).
		Uint32("GC Count", oldMem.NumGC-newMem.NumGC).
		Msg("memory diff")

}
