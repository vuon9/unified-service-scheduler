import { useMemo } from 'react';
import type { Appointment, Technician, ServiceBay } from '../types';

interface Props {
  technicians: Technician[];
  serviceBays: ServiceBay[];
  appointments: Appointment[];
  selectedTech: string;
  selectedBay: string;
  onSelectTech: (id: string) => void;
  onSelectBay: (id: string) => void;
  fullWidth?: boolean;
}

const COLORS = ['#2563EB', '#16A34A', '#9333EA', '#D97706', '#DC2626', '#0891B2'];

const sectionLabelStyle: React.CSSProperties = {
  fontSize: '11px', fontWeight: 700, textTransform: 'uppercase',
  color: '#9CA3AF', letterSpacing: '0.05em', padding: '8px 10px 4px',
};

const itemStyle = (active: boolean, color: string): React.CSSProperties => ({
  display: 'flex', alignItems: 'center', justifyContent: 'space-between',
  padding: '8px 10px', margin: '2px 0', borderRadius: '6px',
  cursor: 'pointer',
  background: active ? `${color}15` : 'transparent',
  border: active ? `1px solid ${color}40` : '1px solid transparent',
  transition: 'all 0.15s',
});

export default function Sidebar({
  technicians, serviceBays, appointments,
  selectedTech, selectedBay,
  onSelectTech, onSelectBay, fullWidth,
}: Props) {
  const techCounts = useMemo(() => {
    const m: Record<string, number> = {};
    technicians.forEach(t => { m[t.id] = 0; });
    appointments.forEach(a => {
      if (a.status === 'confirmed' && m[a.technician_id] !== undefined) m[a.technician_id]++;
    });
    return m;
  }, [technicians, appointments]);

  const bayCounts = useMemo(() => {
    const m: Record<string, number> = {};
    serviceBays.forEach(b => { m[b.id] = 0; });
    appointments.forEach(a => {
      if (a.status === 'confirmed' && m[a.service_bay_id] !== undefined) m[a.service_bay_id]++;
    });
    return m;
  }, [serviceBays, appointments]);

  const hasFilter = selectedTech || selectedBay;

  // Count filtered results (AND logic)
  const filteredCount = useMemo(() => {
    return appointments.filter(a => {
      if (a.status !== 'confirmed') return false;
      if (selectedTech && a.technician_id !== selectedTech) return false;
      if (selectedBay && a.service_bay_id !== selectedBay) return false;
      return true;
    }).length;
  }, [appointments, selectedTech, selectedBay]);

  return (
    <div style={{
      width: fullWidth ? '100%' : '220px',
      flexShrink: 0, background: '#fff', borderRadius: '8px',
      border: '1px solid #E5E7EB', overflow: 'hidden',
    }}>
      {/* Filter count header */}
      {hasFilter && (
        <div style={{
          padding: '6px 10px', fontSize: '11px', color: '#6B7280',
          background: '#F9FAFB', borderBottom: '1px solid #E5E7EB',
        }}>
          {filteredCount} appointment{filteredCount !== 1 ? 's' : ''}
        </div>
      )}

      <div style={{ overflowY: 'auto', padding: '6px 8px' }}>
        {/* Technicians section */}
        <div style={sectionLabelStyle}>Technicians</div>
        {technicians.length === 0 ? (
          <div style={{ padding: '8px 10px', color: '#9CA3AF', fontSize: '12px' }}>None</div>
        ) : (
          technicians.map((t, i) => {
            const active = selectedTech === t.id;
            const color = COLORS[i % COLORS.length];
            const count = techCounts[t.id] || 0;
            return (
              <div key={t.id} style={itemStyle(active, color)} onClick={() => onSelectTech(t.id)}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <div style={{
                    width: '8px', height: '8px', borderRadius: '50%',
                    background: active ? color : '#D1D5DB', flexShrink: 0,
                  }} />
                  <span style={{ fontSize: '13px', fontWeight: active ? 600 : 400, color: '#111827' }}>
                    {t.name}
                  </span>
                </div>
                <span style={{
                  fontSize: '11px', fontWeight: 600,
                  color: active ? color : '#9CA3AF',
                  background: active ? `${color}15` : '#F3F4F6',
                  padding: '1px 6px', borderRadius: '8px',
                }}>{count}</span>
              </div>
            );
          })
        )}

        {/* Divider */}
        <div style={{ height: '1px', background: '#E5E7EB', margin: '8px 4px' }} />

        {/* Bays section */}
        <div style={sectionLabelStyle}>Service Bays</div>
        {serviceBays.length === 0 ? (
          <div style={{ padding: '8px 10px', color: '#9CA3AF', fontSize: '12px' }}>None</div>
        ) : (
          serviceBays.map((b) => {
            const active = selectedBay === b.id;
            const count = bayCounts[b.id] || 0;
            return (
              <div key={b.id} style={itemStyle(active, '#0891B2')} onClick={() => onSelectBay(b.id)}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <div style={{
                    width: '8px', height: '8px', borderRadius: '2px',
                    background: active ? '#0891B2' : '#D1D5DB', flexShrink: 0,
                  }} />
                  <span style={{ fontSize: '13px', fontWeight: active ? 600 : 400, color: '#111827' }}>
                    {b.name}
                  </span>
                </div>
                <span style={{
                  fontSize: '11px', fontWeight: 600,
                  color: active ? '#0891B2' : '#9CA3AF',
                  background: active ? '#0891B215' : '#F3F4F6',
                  padding: '1px 6px', borderRadius: '8px',
                }}>{count}</span>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
