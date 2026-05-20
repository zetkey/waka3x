# Alias Resolution Fix

**Date**: 2026-05-20
**Branch**: `refactor/vue-frontend`
**Scope**: Backend cache invalidation, projects API, settings UI

## Summary

Aliases configured in `SettingsView.vue` (Project / Language / Editor / OS / Machine) were not being applied to data shown in `HomeView.vue` and `SummaryView.vue`. They were also never applied at all in `ProjectsView.vue`. This document explains the root causes and the fix that landed on this branch.

## Symptoms

1. After adding a project alias, the canonical key still did not appear in the Weekly Report (HomeView) or in the Summary view, even after a full page reload.
2. After adding an OS alias, the same applied to the OS breakdown.
3. ProjectsView always showed the raw project names from heartbeats, regardless of any alias configuration.
4. SettingsView listed configured aliases without showing which entity type each one belonged to, making same-key aliases on different types visually indistinguishable.

## Root Causes

### 1. Stale summary cache (HomeView, SummaryView)

`services/summary.go` exposes `Aliased(from, to, user, filters, customTimeout, skipCache)`. It caches the post-aliased summary by:

```
hash = from.String() + to.String() + user.ID + filters.Hash() + timeout + "--aliased"
```

The cache entry stores the already-resolved `*models.Summary` object — meaning the alias map was applied at write time and frozen into the cache value.

Cache invalidation existed for:

- Heartbeat creation (`config.EventHeartbeatCreate`)
- Project label create/delete (`config.TopicProjectLabel`)
- API key create/delete (`services/api_key.go`)
- Manual summary deletion (`DeleteByUser`, `DeleteByUserAfter`, etc.)

But there was **no invalidation when an alias was created or deleted**. `services/alias.go` updated only its in-memory `userAliases sync.Map` and the DB row. The next request hit the cached pre-alias summary.

This was visible in HomeView because it always sends `from`/`to` as `YYYY-MM-DD` strings, which the backend parses to exact midnight in the user's timezone — a stable cache key that hits the cache forever. SummaryView had the same issue for fixed ranges.

### 2. Aliases never applied to /api/projects (ProjectsView)

`routes/api/projects.go` called `heartbeatService.GetUserProjectStats` and serialized the rows directly. There was no `AliasService` dependency at all, so the projects grid showed raw heartbeat project names. Two raw projects that should collapse into the same alias key (e.g. `wakapi-web`, `wakapi-mobile` → `wakapi`) appeared as separate cards.

### 3. SettingsView alias list lacked the type

The list at `SettingsView.vue` rendered each alias as `{{ alias.key }}` with its mapped values, but never showed `alias.type`. Adding a Project alias and an OS alias with the same key produced two rows that looked identical, and there was no way to tell them apart short of trial-and-error deletion.

## Fix

### Backend

1. **Event bus topics** — `config/eventbus.go:13`
   Added `TopicAlias`, `EventAliasCreate`, `EventAliasDelete`. The naming and "topic-with-wildcard plus discrete events" pattern follows the existing `TopicProjectLabel` / `EventProjectLabel*` precedent so the cache invalidator can subscribe to the topic regardless of whether it's a create or delete.

2. **AliasService publishes events** — `services/alias.go`
   - Injected `*hub.Hub` into the struct (mirroring `ProjectLabelService`).
   - `Create`, `Delete`, and `DeleteMulti` each call a new `notifyUpdate` helper that publishes to `EventAliasCreate` / `EventAliasDelete` with `FieldUserId` set.
   - Both manual cache update and async `MayInitializeUser` reload were preserved; the publish is added on top.

3. **SummaryService subscribes to alias events** — `services/summary.go:56`
   A new subscription on `config.TopicAlias` runs `srv.invalidateUserCache(userID)` for every alias change. This mirrors line-for-line the existing `TopicProjectLabel` subscription, so the fix can be reasoned about by analogy.

