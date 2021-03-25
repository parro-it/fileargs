// Package fileargs implements a scanner that
// allows to read a FileArguments struct from
// an io.Reader.
//
// The semantic of API is the same as bufio.Scanner.
package fileargs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"strconv"
	"strings"
	"time"
)

// A FileArguments contains all informations
// read from an arguments file.
// The type is a Stringer.
type FileArguments struct {
	Periods []*Period
	CfgPath string
}

// String implements fmt.Stringer
func (args FileArguments) String() string {
	var buf bytes.Buffer
	buf.WriteString(args.CfgPath + "\n")

	for _, p := range args.Periods {
		buf.WriteString(p.String() + "\n")
	}
	return buf.String()
}

var _ fmt.Stringer = FileArguments{}

// A Period represents a lapse of time
// with a Start instant and a Duration
// The type is a Stringer.
type Period struct {
	Start    time.Time
	Duration time.Duration
}

// String implements fmt.Stringer
func (p Period) String() string {
	return fmt.Sprintf("%s %02d", p.Start.Format("2006010215"), int(p.Duration.Hours()))
}

var _ fmt.Stringer = Period{}

// Scanner provides a convenient interface for reading an arguments file.
// Successive calls to the Scan method will step through the lines of a file,
// skipping the bytes between the tokens. First line is parsed as a string
// to be used as config file path, other ones as Period.
//
// Scanning stops unrecoverably at EOF, the first I/O error, or a malformed line.
// When a scan stops, the reader may have advanced arbitrarily far past the last token.
type Scanner struct {
	Src           *bufio.Scanner
	Fsys          fs.FS
	firstLineDone bool
	cfgPath       string
	period        *Period
	err           error
}

// New returns a new Scanner initialized for the
// given source, that will use fsys to check for
// config file existence during scan.
func New(source io.Reader, fsys fs.FS) *Scanner {
	scanner := bufio.NewScanner(source)
	return &Scanner{
		Src:  scanner,
		Fsys: fsys,
	}
}

// Err returns the first non-EOF error that was encountered by the Scanner.
func (r *Scanner) Err() error {
	return r.err
}

// Scan advances the Scanner to the next line, parsed data will be available
// through the CfgPath or Period method. It returns false when the scan stops,
// either by reaching the end of the input or an error. After Scan returns false,
// the Err method will return any error that occurred during scanning, except
// that if it was io.EOF, Err will return nil.
func (r *Scanner) Scan() bool {
	if !r.Src.Scan() {
		if !r.firstLineDone {
			r.err = mkerr(r.Src.Err(), `Malformed file: missing config path`)
		} else {
			r.err = r.Src.Err()
		}
		return false
	}

	if !r.firstLineDone {
		r.parseCfgFilePath()
		r.firstLineDone = true
		return r.err == nil

	}

	r.parsePeriod()

	return r.err == nil
}

// CfgPath returns the last value parsed
// if it was the config file path, otherwise
// it fails and return an empty string.
//
// Second bool return value indicates if the returned
// string is valid.
func (r *Scanner) CfgPath() (string, bool) {
	return r.cfgPath, r.cfgPath != "" && r.err == nil
}

// Period returns the last value parsed
// if it was a period, otherwise
// it fails and return nil.
//
// Second bool return value indicates if the returned
// period is valid.
func (r *Scanner) Period() (*Period, bool) {
	return r.period, r.period != nil && r.err == nil
}

// ReadAll is a convenience functions that creates
// a Scanner from reader source, read it to completion
// and returns a *FileArguments structure filled from the
// data read.
//
// fsys is used to check for config file existence during scan.
//
// Scanner error is returned if any happens.
func ReadAll(reader io.Reader, fsys fs.FS) (*FileArguments, error) {
	var args FileArguments

	r := New(reader, fsys)

	for r.Scan() {
		if cfg, ok := r.CfgPath(); ok {
			args.CfgPath = cfg
			continue
		}

		p, ok := r.Period()
		if !ok {
			break
		}
		args.Periods = append(args.Periods, p)

	}

	if err := r.Err(); err != nil {
		return nil, err
	}
	return &args, nil
}

// ReadFile reads arguments contained
// in file, and return a filled *FileArguments
// or an error if any happens.
// A givens fs.FS instance is used to read the file.
func ReadFile(fsys fs.FS, file string) (*FileArguments, error) {
	f, err := fsys.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ReadAll(f, fsys)
}

func mkerr(err error, format string, arguments ...interface{}) error {
	errmsg := fmt.Sprintf(format, arguments...)
	errfrm := strings.ReplaceAll(`
		Expected format for arguments.txt:
		/path/to/cfg/file
		YYYYMMDDHH HOURS
		...
		%s: %w
	`, "\t", "")

	return fmt.Errorf(errfrm, errmsg, err)
}

func (r *Scanner) splitParts(line string) []string {
	if r.err != nil {
		return make([]string, 2)
	}
	parts := strings.Split(line, " ")
	if len(parts) != 2 {
		innerr := fmt.Errorf("2 fields expected, got %d", len(parts))
		r.err = mkerr(innerr, `Cannot parse line "%s"`, line)
		return make([]string, 2)
	}
	return parts
}

func (r *Scanner) parseStart(text string) time.Time {
	if r.err != nil {
		return time.Time{}
	}

	date, innerr := time.Parse("2006010215", text)
	if innerr != nil {
		r.err = mkerr(innerr, `Cannot parse "%s" as date`, text)
		return time.Time{}
	}

	return date
}

func (r *Scanner) parseDuration(text string) time.Duration {
	if r.err != nil {
		return time.Duration(0)
	}
	dur, innerr := strconv.ParseInt(text, 10, 64)
	if innerr != nil {
		r.err = mkerr(innerr, `Cannot parse "%s" as number`, text)
		return time.Duration(0)
	}
	return time.Hour * time.Duration(dur)
}

func (r *Scanner) parseCfgFilePath() {
	if r.err != nil {
		return
	}
	r.cfgPath = ""
	r.period = nil

	cfgPath := strings.TrimSpace(r.Src.Text())

	_, err := fs.Stat(r.Fsys, cfgPath)
	if err != nil {
		r.err = mkerr(err, `Config file "%s" not found`, r.Src.Text())
		return
	}

	r.cfgPath = cfgPath
}

func (r *Scanner) parsePeriod() {
	if r.err != nil {
		return
	}

	r.cfgPath = ""
	r.period = nil

	line := r.Src.Text()

	parts := r.splitParts(line)

	tp := Period{
		Start:    r.parseStart(parts[0]),
		Duration: r.parseDuration(parts[1]),
	}

	if r.err != nil {
		return
	}

	r.period = &tp

}
