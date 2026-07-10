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
  { key: 'all', label: 'All' },
  { key: 'confirmed', label: 'Confirmed' },
  { key: 'cancelled', label: 'Cancelled' },
];

export default function App() {
  const [appointments, setAppointments] = useState<Appointment[]>([]);
  const [technicians, setTechnicians] = useState<Technician[]>([]);
  const [serviceBays, setServiceBays] = useState<ServiceBay[]>([]);
  const [activeTab, setActiveTab] = useState<TabFilter>('all');
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

  const filteredAppointments = appointments.filter((a) => {
    if (activeTab === 'all') return true;
    return a.status === activeTab;
  });

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

  return (
    <div className={styles.app}>
      <div style={{ position: 'sticky', top: 0, zIndex: 20, background: '#FFFFFF' }}>
        <header className={styles.header}>
          <h1 className={styles.logo}>Keyloop Scheduler</h1>
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
          <button className={styles.filterToggle} onClick={() => setShowSidebar(!showSidebar)}>
            {showSidebar ? 'Hide Filters' : 'Filters'}
          </button>
        </nav>
      </div>

      <div className={styles.allLayout}>
        <div className={`${styles.sidebarWrapper} ${!showSidebar ? styles.sidebarWrapperMobileHidden : ''}`}>
          <Sidebar
            technicians={technicians}
            serviceBays={serviceBays}
            appointments={appointments}
            selectedTech={selectedTech}
            selectedBay={selectedBay}
            onSelectTech={(id) => { setSelectedTech(selectedTech === id ? '' : id); setSelectedBay(''); }}
            onSelectBay={(id) => { setSelectedBay(selectedBay === id ? '' : id); setSelectedTech(''); }}
          />
        </div>
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
          ) : (
            <>
              <ViewControls
                mode={viewMode}
                onModeChange={setViewMode}
                currentDate={currentDate}
                onDateChange={setCurrentDate}
              />
              {viewMode === 'timeline' && (
                <TimelineView
                  appointments={appointments}
                  technicians={selectedTech ? technicians.filter(t => t.id === selectedTech) : technicians}
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
          )}
        </div>
      </div>

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
