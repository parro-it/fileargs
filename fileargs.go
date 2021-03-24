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

// ReadAll ...
func (r *Reader) ReadAll(cwd string) (*FileArguments, error) {
	var args *FileArguments
	scanner := bufio.NewScanner(r.R)

	ok, firstLine := scanner.Scan(), scanner.Text()
	if !ok {
		return nil, mkerr(scanner.Err(), `Malformed file: missing config path`)
	}
	args = &FileArguments{
		Periods: []Period{},
		CfgPath: strings.TrimSpace(firstLine),
	}

	if !path.IsAbs(args.CfgPath) {
		args.CfgPath = path.Join(cwd, args.CfgPath)
	}

	_, err := os.Stat(args.CfgPath)
	if err != nil {
		return nil, mkerr(err, `Config file "%s" not found`, firstLine)
	}

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		if len(parts) != 2 {
			err := fmt.Errorf("2 fields expected, got %d", len(parts))
			return nil, mkerr(err, `Cannot parse line "%s"`, line)
		}
		date, err := time.Parse("2006010215", parts[0])
		if err != nil {
			return nil, mkerr(err, `Cannot parse "%s" as date`, parts[0])
		}
		var duration time.Duration
		dur, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, mkerr(err, `Cannot parse "%s" as number`, parts[1])
		}
		duration = time.Hour * time.Duration(dur)
		tp := Period{
			Start:    date,
			Duration: duration,
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
