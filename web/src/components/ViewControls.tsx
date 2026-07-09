import type { ViewMode } from '../types';

interface ViewControlsProps {
  mode: ViewMode;
  onModeChange: (mode: ViewMode) => void;
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
  mode, onModeChange,
  currentDate, onDateChange,
}: ViewControlsProps) {
  const go = (direction: number, viewMode: ViewMode = 'timeline') => {
    const d = new Date(currentDate);
    if (viewMode === 'month') {
      d.setMonth(d.getMonth() + direction);
    } else if (viewMode === 'week') {
      d.setDate(d.getDate() + direction * 7);
    } else {
      d.setDate(d.getDate() + direction);
    }
    onDateChange(d);
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', padding: '8px 16px', gap: '6px' }}>
      {/* Row 1: mode buttons */}
      <div style={{ display: 'flex', gap: '4px' }}>
        <button style={btn(mode === 'timeline', '#2563EB')} onClick={() => onModeChange('timeline')}>Timeline</button>
        <button style={btn(mode === 'week', '#2563EB')} onClick={() => onModeChange('week')}>Week</button>
        <button style={btn(mode === 'month', '#2563EB')} onClick={() => onModeChange('month')}>Month</button>
      </div>

      {/* Row 2: date nav + Today — only for Week/Month */}
      {mode !== 'timeline' && (
        <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
          <button style={{...arrowBtn, fontSize: '16px', padding: '4px 8px'}} onClick={() => go(-1, mode)}>{'<'}</button>
          <span style={{ fontSize: '13px', fontWeight: 600, minWidth: '100px', textAlign: 'center', whiteSpace: 'nowrap' }}>
            {mode === 'month'
              ? currentDate.toLocaleDateString('en-US', { month: 'long', year: 'numeric' })
              : currentDate.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })}
          </span>
          <button style={{...arrowBtn, fontSize: '16px', padding: '4px 8px'}} onClick={() => go(1, mode)}>{'>'}</button>
          <button
            style={{ ...btn(false, '#2563EB'), fontSize: '12px', padding: '4px 8px' }}
            onClick={() => onDateChange(new Date())}
          >
            Today
          </button>
        </div>
      )}
    </div>
  );
}
