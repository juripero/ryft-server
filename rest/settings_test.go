package rest

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// test settings and jobs
func TestSettingsJobs(t *testing.T) {
	setLoggingLevel("core/pending-jobs", testLogLevel)

	path := "/tmp/ryft-test.settings"
	os.RemoveAll(path)
	defer os.RemoveAll(path)

	s, err := OpenSettings(path)
	if !assert.NoError(t, err) {
		return
	}

	assert.EqualValues(t, path, s.GetPath())
	assert.True(t, s.CheckScheme())

	id1, err := s.AddJob("cmd1", "arg1", time.Now().Add(time.Minute))
	assert.NoError(t, err)
	id2, err := s.AddJob("cmd2", "arg2", time.Now().Add(time.Hour))
	assert.NoError(t, err)

	p, err := s.GetNextJobTime()
	assert.NoError(t, err)

	jobsCh, err := s.QueryAllJobs(p)
	if assert.NoError(t, err) {
		jobs := make([]SettingsJobItem, 0)
		for i := range jobsCh {
			jobs = append(jobs, i)
		}

		if assert.EqualValues(t, 1, len(jobs)) {
			// job.String() contains nanoseconds
			assert.Contains(t, jobs[0].String(), fmt.Sprintf("#%d [cmd1 arg1] at %s", id1, p.UTC().Format(jobTimeFormat)))
		}
	}

	err = s.DeleteJobs([]int64{id1, id2})
	assert.NoError(t, err)

	assert.NoError(t, s.ClearAll())
	assert.NoError(t, s.Close())
	assert.NoError(t, s.Close())
}
