package fileargs_test

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"time"

	"github.com/parro-it/fileargs"
)

//go:embed fixtures
var fixtureRootFS embed.FS
var fixtureFS, _ = fs.Sub(fixtureRootFS, "fixtures")

// This example show how to use
// fileargs.ReadFile to parse from
// a file.
func ExampleReadFile() {
	args, err := fileargs.ReadFile(fixtureFS, "dates.txt")
	if err != nil {
		panic(err)
	}
	fmt.Println(args.String())
	// Output: wrfda-runner.cfg
	// 2020112600 24
	// 2020112700 48
}

// This example show how to use
// fileargs.ReadAll to parse from
// an io.Reader instance
func ExampleReadAll() {
	f, err := fixtureFS.Open("dates.txt")
	if err != nil {
		panic(err)
	}
	args, err := fileargs.ReadAll(f, fixtureFS)
	if err != nil {
		panic(err)
	}
	fmt.Println(args.String())
	// Output: wrfda-runner.cfg
	// 2020112600 24
	// 2020112700 48
}

// This example creates a fileargs.FileArguments
// and show how it can be formatted into a string
// with String method.
func ExampleFileArguments() {
	args := fileargs.FileArguments{
		CfgPath: "wrfda-runner.cfg",
		Periods: []*fileargs.Period{
			{time.Date(2020, 11, 26, 0, 0, 0, 0, time.UTC), 24 * time.Hour},
			{time.Date(2020, 11, 27, 0, 0, 0, 0, time.UTC), 48 * time.Hour},
		},
	}

	fmt.Println(args.String())
	fmt.Println(args.Periods[1].String())
	// Output: wrfda-runner.cfg
	// 2020112600 24
	// 2020112700 48
	//
	// 2020112700 48
}

// This example creates a fake
// bytes.Buffer containing a well formatted
// file args, and then manually creates
// a fileargs.Scanner to parse it.
func ExampleScanner() {
	var buf bytes.Buffer
	buf.WriteString("wrfda-runner.cfg\n")
	buf.WriteString("2020112600 24\n")
	buf.WriteString("2020112700 48\n")

	r := fileargs.New(&buf, fixtureFS)

	for r.Scan() {
		if cfg, ok := r.CfgPath(); ok {
			fmt.Println(cfg)
			continue
		}

		p, _ := r.Period()

		fmt.Println(p.String())
	}

	if err := r.Err(); err != nil {
		panic(err)
	}

	// Output: wrfda-runner.cfg
	// 2020112600 24
	// 2020112700 48
}
