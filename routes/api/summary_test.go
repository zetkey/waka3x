package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zetkey/waka3x/models"
)

func TestNewHourlyActivityResponse_UsesUserTimezone(t *testing.T) {
	tz, err := time.LoadLocation("Asia/Jakarta")
	require.NoError(t, err)

	response := newHourlyActivityResponse(models.Durations{
		{
			Time:     models.CustomTime(time.Date(2026, time.May, 18, 18, 30, 0, 0, time.UTC)),
			Duration: 20 * time.Minute,
		},
	}, tz)

	assert.Len(t, response, 24)
	assert.Equal(t, int64(20*60), response[1].Duration)
	assert.Equal(t, int64(0), response[18].Duration)
}
