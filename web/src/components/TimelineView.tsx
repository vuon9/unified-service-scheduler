import type { Appointment, Technician } from '../types';
import { useMemo } from 'react';

interface Props {
  appointments: Appointment[];
  technicians: Technician[];
  selectedTech: string;
  currentDate: Date;
  onViewDetail: (apt: Appointment) => void;
}

const HOUR_HEIGHT = 60;
const HOURS = Array.from({ length: 10 }, (_, i) => i + 8); // 8AM to 5PM

function fmtTime(d: Date) {
  return d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
}

const COLORS = ['#2563EB', '#16A34A', '#9333EA', '#D97706', '#DC2626', '#0891B2'];

export default function TimelineView({ appointments, technicians, selectedTech, currentDate, onViewDetail }: Props) {
  const dayStart = useMemo(() => {
    const d = new Date(currentDate);
    d.setHours(0, 0, 0, 0);
    return d;
  }, [currentDate]);

  const dayEnd = useMemo(() => {
    const d = new Date(dayStart);
    d.setDate(d.getDate() + 1);
    return d;
  }, [dayStart]);

  const filteredTechs = useMemo(() => {
    let list = technicians;
    if (selectedTech) list = list.filter(t => t.id === selectedTech);
    return list;
  }, [technicians, selectedTech]);

  const dayAppts = useMemo(() => {
    return appointments.filter(a => {
      if (a.status !== 'confirmed') return false;
      const s = new Date(a.scheduled_start);
      return s >= dayStart && s < dayEnd;
    });
  }, [appointments, dayStart, dayEnd]);

  const dayStr = dayStart.toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' });
  const noAppts = dayAppts.length === 0;

  // Compute column width
  const colW = 120;

  return (
    <div style={{ background: '#fff', borderRadius: '8px', border: '1px solid #E5E7EB', overflow: 'auto' }}>
      <div style={{ padding: '12px 16px', borderBottom: '1px solid #E5E7EB', fontWeight: 600, fontSize: '15px', color: '#111827' }}>
        {dayStr}
        {noAppts && <span style={{ color: '#9CA3AF', fontWeight: 400, marginLeft: '8px', fontSize: '13px' }}>— no appointments</span>}
      </div>

      {filteredTechs.length === 0 ? (
        <div style={{ padding: '40px', textAlign: 'center', color: '#9CA3AF' }}>No technicians available.</div>
      ) : (
        <div style={{ minWidth: filteredTechs.length > 0 ? `${HOURS.length * colW + 140}px` : '100%' }}>
          {/* Header row: time labels */}
          <div style={{ display: 'flex', borderBottom: '2px solid #E5E7EB', position: 'sticky', top: 0, background: '#F9FAFB' }}>
            <div style={{ width: '130px', flexShrink: 0, padding: '8px 12px', fontSize: '12px', fontWeight: 600, color: '#6B7280' }}>Technician</div>
            {HOURS.map(h => (
              <div key={h} style={{ width: `${colW}px`, flexShrink: 0, padding: '8px 4px', fontSize: '12px', fontWeight: 600, color: '#6B7280', textAlign: 'center', borderLeft: '1px solid #E5E7EB' }}>
                {h % 12 || 12}:00 {h < 12 ? 'AM' : 'PM'}
              </div>
            ))}
          </div>

          {/* Tech rows */}
          {filteredTechs.map((tech, ti) => {
            const techAppts = dayAppts.filter(a => a.technician_id === tech.id);
            const color = COLORS[ti % COLORS.length];
            return (
              <div key={tech.id} style={{ display: 'flex', borderBottom: '1px solid #E5E7EB', minHeight: `${HOUR_HEIGHT}px` }}>
                <div style={{
                  width: '130px', flexShrink: 0, padding: '8px 12px', fontSize: '13px', fontWeight: 600,
                  color: '#111827', display: 'flex', alignItems: 'center',
                }}>
                  {tech.name}
                </div>
                {HOURS.map((h, hi) => {
                  const slotStart = new Date(dayStart);
                  slotStart.setHours(h, 0, 0, 0);
                  const slotEnd = new Date(slotStart);
                  slotEnd.setHours(h + 1, 0, 0, 0);

                  const cellAppts = techAppts.filter(a => {
                    const as = new Date(a.scheduled_start);
                    const ae = new Date(a.scheduled_end);
                    return as < slotEnd && ae > slotStart;
                  });

                  return (
                    <div key={hi} style={{
                      width: `${colW}px`, flexShrink: 0, borderLeft: '1px solid #E5E7EB',
                      position: 'relative', minHeight: `${HOUR_HEIGHT}px`,
                      background: hi % 2 === 0 ? '#F9FAFB' : '#fff',
                    }}>
                      {cellAppts.map(a => {
                        const as = new Date(a.scheduled_start);
                        const ae = new Date(a.scheduled_end);
                        const durMin = (ae.getTime() - as.getTime()) / 60000;
                        const offsetMin = Math.max(0, (as.getTime() - slotStart.getTime()) / 60000);
                        const blockH = Math.max(28, (durMin / 60) * HOUR_HEIGHT);
                        const blockT = Math.max(2, (offsetMin / 60) * HOUR_HEIGHT);
                        return (
                          <div
                            key={a.id}
                            onClick={() => onViewDetail(a)}
                            style={{
                              position: 'absolute', top: `${blockT}px`,
                              left: '2px', right: '2px', height: `${blockH}px`,
                              background: color, borderRadius: '6px', padding: '2px 6px',
                              cursor: 'pointer', overflow: 'hidden', zIndex: 2,
                              display: 'flex', flexDirection: 'column',
                              justifyContent: 'center',
                              boxShadow: '0 1px 2px rgba(0,0,0,0.1)',
                            }}
                          >
                            <div style={{ fontWeight: 600, fontSize: '11px', color: '#fff', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                              {a.service_type_name}
                            </div>
                            {durMin >= 90 && (
                              <div style={{ fontSize: '10px', color: 'rgba(255,255,255,0.85)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                                {a.vehicle_make} {a.vehicle_model}
                              </div>
                            )}
                          </div>
                        );
                      })}
                    </div>
                  );
                })}
              </div>
            );
          })}
        </div>
      )}

      {/* Legend */}
      {!noAppts && (
        <div style={{ padding: '8px 16px', borderTop: '1px solid #E5E7EB', display: 'flex', gap: '16px', flexWrap: 'wrap', fontSize: '12px', color: '#6B7280' }}>
          {filteredTechs.map((t, i) => (
            <span key={t.id} style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
              <span style={{ width: '10px', height: '10px', borderRadius: '2px', background: COLORS[i % COLORS.length], display: 'inline-block' }} />
              {t.name}
            </span>
          ))}
        </div>
      )}
    </div>
  );
}
