package helpers

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/zetkey/waka3x/models"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Heartbeat{},
		&models.Summary{},
		&models.SummaryItem{},
		&models.Alias{},
		&models.Duration{},
		&models.ApiKey{},
		&models.LanguageMapping{},
		&models.ProjectLabel{},
		&models.LeaderboardItem{},
		&models.WebAuthnCredential{},
		&models.Diagnostics{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// CleanupTestDB closes the database connection
func CleanupTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()

	sqlDB, err := db.DB()
	if err != nil {
		t.Errorf("Failed to get database connection: %v", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		t.Errorf("Failed to close database: %v", err)
	}
}

// NewTestUser creates a test user with default values
func NewTestUser(username, email string) *models.User {
	return &models.User{
		ID:          username,
		Email:       email,
		ApiKey:      "test-api-key-" + username,
		Password:    "$2a$10$test.hashed.password",
		CreatedAt:   models.CustomTime(time.Now()),
		HasData:     true,
		IsAdmin:     false,
		AuthType:    "local",
		StartOfWeek: 1,
	}
}

// NewTestHeartbeat creates a test heartbeat
func NewTestHeartbeat(userID string, timestamp time.Time) *models.Heartbeat {
	return &models.Heartbeat{
		UserID:          userID,
		Entity:          "/path/to/file.go",
		Type:            "file",
		Category:        "coding",
		Project:         "test-project",
		Branch:          "main",
		Language:        "Go",
		IsWrite:         true,
		Editor:          "VS Code",
		OperatingSystem: "Linux",
		Machine:         "test-machine",
		Time:            models.CustomTime(timestamp),
		Hash:            "test-hash-" + timestamp.Format(time.RFC3339),
	}
}

// NewTestHeartbeats creates multiple test heartbeats
func NewTestHeartbeats(userID string, count int) []*models.Heartbeat {
	heartbeats := make([]*models.Heartbeat, count)
	baseTime := time.Now()

	for i := 0; i < count; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute)
		heartbeats[i] = NewTestHeartbeat(userID, timestamp)
	}

	return heartbeats
}

// NewTestSummary creates a test summary
func NewTestSummary(userID string, from, to time.Time) *models.Summary {
	return &models.Summary{
		UserID:           userID,
		FromTime:         models.CustomTime(from),
		ToTime:           models.CustomTime(to),
		Projects:         []*models.SummaryItem{},
		Languages:        []*models.SummaryItem{},
		Editors:          []*models.SummaryItem{},
		OperatingSystems: []*models.SummaryItem{},
		Machines:         []*models.SummaryItem{},
		NumHeartbeats:    0,
	}
}

// NewTestAlias creates a test alias
func NewTestAlias(userID string, aliasType uint8, key, value string) *models.Alias {
	return &models.Alias{
		UserID: userID,
		Type:   aliasType,
		Key:    key,
		Value:  value,
	}
}

// AssertNoError asserts that an error is nil
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	assert.NoError(t, err)
}

// AssertError asserts that an error is not nil
func AssertError(t *testing.T, err error) {
	t.Helper()
	assert.Error(t, err)
}

// AssertErrorContains asserts that an error contains a specific message
func AssertErrorContains(t *testing.T, err error, message string) {
	t.Helper()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), message)
}

// AssertEqual asserts that two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	assert.Equal(t, expected, actual)
}

// AssertNotEqual asserts that two values are not equal
func AssertNotEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	assert.NotEqual(t, expected, actual)
}

// AssertNil asserts that a value is nil
func AssertNil(t *testing.T, value interface{}) {
	t.Helper()
	assert.Nil(t, value)
}

// AssertNotNil asserts that a value is not nil
func AssertNotNil(t *testing.T, value interface{}) {
	t.Helper()
	assert.NotNil(t, value)
}

// AssertTrue asserts that a condition is true
func AssertTrue(t *testing.T, condition bool) {
	t.Helper()
	assert.True(t, condition)
}

// AssertFalse asserts that a condition is false
func AssertFalse(t *testing.T, condition bool) {
	t.Helper()
	assert.False(t, condition)
}

// AssertLen asserts that a collection has a specific length
func AssertLen(t *testing.T, collection interface{}, length int) {
	t.Helper()
	assert.Len(t, collection, length)
}

// AssertEmpty asserts that a collection is empty
func AssertEmpty(t *testing.T, collection interface{}) {
	t.Helper()
	assert.Empty(t, collection)
}

// AssertNotEmpty asserts that a collection is not empty
func AssertNotEmpty(t *testing.T, collection interface{}) {
	t.Helper()
	assert.NotEmpty(t, collection)
}

// AssertContains asserts that a string contains a substring
func AssertContains(t *testing.T, str, substr string) {
	t.Helper()
	assert.Contains(t, str, substr)
}

// CreateTestUsers creates multiple test users in the database
func CreateTestUsers(t *testing.T, db *gorm.DB, count int) []*models.User {
	t.Helper()

	users := make([]*models.User, count)
	for i := 0; i < count; i++ {
		user := NewTestUser(
			"testuser"+string(rune('0'+i)),
			"test"+string(rune('0'+i))+"@example.com",
		)
		if err := db.Create(user).Error; err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}
		users[i] = user
	}

	return users
}

// CreateTestHeartbeats creates multiple test heartbeats in the database
func CreateTestHeartbeats(t *testing.T, db *gorm.DB, userID string, count int) []*models.Heartbeat {
	t.Helper()

	heartbeats := NewTestHeartbeats(userID, count)
	for _, hb := range heartbeats {
		if err := db.Create(hb).Error; err != nil {
			t.Fatalf("Failed to create test heartbeat: %v", err)
		}
	}

	return heartbeats
}

// TruncateTable truncates a table in the test database
func TruncateTable(t *testing.T, db *gorm.DB, model interface{}) {
	t.Helper()

	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(model).Error; err != nil {
		t.Fatalf("Failed to truncate table: %v", err)
	}
}
