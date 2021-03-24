package fileargs

import (
	"fmt"
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

// ReadArguments reads arguments contained
// in file, and return a filled *FileArguments
func ReadArguments(file string) (*FileArguments, error) {
	url := 0
	_ = url
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	contentS := strings.Split(string(content), "\n")
	args := FileArguments{
		Periods: make([]Period, len(contentS)-1),
		CfgPath: "",
	}
	for idx, line := range contentS {
		if idx == 0 {
			args.CfgPath = strings.TrimSpace(line)
			if !path.IsAbs(args.CfgPath) {
				args.CfgPath = path.Join(path.Dir(file), args.CfgPath)
			}
			_, err := os.Stat(args.CfgPath)
			if err != nil {
				return nil, fmt.Errorf(`
Expected format for arguments.txt:
/path/to/cfg/file
YYYYMMDDHH HOURS
...
Cannot stat config file "%s": %w`, line, err)
			}
			continue
		}
		parts := strings.Split(line, " ")
		if len(parts) != 2 {
			return nil, fmt.Errorf(`
Expected format for arguments.txt:
/path/to/cfg/file
YYYYMMDDHH HOURS
...
Cannot parse line
%s`, line)
		}
		date, err := time.Parse("2006010215", parts[0])
		if err != nil {
			return nil, fmt.Errorf(`
Expected format for arguments.txt:
/path/to/cfg/file
YYYYMMDDHH HOURS
...
Cannot parse line "%s": %w`, line, err)
		}
		var duration time.Duration
		dur, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf(`
Expected format for arguments.txt:
/path/to/cfg/file
YYYYMMDDHH HOURS
...
Cannot parse line "%s": %w`, line, err)

		}
		duration = time.Hour * time.Duration(dur)
		tp := Period{
			Start:    date,
			Duration: duration,
		}

		args.Periods[idx-1] = tp
	}

	return &args, nil
}
