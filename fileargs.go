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

// Reader ...
type Reader struct {
	R io.Reader
}

func (r *Reader) parseConfigFilePath(cwd string, scanner *bufio.Scanner) (string, error) {
	ok, firstLine := scanner.Scan(), scanner.Text()
	if !ok {
		return "", mkerr(scanner.Err(), `Malformed file: missing config path`)
	}

	cfgPath := strings.TrimSpace(firstLine)
	if !path.IsAbs(cfgPath) {
		cfgPath = path.Join(cwd, cfgPath)
	}

	_, err := os.Stat(cfgPath)
	if err != nil {
		return "", mkerr(err, `Config file "%s" not found`, firstLine)
	}

	return cfgPath, nil
}

type ctx struct {
	Err error
}

func (err *ctx) splitParts(line string) []string {
	parts := strings.Split(line, " ")
	if len(parts) != 2 {
		innerr := fmt.Errorf("2 fields expected, got %d", len(parts))
		err.Err = mkerr(innerr, `Cannot parse line "%s"`, line)
	}
	return parts
}

func (err *ctx) parseStart(text string) time.Time {
	if err.Err != nil {
		return time.Time{}
	}

	date, innerr := time.Parse("2006010215", text)
	if innerr != nil {
		err.Err = mkerr(innerr, `Cannot parse "%s" as date`, text)
		return time.Time{}
	}

	return date
}

func (err *ctx) parseDuration(text string) time.Duration {
	if err.Err != nil {
		return time.Duration(0)
	}
	dur, innerr := strconv.ParseInt(text, 10, 64)
	if innerr != nil {
		err.Err = mkerr(innerr, `Cannot parse "%s" as number`, text)
		return time.Duration(0)
	}
	return time.Hour * time.Duration(dur)
}

// ReadAll ...
func (r *Reader) ReadAll(cwd string) (*FileArguments, error) {
	var args *FileArguments
	scanner := bufio.NewScanner(r.R)

	cfgPath, err := r.parseConfigFilePath(cwd, scanner)
	if err != nil {
		return nil, err
	}

	args = &FileArguments{
		Periods: []Period{},
		CfgPath: cfgPath,
	}

	for scanner.Scan() {
		line := scanner.Text()
		var c ctx

		parts := c.splitParts(line)

		tp := Period{
			Start:    c.parseStart(parts[0]),
			Duration: c.parseDuration(parts[1]),
		}
		if c.Err != nil {
			return nil, c.Err
		}
		args.Periods = append(args.Periods, tp)
	}

	return args, scanner.Err()
}

// ReadArguments reads arguments contained
// in file, and return a filled *FileArguments
func ReadArguments(file string) (*FileArguments, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := &Reader{f}

	return reader.ReadAll(path.Dir(file))

}
