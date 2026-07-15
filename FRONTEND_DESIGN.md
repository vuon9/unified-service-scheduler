# FRONTEND DESIGN — Unified Service Scheduler

> React + Vite demo UI. Calendar-style views with full mobile responsiveness.
> Date: 2026-07-15 (v3 — reflects actual implementation)

---

## Design Goals

- **Video-ready**: every screen designed to look good in a 5-10 min walkthrough
- **API-driven**: all data from Go backend (`localhost:8080/api/v1`)
- **No router**: single-page with tab + view-mode navigation; state in URL via `?view=week&date=...&tech=t1&bay=b12`
- **No state management lib**: React hooks + fetch is sufficient
- **Mobile-first responsive**: same app works on desktop and phone

---

## Screen 1: Appointments Dashboard

```
┌──────────────────────────────────────────────────────────┐
│ [K] Unified Service Scheduler                      [+ New Booking]│  ← sticky header
├──────────────────────────────────────────────────────────┤
│ [Confirmed] [Cancelled]                        [🗘 filter]│  ← tab bar
├──────────────────────────────────────────────────────────┤
│ [Timeline] [Week] [Month]        < Jul 15, 2026 > [Today]│  ← view controls
├──────────────┬─────────────────────────────────────────────┤
│ Sidebar      │                                            │
│ ──────────── │  ┌── Today ────────────────────────────┐  │
│ TECHNICIANS  │  │  08:00  Oil Change                  │  │
│ ⬤ Minh   3   │  │         Toyota Camry · Minh · Bay 1A│  │
│ ⬤ Hai    1   │  │  09:00  Brake Replacement           │  │
│ ⬤ Nam    2   │  │         Mazda CX-5 · Nam · Bay 2A   │  │
│ ──────────── │  └────────────────────────────────────┘  │
│ SERVICE BAYS│                                            │
│ ■ Bay 1A  2  │  (Timeline view — shown)                  │
│ ■ Bay 2A  1  │                                            │
│ ■ Bay 3A  0  │                                            │
└──────────────┴────────────────────────────────────────────┘

  Week view: 7-column grid, horizontal scroll on mobile
  Month view: calendar grid, click day to jump to date

  Mobile (<768px): sidebar hidden; toggle via filter icon 🗘
  Filters: AND logic — select tech AND bay to narrow down
```

### Views

| View | Description |
|------|-------------|
| **Timeline** | Vertical card list grouped by day. Each card shows time, service, vehicle, tech name, bay name. |
| **Week** | 7-column grid (Mon-Sun). `minmax(110px, 1fr)` columns, horizontal scroll on mobile. |
| **Month** | Calendar grid. Click a day to jump to that date in any view. |

### States

- **Empty**: "No confirmed appointments" + CTA text
- **Loading**: Spinner + "Loading appointments..."
- **Error**: Red error message + Retry button
- **Filtered**: Count badge shows matching results

---

## Screen 2: Booking Modal

```
Desktop:                              Mobile (slide-up bottom sheet):
┌─────────────────────────────┐       ┌─────────────────────────┐
│ Book Appointment        [X] │       │ Book Appointment    [X] │
├─────────────────────────────┤       ├─────────────────────────┤
│ Vehicle       Service       │       │ Vehicle                  │
│ [Toyota ▼]  [Oil Chg ▼]   │       │ [Toyota Camry ▼]         │
│                             │       │ Service                  │
│ Date & Time     (60 min)   │       │ [Oil Change ▼]           │
│ [2026-07-20 09:00     ]    │       │                         │
│                             │       │ Date & Time              │
│ Technician(s)               │       │ [2026-07-20 09:00 ] (60m)│
│ ┌────────┐ ┌────────┐      │       │                         │
│ │┄ Minh ┤│ │  Hai   │      │       │ Technician(s)            │
│ └────────┘ └────────┘      │       │ ┌────────┐ ┌────────┐  │
│ Service Bay(s)              │       │ │┄ Minh ┤│ │  Hai   │  │
│ ┌────────┐ ┌────────┐      │       │ └────────┘ └────────┘  │
│ │┄ Bay1A ┤│ │ Bay2A  │      │       │ Service Bay(s)          │
│ └────────┘ └────────┘      │       │ ┌────────┐ ┌────────┐  │
│                             │       │ │┄ Bay1A ┤│ │ Bay2A  │  │
│ Notes (optional)           │       │ └────────┘ └────────┘  │
│ ┌───────────────────────┐  │       │                         │
│ │ Please check tires... │  │       │ Notes (optional)         │
│ └───────────────────────┘  │       │ ┌─────────────────────┐ │
│                             │       │ │ Please check tires..│ │
│ [Cancel]   [Book Appt]     │       │ └─────────────────────┘ │
└─────────────────────────────┘       │                         │
                                      │ [Cancel] [Book Appt]    │
                                      └─────────────────────────┘

Badge legend:
  ┄┄┄ dotted border = suggested (auto-selected first available)
  ─── solid border = user-selected
  ─── gray       = not selected
```

### Booking Flow

1. Select Vehicle + Service (2-column grid)
2. Pick Date & Time (`datetime-local` input)
3. **Auto-check**: 500ms debounce → API call → badges appear
4. Click badges to select specific tech/bay (first one auto-selected with dotted border)
5. Optional: add notes
6. Click "Book Appointment" — sends `technician_id` + `service_bay_id` if selected, backend validates availability
7. On success: modal closes, list auto-refreshes

