# Waka3x Background Jobs

**AI Agent Reference Document**  
Last Updated: 2026-05-17

## Overview

Waka3x runs several background jobs for data aggregation, reporting, cleanup, and maintenance. Jobs are scheduled using cron expressions and run in separate goroutines.

## Job Scheduler

**Library**: `github.com/robfig/cron/v3`

**Configuration**: `config/jobqueue.go`

**Access**: `conf.Cron()` returns global cron instance

**Format**: Extended cron (6 fields: second, minute, hour, day, month, day-of-week)

## Core Background Jobs

### 1. Aggregation Job

**Purpose**: Generate daily summaries for all users

**Service**: `AggregationService`

**Schedule**: `app.aggregation_time` (default: `0 15 2 * * *` - 2:15 AM daily)

**Process**:
1. Fetch all active users (has_data = true)
2. For each user:
   - Get yesterday's date range
   - Check if summary already exists
   - If not, generate summary from heartbeats
   - Calculate durations
   - Apply aliases
   - Store summary in database
3. Log progress and errors

**Code Location**: `services/aggregation.go`

**Scheduling**:
```go
func (srv *AggregationService) Schedule() {
    config := conf.Get()
    _, err := conf.Cron().AddFunc(config.App.AggregationTime, func() {
        srv.Run()
    })
}
```

**Invocation**: `main.go` line 211
```go
go aggregationService.Schedule()
```

**Performance**: Processes users sequentially, ~1-2 seconds per user

### 2. Weekly Report Job

**Purpose**: Send weekly email reports to subscribed users

**Service**: `ReportService`

**Schedule**: `app.report_time_weekly` (default: `0 0 18 * * 5` - Friday 6 PM)

**Process**:
1. Fetch users with `reports_weekly = true`
2. For each user:
   - Generate summary for past 7 days
   - Format HTML email with statistics
   - Send email via SMTP
   - Log success/failure
3. Continue on individual failures

**Code Location**: `services/report.go`

**Email Content**:
- Total coding time
- Top projects
- Top languages
- Top editors
- Daily breakdown chart

**Requirements**: Mail must be configured (`mail.enabled = true`)

**Scheduling**:
```go
func (srv *ReportService) Schedule() {
    config := conf.Get()
    _, err := conf.Cron().AddFunc(config.App.ReportTimeWeekly, func() {
        srv.SendWeekly()
    })
}
```

### 3. Housekeeping Job

**Purpose**: Data cleanup and database maintenance

**Service**: `HousekeepingService`

**Schedule**: `app.data_cleanup_time` (default: `0 0 6 * * 0` - Sunday 6 AM)

**Tasks**:
1. **Delete old heartbeats** (if `data_retention_months > 0`)
   - Calculate cutoff date
   - Delete heartbeats older than cutoff
   - Delete associated durations
2. **Delete inactive users** (if `max_inactive_months > 0`)
   - Find users not logged in for N months
   - Delete user and all associated data (cascade)
3. **Optimize database**
   - SQLite: VACUUM
   - PostgreSQL: VACUUM ANALYZE
   - MySQL: OPTIMIZE TABLE

**Code Location**: `services/housekeeping.go`

**Scheduling**:
```go
func (srv *HousekeepingService) Schedule() {
    config := conf.Get()
    _, err := conf.Cron().AddFunc(config.App.DataCleanupTime, func() {
        srv.CleanUp()
    })
}
```

### 4. Leaderboard Job

**Purpose**: Calculate and cache leaderboard rankings

**Service**: `LeaderboardService`

**Schedule**: `app.leaderboard_generation_time` (default: `0 0 6 * * *,0 0 18 * * *` - 6 AM and 6 PM daily)

**Process**:
1. Determine time range (e.g., last 7 days)
2. Fetch all users with `public_leaderboard = true`
3. For each user:
   - Calculate total coding time for period
   - Store in leaderboard table
4. Rank users by total time
5. Cache results

**Code Location**: `services/leaderboard.go`

**Enabled**: Only if `app.leaderboard_enabled = true`

**Scheduling**:
```go
func (srv *LeaderboardService) Schedule() {
    config := conf.Get()
    _, err := conf.Cron().AddFunc(config.App.LeaderboardGenerationTime, func() {
        srv.Compute()
    })
}
```

### 5. Database Optimization Job

**Purpose**: Periodic database maintenance

**Service**: `HousekeepingService` (part of housekeeping)

**Schedule**: `app.optimize_database_time` (default: `0 0 8 1 * *` - 1st of month at 8 AM)

