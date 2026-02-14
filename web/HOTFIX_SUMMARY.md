# SniffOps Web UI Hotfix Summary
**Date:** 2026-02-14  
**Commit:** c0f5797  
**Author:** Pixel <pixel@sniffops.dev>

## Changes Applied

### ✅ 1. Timestamp Format Fixed
**File:** `src/components/traces/TracesColumns.tsx`

- **Before:** `formatDistanceToNow(new Date(timestamp * 1000), { addSuffix: true })`
  - Displayed: "in almost 56066 years" (bug caused by double conversion)
- **After:** `format(new Date(timestamp), 'yyyy-MM-dd HH:mm:ss')`
  - Displays: "2026-02-14 11:35:23" (absolute timestamp)
- **Changes:**
  - Removed `* 1000` multiplication (timestamp already in ms)
  - Switched from `formatDistanceToNow` to `format` from date-fns
  - Added `font-mono` and `whitespace-nowrap` for better readability

### ✅ 2. Table Optimized - No Horizontal Scroll
**File:** `src/components/traces/TracesColumns.tsx`

- **Removed Column:**
  - Command column (visible in detail modal instead)
  
- **Optimized Column Widths:**
  - Tool: `max-w-[120px]` with truncate
  - Namespace: `max-w-[100px]` with truncate, smaller text (`text-xs`)
  - Resource: `max-w-[180px]` with truncate
  - Latency: Added `font-mono whitespace-nowrap`

- **Final Column Order:**
  1. Time (sortable)
  2. Risk (badge with icon)
  3. Tool (sortable)
  4. Namespace (badge)
  5. Resource (truncated)
  6. Status (success/error badge)
  7. Latency (sortable)

### ✅ 3. Sidebar Mini Mode Enabled
**File:** `src/components/layout/AppSidebar.tsx`

- **Before:** `<Sidebar>` (default: `collapsible="offcanvas"`)
  - Sidebar completely disappeared when collapsed
- **After:** `<Sidebar collapsible="icon">`
  - Sidebar shows icons when collapsed (mini mode)
  - Width changes from 16rem to 3rem when collapsed
  - Tooltips appear on hover in collapsed state

## Build Verification

```bash
npm run build
```

**Result:** ✅ Success
- TypeScript compilation: ✅ Passed
- Vite build: ✅ Completed in 22.51s
- Output size:
  - HTML: 0.48 kB (gzipped: 0.31 kB)
  - CSS: 44.15 kB (gzipped: 7.96 kB)
  - JS: 489.62 kB (gzipped: 151.64 kB)

## Testing Checklist

- [ ] Verify timestamp shows as "YYYY-MM-DD HH:mm:ss" format
- [ ] Confirm no horizontal scroll on traces table
- [ ] Check all 7 columns fit in viewport width
- [ ] Test sidebar collapse/expand with toggle
- [ ] Verify icons visible when sidebar collapsed
- [ ] Confirm tooltips show in collapsed sidebar
- [ ] Open trace detail sheet and verify Command is visible there

## Notes

- The sidebar component uses ShadCN's sidebar with built-in `collapsible="icon"` support
- The `SidebarRail` component provides the drag-to-resize functionality
- Tooltips automatically show when sidebar is collapsed (`state !== "collapsed"`)
