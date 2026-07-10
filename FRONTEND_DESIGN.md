# FRONTEND DESIGN -- Unified Service Scheduler

> React + Vite demo UI. Calendar-style views with full mobile responsiveness.
> Date: 2026-07-10 (v2 -- reflects actual implementation)

---

## Design Goals

- **Video-ready**: every screen designed to look good in a 5-10 min walkthrough
- **API-driven**: all data from Go backend (`localhost:8080/api/v1`)
- **No router**: single-page with tab + view-mode navigation
- **No state management lib**: React hooks + fetch is sufficient
- **Mobile-first responsive**: same app works on desktop and phone

---

## Screen 1: Appointments Dashboard

```
┌──────────────────────────────────────────────────────┐
│ [K] Keyloop Scheduler                    [+ New]      │  ← sticky header
├──────────────────────────────────────────────────────┤
│ [Confirmed] [Cancelled]                   [🗘]        │  ← tab bar + filter toggle
├──────────────────────────────────────────────────────┤
│ [Timeline] [Week] [Month]                  [Today]    │  ← ViewControls row
│         <  Mon, Jul 20, 2026  >                       │  ← date nav (hidden in Timeline)
├──────────────────────────────────────────────────────┤
│ ┌──────────────┬────────────────────────────────────┐ │
│ │ Sidebar      │  │ 08:00 ────────┐                 │ │
│ │ ──────────── │  │     Oil Change │ Toyota Camry  │ │
│ │ ✓ Technicians│  │     Minh @ Bay1│               │ │
│ │ ✓ Minh       │  ├───────────────┘                 │ │
│ │ ✓ Hai        │  │ 09:00 ───────────────┐          │ │
│ │ ✓ Nam        │  │     Brake Repl.     │ Mazda CX-5│ │
│ │              │  │     Nam @ Bay1      │           │ │
│ │ ✓ Bays       │  ├────────────────────┘           │ │
│ │ ✓ Bay 1      │                                     │
│ │ ✓ Bay 2      │  (Timeline view — shown)            │
│ └──────────────┴────────────────────────────────────┘ │
├──────────────────────────────────────────────────────┤
│  Week view: 7-column grid with horizontal scroll     │
│  Month view: date-grid with appointment dots          │
│                                                        │
│  Mobile (<768px): full-width sidebar overlay via 🗘   │
└──────────────────────────────────────────────────────┘
```

### Views

| View | Description |
|------|-------------|
| **Timeline** | Vertical card list with sticky day headers. Each card shows time range, service, vehicle, tech, bay. Horizontal layout on desktop. |
| **Week** | 7-column grid (Mon-Sun). `minmax(110px, 1fr)` columns with horizontal scroll on mobile. |
| **Month** | Calendar grid with day cells. Click a day to jump to that date in current view. |

### States

- **Empty**: "No appointments yet. Book your first one!" + CTA
- **Loading**: Skeleton cards
- **Error**: "Could not load appointments" + retry
- **Filtered empty**: "No appointments match this filter"

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
│ ✓ Available! Minh · Bay 1  │       │ [2026-07-20 09:00 ] (60m)│
│                             │       │                         │
│ Notes (optional)           │       │ ✓ Available! Minh·Bay1  │
│ ┌───────────────────────┐  │       │                         │
│ │ Please check tires... │  │       │ Notes (optional)         │
│ └───────────────────────┘  │       │ ┌─────────────────────┐ │
│                             │       │ │ Please check tires..│ │
│ [Cancel]  [Book Appt]     │       │ └─────────────────────┘ │
└─────────────────────────────┘       │                         │
                                      │ [Cancel] [Book Appt]    │
                                      └─────────────────────────┘
```

### Booking Flow

1. Select Vehicle + Service (2-column grid)
2. Pick Date & Time (single `datetime-local` input)
3. **Auto-check**: 500ms debounce → API call → green/red alert appears inline
4. Optional: add notes
5. Click "Book Appointment" (disabled until available)
6. On success: modal closes, list refreshes

### Key UX Decisions

| Decision | Rationale |
|----------|-----------|
| Flat form (no steps) | Faster, fewer taps on mobile |
| `datetime-local` input | Single field, native picker on mobile |
| Auto-check availability | No extra button needed, instant feedback |
| Notes below alert | Alert is primary feedback; notes are secondary |
| Inline CTA buttons | Cancel + Book side by side with proportional width (1:2) |
| Slide-up bottom sheet on mobile | Familiar pattern on iOS/Android |

---

## Mobile Responsive Strategy

| Element | Desktop (>768px) | Mobile |
|---------|:----------------:|:------:|
| Layout | Sidebar + content | Full-width stacked |
| Sidebar | Always visible, 220px | Hidden, toggle via filter icon 🗘 |
| Form rows | 2-column grid | Single column full-width |
| Booking modal | Centered dialog | Slide-up bottom sheet |
| Timeline | Horizontal cards | Vertical card list |
| Week/Month | 7-column grid | `minmax` forced horizontal scroll |
| FAB filter | Hidden | 🔄 (replaced by tab-bar icon) |

---

## Component Tree

```
App
├── Header (sticky)
│   ├── Logo (K icon + "Keyloop Scheduler")
│   └── New Booking button
├── TabBar (sticky)
│   ├── Confirmed tab
│   ├── Cancelled tab
│   └── Filter icon (mobile only)
├── ViewControls (sticky)
│   ├── Timeline / Week / Month buttons
│   ├── Date navigation (< date >)
│   └── Today button
├── Main Content
│   ├── Sidebar (desktop) / Overlay (mobile)
│   │   ├── Technicians list (checkboxes)
│   │   └── Bays list (checkboxes)
│   └── Active View
│       ├── TimelineView (vertical card list)
│       ├── WeekView (7-column grid)
│       └── MonthView (calendar grid)
└── BookingModal
    ├── Vehicle + Service selects
    ├── Datetime-local input
    ├── Auto-check alert
    ├── Notes textarea
    └── Cancel + Book buttons
```

---

## API Integration

- **`fetchAppointments(params?)`** → populates current view
- **`checkAvailability(data)`** → auto-check on form field change (500ms debounce)
- **`createAppointment(data)`** → books and closes modal
- **`cancelAppointment(id)`** → updates appointment status
- **`fetchVehicles()`** + **`fetchServiceTypes()`** → populate dropdowns

All API calls go through `apiFetch<T>()` wrapper with error handling.

---

## Styling

- **CSS Modules**: Scoped styles per component (`*.module.css`)
- **Light theme**: White + gray-50 background, blue-600 primary
- **Font sizes**: 16px base (prevents iOS zoom on focus)
- **Grid layout**: 2-column grid for form rows, single column on mobile
- **Consistent heights**: All inputs 44px with `box-sizing: border-box`
- **Duration hint**: Shows service duration below datetime input

---

## Known Issues

- **Safari datetime input**: Native picker icons cause slight visual mismatch vs `<select>` elements. Height is identical (44px) but internal spacing differs.
- **No pagination**: All appointments loaded at once. Fine for demo (<100 records).
