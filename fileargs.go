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
				return nil, mkerr(err, `Config file "%s" not found`, line)
			}
			continue
		}
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

		args.Periods[idx-1] = tp
	}

	return &args, nil
}
