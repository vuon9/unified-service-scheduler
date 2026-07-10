import type { Appointment } from '../types';
import { useMemo } from 'react';

interface Props {
  appointments: Appointment[];
  technicians?: never; // kept for API compat
  selectedTech: string;
  currentDate: Date;
  onViewDetail: (apt: Appointment) => void;
}

const COLORS = ['#2563EB', '#16A34A', '#9333EA', '#D97706', '#DC2626', '#0891B2'];

function fmtTime(d: Date) {
  return d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: true });
}

function isSameDay(a: Date, b: Date) {
  return a.getFullYear() === b.getFullYear()
    && a.getMonth() === b.getMonth()
    && a.getDate() === b.getDate();
}

function formatDayHeader(d: Date) {
  const today = new Date();
  const tomorrow = new Date(today);
  tomorrow.setDate(tomorrow.getDate() + 1);
  const yesterday = new Date(today);
  yesterday.setDate(yesterday.getDate() - 1);

  if (isSameDay(d, today)) return 'Today';
  if (isSameDay(d, tomorrow)) return 'Tomorrow';
  if (isSameDay(d, yesterday)) return 'Yesterday';

  return d.toLocaleDateString('en-US', { weekday: 'long', month: 'short', day: 'numeric', year: 'numeric' });
}

function formatCompactDay(d: Date) {
  return d.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' });
}

export default function TimelineView({ appointments, selectedTech, currentDate, onViewDetail }: Props) {
  // Show ALL confirmed appointments, no date range filter
  const filtered = useMemo(() => {
    let list = appointments.filter(a => {
      if (a.status !== 'confirmed') return false;
      return true; // no date range — show everything
    });
    if (selectedTech) {
      list = list.filter(a => a.technician_id === selectedTech);
    }
    // Sort by start time
    list.sort((a, b) => new Date(a.scheduled_start).getTime() - new Date(b.scheduled_start).getTime());
    return list;
  }, [appointments, selectedTech]);

  // Group by date
  const days = useMemo(() => {
    const map = new Map<string, Appointment[]>();
    for (const a of filtered) {
      const key = new Date(a.scheduled_start).toDateString();
      if (!map.has(key)) map.set(key, []);
      map.get(key)!.push(a);
    }
    return Array.from(map.entries());
  }, [filtered]);

  if (filtered.length === 0) {
    return (
      <div style={{ background: '#fff', borderRadius: '8px', border: '1px solid #E5E7EB', padding: '40px 16px', textAlign: 'center' }}>
        <div style={{ fontSize: '15px', fontWeight: 600, color: '#6B7280' }}>No confirmed appointments</div>
        <div style={{ fontSize: '13px', color: '#9CA3AF', marginTop: '4px' }}>Book a new appointment to get started</div>
      </div>
    );
  }

  return (
    <div style={{ background: '#fff', borderRadius: '8px', border: '1px solid #E5E7EB', overflow: 'hidden' }}>
      {days.map(([dateKey, appts]) => {
        const firstDate = new Date(appts[0].scheduled_start);
        return (
          <div key={dateKey}>
            {/* Sticky date header — sticks to top when scrolling */}
            <div style={{
              position: 'sticky', top: 0, zIndex: 10,
              background: '#F9FAFB', borderBottom: '1px solid #E5E7EB',
              padding: '10px 16px',
              display: 'flex', alignItems: 'center', justifyContent: 'space-between',
            }}>
              <span style={{ fontWeight: 700, fontSize: '15px', color: '#111827' }}>
                {formatDayHeader(firstDate)}
              </span>
              <span style={{ fontSize: '12px', color: '#9CA3AF', fontWeight: 500 }}>
                {appts.length} appointment{appts.length !== 1 ? 's' : ''}
              </span>
            </div>

            {/* Appointment cards */}
            <div style={{ padding: '4px 12px' }}>
              {appts.map((a, i) => {
                const start = new Date(a.scheduled_start);
                const end = new Date(a.scheduled_end);
                const color = COLORS[i % COLORS.length];
                return (
                  <div
                    key={a.id}
                    onClick={() => onViewDetail(a)}
                    style={{
                      display: 'flex', gap: '12px', padding: '10px 0',
                      borderBottom: i < appts.length - 1 ? '1px solid #F3F4F6' : 'none',
                      cursor: 'pointer', transition: 'background 0.15s',
                      borderRadius: '6px',
                    }}
                    onMouseEnter={e => (e.currentTarget.style.background = '#F9FAFB')}
                    onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
                  >
                    {/* Time column */}
                    <div style={{
                      width: '70px', flexShrink: 0, textAlign: 'right',
                      paddingTop: '2px',
                    }}>
                      <div style={{ fontSize: '14px', fontWeight: 600, color: '#111827', lineHeight: 1.3 }}>
                        {fmtTime(start)}
                      </div>
                      <div style={{ fontSize: '11px', fontWeight: 500, color: '#9CA3AF', lineHeight: 1.3 }}>
                        {fmtTime(end)}
                      </div>
                    </div>

                    {/* Color indicator */}
                    <div style={{
                      width: '4px', flexShrink: 0, borderRadius: '2px',
                      background: color, alignSelf: 'stretch',
                    }} />

                    {/* Details */}
                    <div style={{ flex: 1, minWidth: 0, paddingTop: '1px' }}>
                      <div style={{ fontSize: '14px', fontWeight: 600, color: '#111827', marginBottom: '2px' }}>
                        {a.service_type_name}
                      </div>
                      <div style={{ fontSize: '12px', color: '#6B7280', lineHeight: 1.5 }}>
                        {a.vehicle_make} {a.vehicle_model}
                      </div>
                      <div style={{ fontSize: '12px', color: '#6B7280', lineHeight: 1.5 }}>
                        <span style={{ display: 'inline-flex', alignItems: 'center', gap: '4px' }}>
                          🔧 {a.technician_name || a.technician_id}
                        </span>
                        <span style={{ margin: '0 6px', color: '#D1D5DB' }}>·</span>
                        <span style={{ display: 'inline-flex', alignItems: 'center', gap: '4px' }}>
                          🅿️ {a.service_bay_name || a.service_bay_id}
                        </span>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        );
      })}
    </div>
  );
}
