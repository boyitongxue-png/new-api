package model

import (
	"net/url"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestNormalizeSQLiteDSNAppliesConcurrencySettings(t *testing.T) {
	tests := []struct {
		name             string
		dsn              string
		wantBusyTimeout  string
		wantJournalMode  string
		wantTransaction  string
		wantLegacyAbsent bool
	}{
		{
			name:             "plain path gets safe defaults",
			dsn:              "one-api.db",
			wantBusyTimeout:  "busy_timeout(30000)",
			wantJournalMode:  "journal_mode(WAL)",
			wantTransaction:  "immediate",
			wantLegacyAbsent: true,
		},
		{
			name:             "legacy timeout is preserved using supported syntax",
			dsn:              "data.db?_busy_timeout=7000&cache=shared",
			wantBusyTimeout:  "busy_timeout(7000)",
			wantJournalMode:  "journal_mode(WAL)",
			wantTransaction:  "immediate",
			wantLegacyAbsent: true,
		},
		{
			name:             "explicit settings are preserved",
			dsn:              "data.db?_pragma=busy_timeout(9000)&_pragma=journal_mode(DELETE)&_txlock=exclusive",
			wantBusyTimeout:  "busy_timeout(9000)",
			wantJournalMode:  "journal_mode(DELETE)",
			wantTransaction:  "exclusive",
			wantLegacyAbsent: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			normalized, err := normalizeSQLiteDSN(test.dsn)
			require.NoError(t, err)
			parsed, err := url.Parse(normalized)
			require.NoError(t, err)
			query := parsed.Query()

			assert.Contains(t, query["_pragma"], test.wantBusyTimeout)
			assert.Contains(t, query["_pragma"], test.wantJournalMode)
			assert.Equal(t, test.wantTransaction, query.Get("_txlock"))
			if test.wantLegacyAbsent {
				assert.Empty(t, query.Get("_busy_timeout"))
			}
		})
	}
}

func TestNormalizedSQLiteDSNSettingsAreAppliedByDriver(t *testing.T) {
	dsn, err := normalizeSQLiteDSN(filepath.Join(t.TempDir(), "settings.db"))
	require.NoError(t, err)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	var busyTimeout int
	require.NoError(t, db.Raw("PRAGMA busy_timeout").Scan(&busyTimeout).Error)
	assert.Equal(t, 30000, busyTimeout)

	var journalMode string
	require.NoError(t, db.Raw("PRAGMA journal_mode").Scan(&journalMode).Error)
	assert.Equal(t, "wal", journalMode)
}
