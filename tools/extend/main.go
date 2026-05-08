package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/k1LoW/subAnimJ/tools/internal/animcjk"
	"github.com/k1LoW/subAnimJ/tools/internal/extend"
	"github.com/k1LoW/subAnimJ/tools/internal/runner"
)

func main() {
	log.SetFlags(0)
	var (
		targetsDir = flag.String("targets-dir", "", "directory of per-kanji JSONL targets")
		kanji      = flag.String("kanji", "", "single-shot: target kanji (one rune)")
		stroke     = flag.Int("stroke", 0, "single-shot: 1-origin stroke index")
		direction  = flag.String("direction", "", "single-shot: horizontal|vertical")
		upstream   = flag.String("upstream", "../vendor/animCJK", "upstream animCJK root")
		outRoot    = flag.String("out", "..", "output root (where svgsJa/ and graphicsJa.txt are written)")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s --targets-dir DIR\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --kanji 日 --stroke 3 --direction horizontal\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *targetsDir != "" {
		if err := runner.RunDir(*upstream, *outRoot, *targetsDir); err != nil {
			log.Fatal(err)
		}
		return
	}

	if *kanji == "" || *stroke == 0 || *direction == "" {
		flag.Usage()
		log.Fatal("specify --targets-dir or --kanji/--stroke/--direction")
	}

	g, err := animcjk.LoadFromUpstream(*upstream, *kanji)
	if err != nil {
		log.Fatal(err)
	}
	dir, err := extend.ParseDirection(*direction)
	if err != nil {
		log.Fatal(err)
	}
	if err := extend.Apply(g, *stroke, dir, extend.BothSides); err != nil {
		log.Fatal(err)
	}
	if err := animcjk.WriteSVG(*upstream, *outRoot, g); err != nil {
		log.Fatal(err)
	}
	if err := animcjk.WriteGraphicsJa(*outRoot, g); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "extended %s stroke %d (%s)\n", *kanji, *stroke, *direction)
}
