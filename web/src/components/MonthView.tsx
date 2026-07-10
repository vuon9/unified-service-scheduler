import type { Appointment, Technician } from '../types';

interface Props {
  appointments: Appointment[];
  technicians: Technician[];
  selectedTech: string;
  currentDate: Date;
  onViewDetail: (apt: Appointment) => void;
  onDateChange: (d: Date) => void;
}

const DAY_NAMES = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
const COLORS = ['#2563EB', '#16A34A', '#9333EA', '#D97706', '#DC2626', '#0891B2'];

export default function MonthView({ appointments, technicians, selectedTech, currentDate, onViewDetail, onDateChange }: Props) {
  const year = currentDate.getFullYear();
  const month = currentDate.getMonth();

  const firstDay = new Date(year, month, 1);
  const lastDay = new Date(year, month + 1, 0);
  const startPad = firstDay.getDay(); // 0=Sun
  const totalDays = lastDay.getDate();

  // All cells: pad from prev month + current month
  const cells: (Date | null)[] = [];
  for (let i = 0; i < startPad; i++) cells.push(null);
  for (let d = 1; d <= totalDays; d++) cells.push(new Date(year, month, d));

  const monthStart = new Date(year, month, 1);
  const monthEnd = new Date(year, month + 1, 0, 23, 59, 59, 999);

  const monthAppts = appointments.filter(a => {
    if (a.status !== 'confirmed') return false;
    const s = new Date(a.scheduled_start);
    return s >= monthStart && s <= monthEnd;
  });

  // Build a map of date -> appointments
  const byDay: Record<string, Appointment[]> = {};
  monthAppts.forEach(a => {
    const key = new Date(a.scheduled_start).toISOString().split('T')[0];
    if (!byDay[key]) byDay[key] = [];
    byDay[key].push(a);
  });

  const techNameById: Record<string, string> = {};
  technicians.forEach(t => { techNameById[t.id] = t.name; });

  const now = new Date();
  const todayStr = now.toISOString().split('T')[0];

  return (
    <div style={{ background: '#fff', borderRadius: '8px', border: '1px solid #E5E7EB' }}>
      {/* Day name headers */}
      <div style={{ display: 'inline-grid', gridTemplateColumns: 'repeat(7, minmax(100px, 1fr))', background: '#F9FAFB', borderBottom: '2px solid #E5E7EB', minWidth: '100%' }}>
        {DAY_NAMES.map((d, i) => (
          <div key={d} style={{ padding: '10px 4px', textAlign: 'center', fontSize: '12px', fontWeight: 600, color: '#6B7280', borderRight: i === DAY_NAMES.length - 1 ? '1px solid #E5E7EB' : 'none' }}>
            {d}
          </div>
        ))}
      </div>

      {/* Calendar grid */}
      <div style={{ display: 'inline-grid', gridTemplateColumns: 'repeat(7, minmax(100px, 1fr))', minWidth: '100%' }}>
        {cells.map((date, i) => {
          if (!date) {
            return <div key={`pad-${i}`} style={{ borderRight: '1px solid #E5E7EB', borderBottom: '1px solid #E5E7EB', background: '#F9FAFB' }} />;
          }

          const dateStr = date.toISOString().split('T')[0];
          const isToday = dateStr === todayStr;
          const apts = byDay[dateStr] || [];
          const filteredApts = selectedTech
            ? apts.filter(a => a.technician_id === selectedTech)
            : apts;

          return (
            <div
              key={dateStr}
              onClick={() => onDateChange(date)}
              style={{
                borderRight: '1px solid #E5E7EB', borderBottom: '1px solid #E5E7EB',
                padding: '6px', minHeight: '100px', cursor: 'pointer',
                background: isToday ? '#EFF6FF' : '#fff',
                transition: 'background 0.1s',
              }}
            >
              <div style={{
                fontSize: '13px', fontWeight: 600,
                color: isToday ? '#2563EB' : '#6B7280',
                marginBottom: '4px',
              }}>
                {date.getDate()}
              </div>
              {filteredApts.slice(0, 3).map(a => {
                const ti = technicians.findIndex(t => t.id === a.technician_id);
                const color = COLORS[Math.max(0, ti) % COLORS.length];
                return (
                  <div
                    key={a.id}
                    onClick={(e) => { e.stopPropagation(); onViewDetail(a); }}
                    style={{
                      background: color, borderRadius: '3px', padding: '2px 4px', marginBottom: '2px',
                      cursor: 'pointer',
                    }}
                  >
                    <div style={{ fontSize: '10px', fontWeight: 600, color: '#fff', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                      {new Date(a.scheduled_start).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}
                    </div>
                    <div style={{ fontSize: '9px', color: 'rgba(255,255,255,0.9)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                      {a.service_type_name}
                    </div>
                  </div>
                );
              })}
              {filteredApts.length > 3 && (
                <div style={{ fontSize: '10px', color: '#9CA3AF', fontWeight: 500 }}>
                  +{filteredApts.length - 3} more
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
