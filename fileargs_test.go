package fileargs

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func fixture(filePath string) string {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		panic("cannot retrieve the source file path")
	} else {
		file = filepath.Dir(file)
	}

	return path.Join(file, "fixtures", filePath)
}

func TestMatchDownloadedData(t *testing.T) {
	dateFile := fixture("dates.txt")
	args, err := ReadFile(dateFile)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(args.Periods))
	assert.Equal(t, "2020112600", args.Periods[0].Start.Format("2006010215"))
	assert.Equal(t, "2020112700", args.Periods[1].Start.Format("2006010215"))
	assert.Equal(t, time.Hour*24, args.Periods[0].Duration)
	assert.Equal(t, time.Hour*48, args.Periods[1].Duration)
	assert.Equal(t, fixture("wrfda-runner.cfg"), args.CfgPath)

}

func TestFileWrong(t *testing.T) {
	dateFile := fixture("wrong.txt")
	dates, err := ReadFile(dateFile)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf(`
Expected format for arguments.txt:
/path/to/cfg/file
YYYYMMDDHH HOURS
...
Config file "2020112600 24" not found: stat %s: no such file or directory
`, fixture("2020112600 24")), err.Error())

	assert.Nil(t, dates)

}

func TestFileWrong2(t *testing.T) {
	dateFile := fixture("wrong2.txt")
	dates, err := ReadFile(dateFile)
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
	dateFile := fixture("wrong3.txt")
	dates, err := ReadFile(dateFile)
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
	dateFile := fixture("wrong4.txt")
	dates, err := ReadFile(dateFile)
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
