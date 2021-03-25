package fileargs

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var fixtureDir = func() string {
	_, file, _, _ := runtime.Caller(1)
	return path.Join(filepath.Dir(file), "fixtures")
}()

var fixtureFS = os.DirFS(fixtureDir)

func TestMatchDownloadedData(t *testing.T) {
	args, err := ReadFile(fixtureFS, "dates.txt")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(args.Periods))
	assert.Equal(t, "2020112600", args.Periods[0].Start.Format("2006010215"))
	assert.Equal(t, "2020112700", args.Periods[1].Start.Format("2006010215"))
	assert.Equal(t, time.Hour*24, args.Periods[0].Duration)
	assert.Equal(t, time.Hour*48, args.Periods[1].Duration)
	assert.Equal(t, "wrfda-runner.cfg", args.CfgPath)

}

func TestFileWrong(t *testing.T) {
	dateFile := "wrong.txt"
	dates, err := ReadFile(fixtureFS, dateFile)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf(`
Expected format for arguments.txt:
/path/to/cfg/file
YYYYMMDDHH HOURS
...
Config file "2020112600 24" not found: open %s/2020112600 24: no such file or directory
`, fixtureDir), err.Error())

	assert.Nil(t, dates)

}

func TestFileWrong2(t *testing.T) {
	dateFile := "wrong2.txt"
	dates, err := ReadFile(fixtureFS, dateFile)
	assert.Error(t, err)
	assert.Equal(t, `
Expected format for arguments.txt:
/path/to/cfg/file
YYYYMMDDHH HOURS
...
Cannot parse line "2020112600 24 IT": 2 fields expected, got 3
`, err.Error())

	assert.Nil(t, dates)

}

func TestFileWrong3(t *testing.T) {
	dateFile := "wrong3.txt"
	dates, err := ReadFile(fixtureFS, dateFile)
	assert.Error(t, err)
	assert.Equal(t, `
Expected format for arguments.txt:
/path/to/cfg/file
YYYYMMDDHH HOURS
...
Cannot parse "2020112699" as date: parsing time "2020112699": hour out of range
`, err.Error())

	assert.Nil(t, dates)

}

func TestFileWrong4(t *testing.T) {
	dateFile := "wrong4.txt"
	dates, err := ReadFile(fixtureFS, dateFile)
	assert.Error(t, err)
	assert.Equal(t, `
Expected format for arguments.txt:
/path/to/cfg/file
YYYYMMDDHH HOURS
...
Cannot parse "AA" as number: strconv.ParseInt: parsing "AA": invalid syntax
`, err.Error())

	assert.Nil(t, dates)

}