**Tasks**:
- **SQLite**: VACUUM (reclaim space, defragment)
- **PostgreSQL**: VACUUM ANALYZE (reclaim space, update statistics)
- **MySQL**: OPTIMIZE TABLE (defragment, update indexes)

**Code Location**: `services/housekeeping.go`

### 6. Miscellaneous Jobs

**Service**: `MiscService`

**Purpose**: Various utility tasks

**Tasks**:
- Cache warming
- Statistics updates
- Cleanup temporary data

**Code Location**: `services/misc.go`

**Scheduling**:
```go
func (srv *MiscService) Schedule() {
    // Custom schedules for various tasks
}
```

## Job Initialization

**Location**: `main.go` (lines 210-218)

**Order**:
1. Start global cron scheduler: `conf.StartJobs()`
2. Schedule service jobs:
   ```go
   go aggregationService.Schedule()
   go reportService.Schedule()
   go housekeepingService.Schedule()
   go miscService.Schedule()
   if config.App.LeaderboardEnabled {
       go leaderboardService.Schedule()
   }
   ```

**Goroutines**: Each `Schedule()` call runs in separate goroutine

## Cron Expression Examples

```
# Every 5 minutes
0 */5 * * * *

# Daily at 2:15 AM
0 15 2 * * *

# Every Friday at 6 PM
0 0 18 * * 5

# Every Sunday at 6 AM
0 0 6 * * 0

# Twice daily (6 AM and 6 PM)
0 0 6 * * *,0 0 18 * * *

# First of month at 8 AM
0 0 8 1 * *
```

## Monitoring Jobs

### Logging

All jobs log start, progress, and completion:

```go
conf.Log().Info("starting aggregation job")
// ... work ...
conf.Log().Info("aggregation job completed", "users", count)
```

**Log Level**: Info for normal operation, Error for failures

### Error Handling

Jobs continue on individual failures:

```go
for _, user := range users {
    if err := processUser(user); err != nil {
        conf.Log().Error("failed to process user", "user", user.ID, "error", err)
        continue  // Don't stop entire job
    }
}
```

### Sentry Integration

Errors automatically reported to Sentry (if configured):

```go
if config.Sentry.Dsn != "" {
    sentry.CaptureException(err)
}
```

## Job Configuration

### Disable Jobs

Set schedule to empty string or invalid cron:

```yaml
app:
  aggregation_time: ""  # Disable aggregation
```

### Adjust Timing

Modify cron expressions:

```yaml
app:
  aggregation_time: '0 0 3 * * *'  # Run at 3 AM instead
  report_time_weekly: '0 0 9 * * 1'  # Monday 9 AM instead
```

### Conditional Jobs

Some jobs only run if enabled:

```yaml
app:
  leaderboard_enabled: false  # Disables leaderboard job
mail:
  enabled: false  # Disables report emails
```

## Performance Considerations

### Aggregation Job

- **Duration**: ~1-2 seconds per user
- **Database Load**: High (reads heartbeats, writes summaries)
- **Optimization**: Run during low-traffic hours (2-4 AM)

### Report Job

- **Duration**: ~1-2 seconds per user (email sending)
- **External Dependency**: SMTP server
- **Failure Handling**: Individual failures don't stop job

### Housekeeping Job

- **Duration**: Varies (depends on data volume)
- **Database Load**: Very high (deletes, vacuum)
- **Optimization**: Run during lowest traffic (Sunday morning)

### Leaderboard Job

- **Duration**: ~5-10 seconds (all users)
- **Database Load**: Medium (reads summaries)
- **Frequency**: Twice daily (sufficient for most use cases)

## Troubleshooting

### Job not running

- Check cron expression syntax
- Verify service is scheduled in `main.go`
- Check logs for errors
- Ensure config loaded correctly

### Job running too slowly

- Check database indexes
- Reduce batch size
- Optimize queries
- Consider running less frequently

### Job failing

- Check logs for error messages
- Verify database connection
- Check external dependencies (SMTP for reports)
- Review Sentry for stack traces

### Database locks

- Jobs may conflict with high write traffic
- Use transactions carefully
- Consider read replicas for heavy queries

## Manual Job Execution

Jobs can be triggered manually via API (admin only):

```bash
# Trigger aggregation
curl -X POST http://localhost:3000/api/admin/jobs/aggregation \
  -H "Authorization: Basic $(echo -n 'admin-api-key' | base64)"
```

**Note**: Manual execution endpoints not currently implemented

## Future Enhancements

- Job status dashboard
- Manual job triggering via UI
- Job history and logs
- Configurable retry logic
- Job queue with priority
- Distributed job execution
