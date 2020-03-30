package goBolt

import (
	"github.com/mindstand/go-bolt/bolt_mode"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDecoderTimeIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}

	// setup

	client, err := NewClient(WithBasicAuth("neo4j", "changeme"),
		WithHostPort("0.0.0.0", 7687))
	if err != nil {
		t.Fatal(err)
	}

	driver, err := client.NewDriver()
	if err != nil {
		t.Fatal(err)
	}

	conn, err := driver.Open(bolt_mode.WriteMode)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = conn.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	// tests

	t.Run("Date", func(t *testing.T) {
		sample := time.Date(2020, 3, 24, 0, 0, 0, 0, time.UTC)
		sampleFormatted := sample.Format("2006-01-02")

		all, _, err := conn.QueryWithDb("RETURN date('"+sampleFormatted+"')", nil, "")
		if err != nil {
			t.Fatal(err)
		}

		returned, ok := all[0][0].(time.Time)
		if !ok {
			t.Fatal("malformed response, could not assert at time.Time")
		}
		returnedFormatted := returned.Format("2006-01-02")

		assert.Equal(t, sampleFormatted, returnedFormatted, "Time string mismatch")
		assert.True(t, sample.Equal(returned), "Time object mismatch")
	})

	t.Run("LocalTime", func(t *testing.T) {
		sample := time.Date(1970, 1, 1, 4, 19, 59, 999999999, time.UTC)
		sampleFormatted := sample.Format("15:04:05.000000000")

		all, _, err := conn.QueryWithDb("RETURN localtime('"+sampleFormatted+"')", nil, "")
		if err != nil {
			t.Fatal(err)
		}

		returned, ok := all[0][0].(time.Time)
		if !ok {
			t.Fatal("malformed response, could not assert at time.Time")
		}
		returnedFormatted := returned.Format("15:04:05.000000000")

		assert.Equal(t, sampleFormatted, returnedFormatted, "Time string mismatch")
		assert.True(t, sample.Equal(returned), "Time object mismatch")
	})

	t.Run("Time", func(t *testing.T) {
		loc, err := time.LoadLocation("Pacific/Guam")
		if err != nil {
			t.Fatal(err)
		}
		sample := time.Date(1970, 1, 1, 4, 19, 59, 999999999, loc)
		sampleFormatted := sample.Format("15:04:05.000000000Z07:00")

		all, _, err := conn.QueryWithDb("RETURN time('"+sampleFormatted+"')", nil, "")
		if err != nil {
			t.Fatal(err)
		}

		returned, ok := all[0][0].(time.Time)
		if !ok {
			t.Fatal("malformed response, could not assert at time.Time")
		}
		returnedFormatted := returned.Format("15:04:05.000000000Z07:00")

		assert.Equal(t, sampleFormatted, returnedFormatted, "Time string mismatch")
		assert.True(t, sample.Equal(returned), "Time object mismatch")
	})

	t.Run("Now DateTimeWithZoneOffset", func(t *testing.T) {
		now := time.Now()
		nowFormatted := now.Format(time.RFC3339Nano)

		all, _, err := conn.QueryWithDb("RETURN datetime('"+nowFormatted+"')", nil, "")
		if err != nil {
			t.Fatal(err)
		}

		returned, ok := all[0][0].(time.Time)
		if !ok {
			t.Fatal("malformed response, could not assert as time.Time")
		}
		returnedFormatted := returned.Format(time.RFC3339Nano)

		assert.Equal(t, nowFormatted, returnedFormatted, "Time string mismatch")
		assert.True(t, now.Equal(returned), "Time object mismatch")
	})

	t.Run("DateTimeWithZoneOffset", func(t *testing.T) {
		loc, err := time.LoadLocation("Japan")
		if err != nil {
			t.Fatal(err)
		}
		sample := time.Date(1995, 10, 4, 12, 32, 18, 1600, loc)
		sampleFormatted := sample.Format(time.RFC3339Nano)

		all, _, err := conn.QueryWithDb("RETURN datetime('"+sampleFormatted+"')", nil, "")
		if err != nil {
			t.Fatal(err)
		}

		returned, ok := all[0][0].(time.Time)
		if !ok {
			t.Fatal("malformed response, could not assert as time.Time")
		}
		returnedFormatted := returned.Format(time.RFC3339Nano)

		assert.Equal(t, sampleFormatted, returnedFormatted, "Time string mismatch")
		assert.True(t, sample.Equal(returned), "Time object mismatch")
	})

	t.Run("DateTimeWithZoneId", func(t *testing.T) {
		loc, err := time.LoadLocation("Europe/Samara")
		if err != nil {
			t.Fatal(err)
		}
		sample := time.Date(2019, 5, 4, 10, 00, 45, 1600, loc)
		sampleFormatted := sample.Format("2006-01-02T15:04:05.999999999")

		all, _, err := conn.QueryWithDb("RETURN datetime('"+sampleFormatted+"["+loc.String()+"]')", nil, "")
		if err != nil {
			t.Fatal(err)
		}

		returned, ok := all[0][0].(time.Time)
		if !ok {
			t.Fatal("malformed response, could not assert as time.Time")
		}
		returnedFormatted := returned.Format(time.RFC3339Nano)

		assert.Equal(t, sample.Format(time.RFC3339Nano), returnedFormatted, "Time string mismatch")
		assert.True(t, sample.Equal(returned), "Time object mismatch")
	})

	t.Run("LocalDateTime", func(t *testing.T) {
		sample := time.Date(2020, 3, 29, 16, 34, 18, 1600, time.UTC)
		sampleFormatted := sample.Format("2006-01-02T15:04:05.999999999")

		all, _, err := conn.QueryWithDb("RETURN localdatetime('"+sampleFormatted+"')", nil, "")
		if err != nil {
			t.Fatal(err)
		}

		returned, ok := all[0][0].(time.Time)
		if !ok {
			t.Fatal("malformed response, could not assert as time.Time")
		}
		returnedFormatted := returned.Format("2006-01-02T15:04:05.999999999")

		assert.Equal(t, sampleFormatted, returnedFormatted, "Time string mismatch")
		assert.True(t, sample.Equal(returned), "Time object mismatch")
	})
}

func TestEncoderTimeIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}

	// setup

	client, err := NewClient(WithBasicAuth("neo4j", "changeme"),
		WithHostPort("0.0.0.0", 7687))
	if err != nil {
		t.Fatal(err)
	}

	driver, err := client.NewDriver()
	if err != nil {
		t.Fatal(err)
	}

	conn, err := driver.Open(bolt_mode.WriteMode)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = conn.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	// tests

	t.Run("Now Native Time", func(t *testing.T) {
		now := time.Now()

		all, _, err := conn.QueryWithDb("RETURN $now", map[string]interface{}{
			"now": now,
		}, "")
		if err != nil {
			t.Fatal(err)
		}

		returned, ok := all[0][0].(time.Time)
		if !ok {
			t.Fatal("malformed response, could not assert at time.Time")
		}

		assert.Equal(t, now.Format(time.RFC3339Nano), returned.Format(time.RFC3339Nano))
		assert.True(t, now.Equal(returned), "Time object mismatch")
	})
}