4. **Projects API resolves and merges aliases** — `routes/api/projects.go`
   - Added `aliasService services.IAliasService` dependency (wired in `main.go:235`).
   - Fetches the **full** unfiltered project list, not a paginated slice, because aliases collapse multiple raw projects into one canonical entry. Pagination must happen *after* merging or pages would be misaligned.
   - For each raw `*ProjectStats`, calls `aliasService.GetAliasOrDefault(userID, SummaryProject, raw.Project)` to get the canonical key.
   - Merges entries that share a canonical key: sums `Count`, takes min `First` / max `Last`, and adopts the heaviest contributor's `TopLanguage` (the row with the largest count usually has the most representative language).
   - Search filter and pagination apply to the merged list.

### Frontend

`frontend/src/views/SettingsView.vue`:

1. Single source of truth for alias types — extracted to:
   ```ts
   const aliasTypes = [
     { label: "Project", value: 0 },
     { label: "Language", value: 1 },
     { label: "Editor", value: 2 },
     { label: "OS", value: 3 },
     { label: "Machine", value: 4 },
   ];
   function aliasTypeLabel(type: number) {
     return aliasTypes.find((t) => t.value === type)?.label || "Unknown";
   }
   ```
2. The "Add alias" `<Select>` iterates `aliasTypes` instead of hardcoding five `SelectItem`s.
3. Each alias row now renders a small uppercase chip (`bg-muted` rounded pill, same look as the project language chip on ProjectsView) showing `aliasTypeLabel(alias.type)` next to the canonical key.

## Files Changed

| File | Change |
| --- | --- |
| `config/eventbus.go` | Added `TopicAlias`, `EventAliasCreate`, `EventAliasDelete` |
| `services/alias.go` | Inject event bus; publish events on create/delete |
| `services/summary.go` | Subscribe to `TopicAlias`; invalidate cache on change |
| `routes/api/projects.go` | Add alias service dep; resolve + merge + paginate |
| `main.go` | Pass `aliasService` to `NewProjectsApiHandler` |
| `frontend/src/views/SettingsView.vue` | `aliasTypes` const, dynamic Select, type chip on rows |

## Verification

- `go build ./services/... ./routes/... ./config/...` — clean
- `go build ./main.go` — clean
- `go test ./services/ -run TestAliasServiceTestSuite` — 2/2 pass
- `cd frontend && npx vue-tsc --noEmit` — clean

Manual checks:

1. Add a project alias `wakapi` mapping `wakapi-mobile` in Settings.
2. Reload HomeView — the Most Active Project / Top Projects chart now uses `wakapi`, with combined totals from all aliased raw names.
3. Reload SummaryView — Projects breakdown and Environment Breakdown reflect the alias.
4. Reload ProjectsView — `wakapi-mobile` and any other matching raw projects collapse into a single `wakapi` card with summed heartbeats.
5. Delete the alias — caches invalidate immediately on the next request, raw names return without a server restart.

## Why the Cache Was Designed This Way

The post-aliased summary is cached because alias resolution touches every `SummaryItem` across eight entity types and merges duplicates. Doing that on every request would be wasteful, especially for the Weekly Report which fires off on every HomeView mount. Caching the resolved object keeps reads fast.

The trade-off is that alias *changes* must invalidate the cache. The original code only invalidated on writes that affect summaries directly (heartbeats, summaries, labels). Aliases were missed because they live one layer above the summary — they don't change the underlying data, only how it's presented.

The event-bus approach (rather than calling `invalidateUserCache` directly from `AliasService`) keeps service dependencies pointing one direction: services don't import each other, they publish to a hub. This mirrors how `ProjectLabelService` already worked.

## Related

- `docs/06-SERVICES-LAYER.md` — service responsibilities and boundaries
- `docs/05-API-ENDPOINTS.md` — `/api/summary`, `/api/summary/details`, `/api/projects`
- `services/summary.go:82` — `Aliased()` cache key construction
- `services/alias.go:75` — `GetAliasOrDefault` with wildcard match (used by both summary resolver and the new projects merging)
