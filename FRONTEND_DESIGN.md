# FRONTEND DESIGN -- Unified Service Scheduler

> React + Vite demo UI. Minimal, clean, focused on the booking flow.
> Date: 2026-07-09

---

## Design Goals

- **Demo-first**: every screen is designed to look good in a 5-10 min video walkthrough
- **API-driven**: all data comes from the Go backend (`localhost:8080/api/v1`)
- **No router**: single-page with tab/step navigation (simpler, faster to build)
- **No state management lib**: React hooks + fetch is enough for this scope
- **Clean aesthetic**: light theme, card-based UI, clear CTAs

---

## Screen 1: Appointments Dashboard (Home)

```
+--------------------------------------------------------------+
|  Keyloop Scheduler                               [+ New Booking] |
+--------------------------------------------------------------+
|  [All] [Confirmed] [Cancelled]                   [Filter: v]  |
+--------------------------------------------------------------+
|                                                                |
|  +----------------------------------------------------------+ |
|  | 2026-07-15 | 09:00 - 10:00 | Oil Change           | ... | |
|  | Toyota Camry | Bay 1 | Minh (Tech) | CONFIRMED     |     | |
|  +----------------------------------------------------------+ |
|  +----------------------------------------------------------+ |
|  | 2026-07-15 | 11:00 - 13:00 | Brake Replacement    | ... | |
|  | Toyota Camry | Bay 2 | Nam (Tech) | CONFIRMED     |     | |
|  +----------------------------------------------------------+ |
|  +----------------------------------------------------------+ |
|  | 2026-07-14 | 14:00 - 15:00 | Engine Diagnostic    | ... | |
|  | Honda Civic  | Bay 1 | Hai (Tech) | CANCELLED     |     | |
|  +----------------------------------------------------------+ |
|                                                                |
+--------------------------------------------------------------+
```

**States:**
- **Empty**: "No appointments yet. Book your first one!" + CTA button
- **Loading**: shimmer/skeleton cards
- **Error**: "Could not load appointments" + retry button
- **Filtered empty**: "No appointments match this filter"

---

## Screen 2: Book New Appointment (Modal or Page)

```
+--------------------------------------------------------------+
|  Book New Appointment                                    [X]  |
+--------------------------------------------------------------+
|                                                                |
|  Step 1: Select Vehicle & Service                              |
|  +----------------------------------------------------------+ |
|  | Vehicle:         [v Toyota Camry 2023     | v]            | |
|  | Service:         [v Oil Change (60 min)   | v]            | |
|  +----------------------------------------------------------+ |
|                                                                |
|  Step 2: Pick Date & Time                                      |
|  +----------------------------------------------------------+ |
|  | Date:        [2026-07-15]                                 | |
|  | Time:        [09:00]                                      | |
|  +----------------------------------------------------------+ |
|                                                                |
|  Step 3: Check Availability                                    |
|  +----------------------------------------------------------+ |
|  |                                       [Check Availability] | |
|  |                                                            | |
|  |  Available: Yes!                                           | |
|  |  Technician: Minh (any qualified tech)                     | |
|  |  Bay:        Bay 1 or Bay 2                               | |
|  +----------------------------------------------------------+ |
|                                                                |
|                        [Cancel]   [  Book Appointment  ]       |
+--------------------------------------------------------------+
```

**Flow:**
1. User selects vehicle + service type -> see computed duration
2. User picks date + time
3. User clicks "Check Availability" -> calls `POST /availability`
4. Shows result inline (green check or red X)
5. If available -> "Book Appointment" button enabled
6. On success -> close modal, refresh list, show toast "Booked!"

**States:**
- **Availability loading**: spinner on the button
- **Available**: green banner with tech + bay info
- **Unavailable (no bay)**: red banner "No service bays available at this time. Try another time."
- **Unavailable (no tech)**: red banner "No qualified technician available. Try another time."
- **Unavailable (both)**: red banner "No resources available."
- **Server error**: red banner with error message

---

## Screen 3: Appointment Detail (Expanded Card or Modal)

