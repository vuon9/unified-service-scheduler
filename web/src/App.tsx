import type { Appointment, Technician, TabFilter, ToastMessage, ViewMode, ServiceBay } from './types';
import { useState, useEffect, useCallback } from 'react';
import { fetchAppointments, fetchTechnicians } from './api';
import BookingModal from './components/BookingModal';
import DetailModal from './components/DetailModal';
import ViewControls from './components/ViewControls';
import TimelineView from './components/TimelineView';
import WeekView from './components/WeekView';
import MonthView from './components/MonthView';
import Sidebar from './components/Sidebar';
import ToastContainer, { createToast } from './components/Toast';
import styles from './App.module.css';

const TABS: { key: TabFilter; label: string }[] = [
  { key: 'confirmed', label: 'Confirmed' },
  { key: 'cancelled', label: 'Cancelled' },
];

export default function App() {
  const [appointments, setAppointments] = useState<Appointment[]>([]);
  const [technicians, setTechnicians] = useState<Technician[]>([]);
  const [serviceBays, setServiceBays] = useState<ServiceBay[]>([]);
  const [activeTab, setActiveTab] = useState<TabFilter>('confirmed');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showBooking, setShowBooking] = useState(false);
  const [showDetail, setShowDetail] = useState<Appointment | null>(null);
  const [toasts, setToasts] = useState<ToastMessage[]>([]);
  const [viewMode, setViewMode] = useState<ViewMode>('timeline');
  const [currentDate, setCurrentDate] = useState(new Date());
  const [selectedTech, setSelectedTech] = useState('');
  const [selectedBay, setSelectedBay] = useState('');
  const [showSidebar, setShowSidebar] = useState(false);

  const addToast = useCallback((type: 'success' | 'error', message: string) => {
    const toast = createToast(type, message);
    setToasts((prev) => [...prev, toast]);
  }, []);

  const removeToast = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const loadAppointments = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchAppointments();
      setAppointments(data);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to load appointments';
      setError(message);
    } finally {
      setLoading(false);
    }
  }, []);

  const loadDataSources = useCallback(async () => {
    try {
      const [techs, bays] = await Promise.all([fetchTechnicians(), import('./api').then(m => m.fetchServiceBays())]);
      setTechnicians(techs ?? []);
      setServiceBays(bays ?? []);
    } catch { /* non-critical */ }
  }, []);

  useEffect(() => {
    loadAppointments();
    loadDataSources();
  }, [loadAppointments, loadDataSources]);

  const handleBooked = () => {
    loadAppointments();
    setShowBooking(false);
    addToast('success', 'Appointment booked!');
  };

  const handleCancelled = () => {
    loadAppointments();
    setShowDetail(null);
    addToast('success', 'Appointment cancelled');
  };

  const handleViewDetail = (appointment: Appointment) => {
    setShowDetail(appointment);
  };

  const handleSelectTech = (id: string) => {
    setSelectedTech(selectedTech === id ? '' : id);
    setSelectedBay('');
  };

  const handleSelectBay = (id: string) => {
    setSelectedBay(selectedBay === id ? '' : id);
    setSelectedTech('');
  };

  const hasFilters = selectedTech || selectedBay;

  const handleFabClick = () => {
    setShowSidebar(v => !v);
  };

  const sidebarContent = (fullWidth?: boolean) => (
    <Sidebar
      technicians={technicians}
      serviceBays={serviceBays}
      appointments={appointments}
      selectedTech={selectedTech}
      selectedBay={selectedBay}
      onSelectTech={handleSelectTech}
      onSelectBay={handleSelectBay}
      fullWidth={fullWidth}
    />
  );

  return (
    <div className={styles.app}>
      {/* Sticky: header + tabs + view controls */}
      <div style={{ position: 'sticky', top: 0, zIndex: 20, background: '#FFFFFF' }}>
        <header className={styles.header}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
            <svg viewBox="0 0 32 32" width="28" height="28" fill="none">
              <rect width="32" height="32" rx="8" fill="#2563EB" />
              <text x="16" y="22" textAnchor="middle" fill="#fff" fontSize="18" fontWeight="700" fontFamily="Inter, sans-serif">K</text>
            </svg>
            <h1 className={styles.logo}>Keyloop Scheduler</h1>
          </div>
          <button className={styles.newBookingBtn} onClick={() => setShowBooking(true)}>
            + New Booking
          </button>
        </header>

        <nav className={styles.tabBar}>
          <div className={styles.tabs}>
            {TABS.map((tab) => (
              <button
                key={tab.key}
                className={`${styles.tab} ${activeTab === tab.key ? styles.tabActive : ''}`}
                onClick={() => setActiveTab(tab.key)}
              >
                {tab.label}
              </button>
            ))}
          </div>
          {activeTab === 'confirmed' && (
            <button
              onClick={handleFabClick}
              style={{
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                width: '32px', height: '32px', border: 'none', borderRadius: '6px',
                background: showSidebar ? '#E5E7EB' : 'transparent',
                color: showSidebar ? '#374151' : '#6B7280',
                cursor: 'pointer', flexShrink: 0, transition: 'all 0.15s',
              }}
              aria-label="Toggle filters"
            >
              <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M4 4h16v2L14 12v6l-4 2v-8L4 6V4z"/>
              </svg>
            </button>
          )}
        </nav>

        {activeTab === 'confirmed' && (
          <ViewControls
            mode={viewMode}
            onModeChange={setViewMode}
            currentDate={currentDate}
            onDateChange={setCurrentDate}
          />
        )}
      </div>

      {/* Main content */}
      <div className={styles.allLayout}>
        {/* Desktop sidebar — always visible on confirmed tab */}
        {activeTab === 'confirmed' && (
          <div className={styles.sidebarWrapper}>
            {sidebarContent(false)}
          </div>
        )}

        {/* Mobile sidebar — shown inline when toggled, only on confirmed tab */}
        {activeTab === 'confirmed' && showSidebar && (
          <div className={styles.mobileSidebar}>
            {hasFilters && (
              <button
                onClick={() => { setSelectedTech(''); setSelectedBay(''); setShowSidebar(false); }}
                style={{
                  display: 'flex', alignItems: 'center', gap: '6px', width: '100%',
                  padding: '10px 16px', border: 'none', borderBottom: '1px solid #E5E7EB',
                  background: '#FEF2F2', color: '#DC2626', fontSize: '13px', fontWeight: 600,
                  cursor: 'pointer',
                }}
              >
                ✕ Clear filters
              </button>
            )}
            {sidebarContent(true)}
          </div>
        )}

        {/* Calendar / Timeline */}
        <div className={styles.mainContent}>
          {error ? (
            <div style={{ padding: '24px', textAlign: 'center' }}>
              <p style={{ color: '#DC2626', marginBottom: '12px', fontSize: '14px' }}>{error}</p>
              <button
                onClick={loadAppointments}
                style={{
                  padding: '8px 20px', background: '#2563EB', color: '#fff',
                  border: 'none', borderRadius: '6px', cursor: 'pointer', fontSize: '14px',
                }}
              >
                Retry
              </button>
            </div>
          ) : loading ? (
            <div style={{ padding: '24px', textAlign: 'center' }}>
              <div style={{
                width: '24px', height: '24px', border: '3px solid #E5E7EB',
                borderTopColor: '#2563EB', borderRadius: '50%', animation: 'spin 0.8s linear infinite',
                margin: '0 auto 12px',
              }} />
              <p style={{ color: '#6B7280', fontSize: '14px' }}>Loading appointments...</p>
            </div>
          ) : activeTab === 'confirmed' ? (
            <>
              {viewMode === 'timeline' && (
                <TimelineView
                  appointments={appointments}
                  selectedTech={selectedTech}
                  currentDate={currentDate}
                  onViewDetail={handleViewDetail}
                />
              )}
              {viewMode === 'week' && (
                <div style={{ overflowX: 'auto', WebkitOverflowScrolling: 'touch' }}>
                  <WeekView
                    appointments={appointments}
                    technicians={technicians}
                    selectedTech={selectedTech}
                    currentDate={currentDate}
                    onViewDetail={handleViewDetail}
                  />
                </div>
              )}
              {viewMode === 'month' && (
                <div style={{ overflowX: 'auto', WebkitOverflowScrolling: 'touch' }}>
                  <MonthView
                    appointments={appointments}
                    technicians={technicians}
                    selectedTech={selectedTech}
                    currentDate={currentDate}
                    onViewDetail={handleViewDetail}
                    onDateChange={setCurrentDate}
                  />
                </div>
              )}
            </>
          ) : (
            /* Cancelled tab — simple list */
            <div style={{ background: '#fff', borderRadius: '8px', border: '1px solid #E5E7EB' }}>
              {appointments.filter(a => a.status === 'cancelled').length === 0 ? (
                <div style={{ padding: '40px 16px', textAlign: 'center' }}>
                  <div style={{ fontSize: '15px', fontWeight: 600, color: '#6B7280' }}>No cancelled appointments</div>
                </div>
              ) : (
                appointments.filter(a => a.status === 'cancelled').map((a, i, arr) => {
                  const start = new Date(a.scheduled_start);
                  const end = new Date(a.scheduled_end);
                  return (
                    <div
                      key={a.id}
                      onClick={() => handleViewDetail(a)}
                      style={{
                        display: 'flex', gap: '12px', padding: '12px 16px',
                        borderBottom: i < arr.length - 1 ? '1px solid #F3F4F6' : 'none',
                        cursor: 'pointer', opacity: 0.7,
                      }}
                    >
                      <div style={{ width: '70px', flexShrink: 0, textAlign: 'right' }}>
                        <div style={{ fontSize: '13px', fontWeight: 600, color: '#6B7280' }}>
                          {start.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
                        </div>
                        <div style={{ fontSize: '11px', color: '#9CA3AF' }}>
                          {start.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}
                        </div>
                      </div>
                      <div style={{ width: '4px', flexShrink: 0, borderRadius: '2px', background: '#D1D5DB' }} />
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={{ fontSize: '13px', fontWeight: 600, color: '#6B7280', textDecoration: 'line-through' }}>
                          {a.service_type_name}
                        </div>
                        <div style={{ fontSize: '12px', color: '#9CA3AF' }}>
                          {a.vehicle_make} {a.vehicle_model}
                        </div>
                        <span style={{ fontSize: '11px', fontWeight: 600, color: '#DC2626', background: '#FEE2E2', padding: '1px 6px', borderRadius: '4px', display: 'inline-block', marginTop: '4px' }}>
                          CANCELLED
                        </span>
                      </div>
                    </div>
                  );
                })
              )}
            </div>
          )}
        </div>
      </div>

      {/* Mobile sidebar overlay — replaced by inline mobileSidebar */}
      {showBooking && (
        <BookingModal onClose={() => setShowBooking(false)} onBooked={handleBooked} />
      )}

      {showDetail && (
        <DetailModal
          appointment={showDetail}
          onClose={() => setShowDetail(null)}
          onCancelled={handleCancelled}
        />
      )}

      <ToastContainer toasts={toasts} removeToast={removeToast} />
    </div>
  );
}
