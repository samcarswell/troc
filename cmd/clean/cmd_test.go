package cmd

import (
	"math/rand"
	"os"
	"path"
	"testing"
	"time"

	"github.com/samcarswell/troc/test"
)

func Test_clearSystemLogs(t *testing.T) {
	invalid1 := "trocsys_notavalidfilename.log"
	invalid2 := "shouldbeignored.log"
	invalid3 := "trocsys_.log"
	invalid4 := "trocsys_" + randStringRunes(5) + getValidDatetimeStr(time.Now()) + ".log"
	invalid5 := "trocsys_" + randStringRunes(5) + "_" + getValidDatetimeStr(time.Now()) + ".log"
	invalid6 := "trocsys_" + randStringRunes(5) + "_" + getValidDatetimeStr(time.Now().AddDate(0, 0, -29)) + ".log"
	valid1 := "trocsys_" + randStringRunes(5) + "_" + getValidDatetimeStr(time.Now().AddDate(0, 0, -40)) + ".log"
	valid2 := "trocsys_" + randStringRunes(15) + "_" + getValidDatetimeStr(time.Now().AddDate(0, 0, -40)) + ".log"

	rootDir, err := os.MkdirTemp("", "troctest")
	if err != nil {
		t.Fatal(err)
	}
	touchFileOrPanic(rootDir, invalid1, t)
	touchFileOrPanic(rootDir, invalid2, t)
	touchFileOrPanic(rootDir, invalid3, t)
	touchFileOrPanic(rootDir, invalid4, t)
	touchFileOrPanic(rootDir, invalid5, t)
	touchFileOrPanic(rootDir, invalid6, t)
	touchFileOrPanic(rootDir, valid1, t)
	touchFileOrPanic(rootDir, valid2, t)

	logFile, logger := test.CreateSysLogFile(t)

	err = clearMiscSystemLogs(rootDir, time.Now().AddDate(0, 0, -30), logger)
	if err != nil {
		t.Fatal(err)
	}
	log := test.NewLogFromFileOrFail(logFile)

	test.AssertLogHasWarn(t, "log file not recognised: "+path.Join(rootDir, invalid1)+". skipping", log)
	test.AssertLogDoesNotHaveInfo(t, "removing syslog file "+path.Join(rootDir, invalid1), log)
	test.AssertLogDoesNotHaveInfo(t, "deleted syslog file "+path.Join(rootDir, invalid1), log)

	test.AssertLogDoesNotHaveWarn(t, "log file not recognised: "+path.Join(rootDir, invalid2)+". skipping", log)
	test.AssertLogDoesNotHaveInfo(t, "removing syslog file "+path.Join(rootDir, invalid2), log)
	test.AssertLogDoesNotHaveInfo(t, "deleted syslog file "+path.Join(rootDir, invalid2), log)

	test.AssertLogHasWarn(t, "log file not recognised: "+path.Join(rootDir, invalid3)+". skipping", log)
	test.AssertLogHasWarn(t, "log file not recognised: "+path.Join(rootDir, invalid4)+". skipping", log)

	// invalid because of date. Should not appear in the logs at all
	test.AssertLogDoesNotHaveWarn(t, "log file not recognised: "+path.Join(rootDir, invalid5)+". skipping", log)
	test.AssertLogDoesNotHaveInfo(t, "removing syslog file "+path.Join(rootDir, invalid5), log)
	test.AssertLogDoesNotHaveInfo(t, "deleted syslog file "+path.Join(rootDir, invalid5), log)
	test.AssertLogDoesNotHaveWarn(t, "log file not recognised: "+path.Join(rootDir, invalid6)+". skipping", log)
	test.AssertLogDoesNotHaveInfo(t, "removing syslog file "+path.Join(rootDir, invalid6), log)
	test.AssertLogDoesNotHaveInfo(t, "deleted syslog file "+path.Join(rootDir, invalid6), log)

	test.AssertLogHasInfo(t, "removing syslog file "+path.Join(rootDir, valid1), log)
	test.AssertLogHasInfo(t, "removing syslog file "+path.Join(rootDir, valid2), log)
	test.AssertLogHasInfo(t, "deleted syslog file "+path.Join(rootDir, valid1), log)
	test.AssertLogHasInfo(t, "deleted syslog file "+path.Join(rootDir, valid2), log)

}

func touchFileOrPanic(dir string, file string, t *testing.T) {
	_, err := os.Create(path.Join(dir, file))
	if err != nil {
		t.Fatal(err)
	}
}

func getValidDatetimeStr(ts time.Time) string {
	return ts.Format("20060102T150405")
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