```
+--------------------------------------------------------------+
|  Appointment Detail                                      [X]  |
+--------------------------------------------------------------+
|                                                                |
|  +----------------------------------------------------------+ |
|  | STATUS: CONFIRMED                                         | |
|  |                                                            | |
|  | Vehicle:      Toyota Camry 2023 (VIN: ...)                | |
|  | Service:      Oil Change (60 min)                         | |
|  | Date:         2026-07-15                                  | |
|  | Time:         09:00 - 10:00                               | |
|  | Dealer:       Saigon Auto                                 | |
|  | Technician:   Minh                                        | |
|  | Bay:          Bay 1                                       | |
|  | Booked at:    2026-07-09 12:00                            | |
|  +----------------------------------------------------------+ |
|                                                                |
|                   [  Cancel Appointment  ]                     |
+--------------------------------------------------------------+
```

**Cancel flow:**
1. Click "Cancel Appointment"
2. Confirmation dialog: "Are you sure you want to cancel this appointment?"
3. On confirm -> `POST /appointments/{id}/cancel`
4. On success -> close detail, refresh list, toast "Cancelled"

---

## Screen 4: Toast / Feedback System

```
+----------------------------+
|  [check] Appointment booked! |
+----------------------------+

+----------------------------+
|  [X] No technician available |
|     Try another time slot.   |
+----------------------------+

+----------------------------+
|  [check] Appointment cancelled |
+----------------------------+
```

Position: top-right, auto-dismiss after 4 seconds.

---

## Component Tree

```
App
|-- Header ("Keyloop Scheduler" + [New Booking] button)
|-- ToastContainer (top-right, stacked)
|-- TabBar: [All] [Confirmed] [Cancelled]
|-- AppointmentList
|   |-- AppointmentCard (×N)
|       |-- StatusBadge (Confirmed/Cancelled)
|       |-- TimeInfo
|       |-- ResourceInfo (tech + bay)
|       |-- "View Details" link
|-- BookingModal
|   |-- VehicleSelect (dropdown, fetch from API)
|   |-- ServiceSelect (dropdown, fetch from API)
|   |-- DateTimePicker (date input + time input)
|   |-- AvailabilityChecker
|   |   |-- CheckButton
|   |   |-- AvailabilityResult
|-- DetailModal
|   |-- AppointmentInfo (all fields)
|   |-- CancelButton (with confirmation)
```

---

## API Calls

| Component | Endpoint | Method |
|-----------|----------|--------|
| App (on mount) | `/appointments` | GET |
| VehicleSelect | `/vehicles` | GET (or seeded) |
| ServiceSelect | `/service-types` | GET (or seeded) |
| AvailabilityChecker | `/availability` | POST |
| BookingModal (submit) | `/appointments` | POST |
| DetailModal (cancel) | `/appointments/{id}/cancel` | POST |

---

## Tech Choices

| Choice | Why |
|--------|-----|
| **Vite** | Fastest React dev server. HMR out of the box. |
| **No React Router** | Single-page app. Tabs + modals = simpler code, less to explain in video. |
| **No UI library** | Plain CSS modules or Tailwind. Demonstrates we can build UI, not just wire up components. |
| **fetch (native)** | No axios needed. Clean, simple, standard. |
| **No state library** | useState + useEffect + custom hooks (useAppointments, useBooking) = enough. |

---

## Color Palette (Clean / Professional)

| Role | Color | Hex |
|------|-------|-----|
| Primary | Blue | `#2563EB` |
| Success | Green | `#16A34A` |
| Danger | Red | `#DC2626` |
| Warning | Amber | `#D97706` |
| Background | White/Gray | `#F9FAFB` |
| Card | White | `#FFFFFF` |
| Text | Dark Gray | `#111827` |
| Text Secondary | Gray | `#6B7280` |

---

## Video Walkthrough Flow

1. **Dashboard**: show list of appointments (pre-seeded with a few)
2. **Book new**: walk through the form, check availability -> book
3. **Show conflict**: try to book same time again -> see "no availability" error
4. **Cancel**: go to detail, cancel an appointment
5. **Verify**: list refreshes, cancelled shows with strikethrough/badge
6. **AI narrative**: mention how AI generated the components from these wireframes

Total video time: ~7 minutes for the demo portion.

---

## What We DON'T Build (MVP)

- Responsive mobile layout (desktop-first for video recording)
- Authentication/login
- Vehicle/service type CRUD (just consume from backend seed data)
- Dark mode
- Loading skeletons (simple spinner is fine)
- Pagination (one page of appointments is enough)
