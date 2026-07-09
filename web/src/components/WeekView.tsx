import type { Appointment, Technician } from '../types';

interface Props {
  appointments: Appointment[];
  technicians: Technician[];
  selectedTech: string;
  currentDate: Date;
  onViewDetail: (apt: Appointment) => void;
}

const COLORS = ['#2563EB', '#16A34A', '#9333EA', '#D97706', '#DC2626', '#0891B2'];

export default function WeekView({ appointments, technicians, selectedTech, currentDate, onViewDetail }: Props) {
  // Compute Mon-Sun of the week containing currentDate
  const weekStart = new Date(currentDate);
  const day = weekStart.getDay();
  const diff = weekStart.getDate() - day + (day === 0 ? -6 : 1); // Start on Monday
  weekStart.setDate(diff);
  weekStart.setHours(0, 0, 0, 0);

  const days = Array.from({ length: 7 }, (_, i) => {
    const d = new Date(weekStart);
    d.setDate(d.getDate() + i);
    return d;
  });

  const weekEnd = new Date(days[6]);
  weekEnd.setHours(23, 59, 59, 999);

  const filteredTechs = selectedTech
    ? technicians.filter(t => t.id === selectedTech)
    : technicians;

  const weekAppts = appointments.filter(a => {
    if (a.status !== 'confirmed') return false;
    const s = new Date(a.scheduled_start);
    return s >= weekStart && s <= weekEnd;
  });

  const byDay: Record<string, Appointment[]> = {};
  days.forEach(d => { byDay[d.toISOString().split('T')[0]] = []; });
  weekAppts.forEach(a => {
    const key = new Date(a.scheduled_start).toISOString().split('T')[0];
    if (byDay[key]) byDay[key].push(a);
  });

  const now = new Date();
  const todayStr = now.toISOString().split('T')[0];

  return (
    <div style={{ background: '#fff', borderRadius: '8px', border: '1px solid #E5E7EB', overflow: 'hidden' }}>
      {/* Day headers */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', borderBottom: '2px solid #E5E7EB', background: '#F9FAFB' }}>
        {days.map((d, i) => {
          const dateStr = d.toISOString().split('T')[0];
          const isToday = dateStr === todayStr;
          return (
            <div key={i} style={{
              padding: '10px 8px', textAlign: 'center', borderLeft: i > 0 ? '1px solid #E5E7EB' : 'none',
            }}>
              <div style={{ fontSize: '12px', fontWeight: 600, color: isToday ? '#2563EB' : '#6B7280' }}>
                {d.toLocaleDateString('en-US', { weekday: 'short' })}
              </div>
              <div style={{
                fontSize: '20px', fontWeight: 700, color: isToday ? '#2563EB' : '#111827',
                width: '36px', height: '36px', display: 'flex', alignItems: 'center', justifyContent: 'center',
                margin: '2px auto', borderRadius: '50%', background: isToday ? '#EFF6FF' : 'transparent',
              }}>
                {d.getDate()}
              </div>
              <div style={{ fontSize: '11px', color: '#9CA3AF' }}>
                {d.toLocaleDateString('en-US', { month: 'short' })}
              </div>
            </div>
          );
        })}
      </div>

      {/* Appointments grid */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', minHeight: '400px' }}>
        {days.map((d, i) => {
          const dateStr = d.toISOString().split('T')[0];
          const dayApts = byDay[dateStr] || [];

          // Group by technician
          const byTech: Record<string, Appointment[]> = {};
          if (selectedTech) {
            byTech[selectedTech] = dayApts.filter(a => a.technician_id === selectedTech);
          } else {
            dayApts.forEach(a => {
              const tid = a.technician_id;
              if (!byTech[tid]) byTech[tid] = [];
              byTech[tid].push(a);
            });
          }

          return (
            <div key={i} style={{
              borderLeft: i > 0 ? '1px solid #E5E7EB' : 'none',
              borderBottom: '1px solid #E5E7EB',
              padding: '8px 6px',
              background: dateStr === todayStr ? '#F0F7FF' : '#fff',
              minHeight: '80px',
            }}>
              {Object.entries(byTech).length === 0 && (
                <div style={{ color: '#D1D5DB', fontSize: '12px', textAlign: 'center', padding: '16px 0' }}>—</div>
              )}
              {Object.entries(byTech).map(([tid, apts]) => {
                const tech = filteredTechs.find(t => t.id === tid);
                const ti = technicians.findIndex(t => t.id === tid);
                const color = COLORS[Math.max(0, ti) % COLORS.length];
                return (
                  <div key={tid} style={{ marginBottom: '6px' }}>
                    {!selectedTech && (
                      <div style={{ fontSize: '11px', fontWeight: 600, color: '#6B7280', marginBottom: '2px' }}>
                        {tech?.name || tid}
                      </div>
                    )}
                    {apts.map(a => (
                      <div key={a.id} onClick={() => onViewDetail(a)} style={{
                        background: color, borderRadius: '4px', padding: '4px 6px', marginBottom: '3px',
                        cursor: 'pointer', boxShadow: '0 1px 2px rgba(0,0,0,0.08)',
                      }}>
                        <div style={{ fontSize: '11px', fontWeight: 600, color: '#fff' }}>
                          {new Date(a.scheduled_start).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}
                        </div>
                        <div style={{ fontSize: '10px', color: 'rgba(255,255,255,0.9)' }}>
                          {a.service_type_name}
                        </div>
                        <div style={{ fontSize: '9px', color: 'rgba(255,255,255,0.7)' }}>
                          {a.vehicle_make} {a.vehicle_model}
                        </div>
                      </div>
                    ))}
                  </div>
                );
              })}
            </div>
          );
        })}
      </div>
    </div>
  );
}