---

## Screen 3: Appointment Detail

```
┌──────────────────────────────────────┐
│ Appointment Detail                [X] │
├──────────────────────────────────────┤
│ STATUS: CONFIRMED                    │
│                                      │
│ Vehicle      Toyota Camry            │
│ Service      Oil Change              │
│ Date         Tuesday, July 15, 2026  │
│ Time         08:00 - 09:00           │
│ Customer     Anh Tuan                │
│ Technician   Minh                    │
│ Bay          Bay 1A                  │
│ Booked at    Jul 15, 2026, 08:00     │
│                                      │
│ Notes                                │
│ Please check tire pressure           │
│ Also rotate tires                    │
│                                      │
│ [Cancel Appointment]                 │
└──────────────────────────────────────┘
```

---

## Screen 4: Live Metrics

| Card | Value | Source |
|------|-------|--------|
| Bookings | `bookings_total` | `GET /debug/vars` (expvar) |
| Conflicts | `bookings_conflicts` | Auto-refresh every 3s |
| Cancellations | `cancellations_total` | Dark theme, green/red/amber |
| Vehicle | `vehicle_conflicts` | |
| Tech | `tech_conflicts` | |
| Bay | `bay_conflicts` | |
| Memory | `memstats.Alloc` (MB) | |

Static HTML file at `/metrics.html` — no React, fetches directly from proxy.

---

## Component Tree

```
App
├── Header (sticky)
│   ├── Logo (K icon + "Unified Service Scheduler")
│   └── New Booking button
├── TabBar (sticky)
│   ├── Confirmed tab
│   ├── Cancelled tab
│   └── Filter toggle icon (mobile)
├── ViewControls (sticky)
│   ├── Timeline / Week / Month buttons
│   ├── Date navigation (< date >)
│   └── Today button
├── Main Content
│   ├── Sidebar (desktop 220px / mobile overlay)
│   │   ├── Technicians section (colored dots + counts)
│   │   └── Service Bays section (square dots + counts)
│   └── Active View
│       ├── TimelineView (vertical cards grouped by day)
│       ├── WeekView (7-column grid)
│       └── MonthView (calendar grid)
├── BookingModal
│   ├── Vehicle + Service selects
│   ├── Datetime-local input
│   ├── Technician badges (clickable, dotted = suggested)
│   ├── Bay badges (clickable, dotted = suggested)
│   ├── Notes textarea
│   └── Cancel + Book buttons
└── DetailModal
    ├── Status badge
    ├── Detail rows (label: value)
    ├── Notes (multiline, pre-wrap)
    └── Cancel button + confirmation
```

## URL State

View state is synced to `window.location` via `replaceState`:

```
/?view=week&date=2026-07-15&tech=t1&bay=b12
```

| Param | Values | Default |
|-------|--------|---------|
| `view` | `timeline`, `week`, `month` | `timeline` |
| `date` | `YYYY-MM-DD` (local timezone) | today |
| `tech` | technician ID | (none) |
| `bay` | service bay ID | (none) |
| `tab` | `confirmed`, `cancelled` | `confirmed` |

Reads from URL on mount, writes on state change. Invalid values fall back to defaults.

---

## API Integration

| Function | Endpoint | Used in |
|----------|----------|---------|
| `fetchAppointments(params?)` | `GET /api/v1/appointments` | All views |
| `fetchVehicles()` | `GET /api/v1/vehicles` | BookingModal dropdown |
| `fetchServiceTypes()` | `GET /api/v1/service-types` | BookingModal dropdown |
| `fetchTechnicians()` | `GET /api/v1/technicians` | Sidebar |
| `fetchServiceBays()` | `GET /api/v1/service-bays` | Sidebar |
| `checkAvailability(data)` | `POST /api/v1/availability` | BookingModal auto-check |
| `createAppointment(data)` | `POST /api/v1/appointments` | BookingModal submit |
| `cancelAppointment(id)` | `POST /api/v1/appointments/{id}/cancel` | DetailModal |

All go through `apiFetch<T>()` wrapper. `createAppointment` now sends optional `technician_id` + `service_bay_id`.

---

## Styling

- **CSS Modules**: Scoped styles (`*.module.css`)
- **Light theme**: White + gray-50 backgrounds, blue-600 primary
- **Font sizes**: 16px base (prevents iOS zoom on focus)
- **Consistent inputs**: 44px height, `box-sizing: border-box`
- **Badge styles**: Dotted border (suggested), solid blue (selected), gray (inactive)
- **Safari datetime**: Calendar icon renders natively (removed `inline-flex` override)

## Seed Data

2 dealerships (Saigon, Ha Noi), 3 customers, 3 vehicles, 4 service types (Oil Change 60m, Brake 120m, Engine Diagnostic 90m, Tire Rotation 30m), 5 technicians, 5 bays (b11/b12/b13 at d1, b21/b22 at d2). Qualifications varied per tech/dealership.

## Known Issues

- **No pagination**: All appointments loaded at once. Fine for demo.
- **Safari datetime input**: Calendar icon renders differently than Chrome, but functional.
