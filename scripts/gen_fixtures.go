//go:build ignore

// gen_fixtures writes deterministic, tiny stand-in "model" files so the whole
// pipeline (hash -> manifest -> magnet -> seed -> fetch -> verify) can be
// exercised without downloading a real multi-GB model. Disk-safe by design:
// each file is 64 KB, the whole set is well under 1 MB, and testdata/fixtures
// is gitignored.
//
// Run with: make fixtures   (i.e. `go run ./scripts/gen_fixtures.go`)
package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
)

const fixtureSize = 64 * 1024 // 64 KB per stand-in file

// files are named like real model artifacts so hflayout/tests look realistic.
var files = []string{
	"model.safetensors",
	"tokenizer.json",
	"Q4_K_M.gguf",
}

func main() {
	outDir := filepath.Join("testdata", "fixtures")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "mkdir:", err)
		os.Exit(1)
	}

	// Fixed seed => byte-for-byte reproducible fixtures across machines/runs.
	rng := rand.New(rand.NewSource(0x4d6f64656c))

	for _, name := range files {
		buf := make([]byte, fixtureSize)
		if _, err := rng.Read(buf); err != nil {
			fmt.Fprintln(os.Stderr, "rng:", err)
			os.Exit(1)
		}
		p := filepath.Join(outDir, name)
		if err := os.WriteFile(p, buf, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "write:", err)
			os.Exit(1)
		}
		fmt.Printf("wrote %s (%d bytes)\n", p, len(buf))
	}
	fmt.Println("fixtures ready under", outDir)
}
