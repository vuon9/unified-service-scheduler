import type { ViewMode, Technician } from '../types';

interface ViewControlsProps {
  mode: ViewMode;
  onModeChange: (mode: ViewMode) => void;
  technicians: Technician[];
  selectedTech: string;
  onTechChange: (techId: string) => void;
  currentDate: Date;
  onDateChange: (d: Date) => void;
}

const btn = (active: boolean, color: string): React.CSSProperties => ({
  padding: '6px 14px',
  border: active ? 'none' : '1px solid #D1D5DB',
  borderRadius: '6px',
  background: active ? color : '#fff',
  color: active ? '#fff' : '#374151',
  cursor: 'pointer',
  fontSize: '13px',
  fontWeight: active ? 600 : 400,
  transition: 'all 0.15s',
});

const arrowBtn: React.CSSProperties = {
  ...btn(false, '#2563EB'),
  padding: '6px 10px',
  fontSize: '16px',
  lineHeight: '1',
};

export default function ViewControls({
  mode, onModeChange, technicians, selectedTech, onTechChange,
  currentDate, onDateChange,
}: ViewControlsProps) {
  const go = (days: number) => {
    const d = new Date(currentDate);
    d.setDate(d.getDate() + days);
    onDateChange(d);
  };

  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: '12px', flexWrap: 'wrap', padding: '8px 0' }}>
      {/* View mode buttons */}
      <div style={{ display: 'flex', gap: '4px' }}>
        <button style={btn(mode === 'timeline', '#2563EB')} onClick={() => onModeChange('timeline')}>Timeline</button>
        <button style={btn(mode === 'week', '#2563EB')} onClick={() => onModeChange('week')}>Week</button>
        <button style={btn(mode === 'month', '#2563EB')} onClick={() => onModeChange('month')}>Month</button>
      </div>

      <span style={{ color: '#D1D5DB' }}>|</span>

      {/* Date navigation */}
      <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
        <button style={arrowBtn} onClick={() => go(-7)}>{'<<'}</button>
        <button style={arrowBtn} onClick={() => go(-1)}>{'<'}</button>
        <span style={{ fontSize: '14px', fontWeight: 600, minWidth: '140px', textAlign: 'center' }}>
          {mode === 'month'
            ? currentDate.toLocaleDateString('en-US', { month: 'long', year: 'numeric' })
            : currentDate.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric', year: 'numeric' })}
        </span>
        <button style={arrowBtn} onClick={() => go(1)}>{'>'}</button>
        <button style={arrowBtn} onClick={() => go(7)}>{'>>'}</button>
        <button
          style={{ ...btn(false, '#2563EB'), fontSize: '12px' }}
          onClick={() => onDateChange(new Date())}
        >
          Today
        </button>
      </div>

      <span style={{ color: '#D1D5DB' }}>|</span>

      {/* Technician filter */}
      <select
        value={selectedTech}
        onChange={(e) => onTechChange(e.target.value)}
        style={{
          padding: '6px 10px', borderRadius: '6px', border: '1px solid #D1D5DB',
          fontSize: '13px', background: '#fff', cursor: 'pointer',
        }}
      >
        <option value="">All Technicians</option>
        {technicians.map(t => (
          <option key={t.id} value={t.id}>{t.name}</option>
        ))}
      </select>
    </div>
  );
}
