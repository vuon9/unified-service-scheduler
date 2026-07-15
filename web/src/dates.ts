// Returns YYYY-MM-DD in the user's local timezone.
// Use this instead of date.toISOString().split('T')[0] to avoid
// timezone-shift bugs when grouping appointments by calendar day.
export function toDateKey(d: Date | string): string {
  const date = typeof d === 'string' ? new Date(d) : d;
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}
