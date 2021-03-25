package fileargs

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// A FileArguments contains all informations
// read from an arguments.txt file.
type FileArguments struct {
	Periods []Period
	CfgPath string
}

// A Period represents a lapse of time
// with a Start instant and a Duration
type Period struct {
	Start    time.Time
	Duration time.Duration
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
	parts := strings.Split(line, " ")
	if len(parts) != 2 {
		innerr := fmt.Errorf("2 fields expected, got %d", len(parts))
		r.err = mkerr(innerr, `Cannot parse line "%s"`, line)
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

// Scanner ...
type Scanner struct {
	Src       *bufio.Scanner
	FirstLine bool
	CfgPath   string
	Period    Period
	err       error
	cwd       string
}

func New(source io.Reader, cwd string) *Scanner {
	scanner := bufio.NewScanner(source)
	return &Scanner{
		Src: scanner,
		cwd: cwd,
	}
}

func (r *Scanner) parseCfgFilePath() {
	cfgPath := strings.TrimSpace(r.Src.Text())
	if !path.IsAbs(cfgPath) {
		cfgPath = path.Join(r.cwd, cfgPath)
	}

	_, err := os.Stat(cfgPath)
	if err != nil {
		r.err = mkerr(err, `Config file "%s" not found`, r.Src.Text())
		return
	}

	r.FirstLine = true
	r.CfgPath = cfgPath
}

func (r *Scanner) parsePeriod() {
	line := r.Src.Text()

	parts := r.splitParts(line)

	tp := Period{
		Start:    r.parseStart(parts[0]),
		Duration: r.parseDuration(parts[1]),
	}

	r.FirstLine = false

	if r.err == nil {
		r.Period = tp
	}

}

func (r *Scanner) Err() error {
	return r.err
}

func (r *Scanner) Scan() bool {
	if !r.Src.Scan() {
		if r.CfgPath == "" {
			r.err = mkerr(r.Src.Err(), `Malformed file: missing config path`)
		} else {
			r.err = r.Src.Err()
		}
		return false
	}
	if r.CfgPath == "" {
		r.parseCfgFilePath()
	} else {
		r.parsePeriod()
	}

	return r.err == nil
}

// ReadAll ...
func ReadAll(reader io.Reader, cwd string) (*FileArguments, error) {
	var args FileArguments

	r := New(reader, cwd)

	for r.Scan() {
		if r.FirstLine {
			args.CfgPath = r.CfgPath
			continue
		}
		args.Periods = append(args.Periods, r.Period)
	}

	if err := r.Err(); err != nil {
		return nil, err
	}
	return &args, nil
}

// ReadFile reads arguments contained
// in file, and return a filled *FileArguments
func ReadFile(file string) (*FileArguments, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ReadAll(f, path.Dir(file))
}
