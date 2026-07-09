import type { Appointment, Technician, TabFilter, ToastMessage, ViewMode, ServiceBay } from './types';
import { useState, useEffect, useCallback } from 'react';
import { fetchAppointments, fetchTechnicians } from './api';
import AppointmentList from './components/AppointmentList';
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

  const isAll = activeTab === 'all';

  return (
    <div className={styles.app}>
      <header className={styles.header}>
        <h1 className={styles.logo}>Keyloop Scheduler</h1>
        <button className={styles.newBookingBtn} onClick={() => setShowBooking(true)}>
          + New Booking
        </button>
      </header>

      <nav className={styles.tabs}>
        {TABS.map((tab) => (
          <button
            key={tab.key}
            className={`${styles.tab} ${activeTab === tab.key ? styles.tabActive : ''}`}
            onClick={() => setActiveTab(tab.key)}
          >
            {tab.label}
          </button>
        ))}
      </nav>

      {isAll && (
        <div className={styles.allLayout}>
          <Sidebar
            technicians={technicians}
            serviceBays={serviceBays}
            appointments={appointments}
            selectedTech={selectedTech}
            selectedBay={selectedBay}
            onSelectTech={(id) => { setSelectedTech(selectedTech === id ? '' : id); setSelectedBay(''); }}
            onSelectBay={(id) => { setSelectedBay(selectedBay === id ? '' : id); setSelectedTech(''); }}
          />
          <div className={styles.mainContent}>
            {error ? (
              <AppointmentList
                appointments={filteredAppointments}
                activeTab={activeTab}
                loading={false}
                error={error}
                onRetry={loadAppointments}
                onViewDetail={handleViewDetail}
                onNewBooking={() => setShowBooking(true)}
              />
            ) : loading ? (
              <AppointmentList
                appointments={filteredAppointments}
                activeTab={activeTab}
                loading={true}
                error={null}
                onRetry={loadAppointments}
                onViewDetail={handleViewDetail}
                onNewBooking={() => setShowBooking(true)}
              />
            ) : (
              <>
                <ViewControls
                  mode={viewMode}
                  onModeChange={setViewMode}
                  technicians={technicians}
                  selectedTech={selectedTech}
                  onTechChange={(id) => { setSelectedTech(id); setSelectedBay(''); }}
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
                  <WeekView
                    appointments={appointments}
                    technicians={technicians}
                    selectedTech={selectedTech}
                    currentDate={currentDate}
                    onViewDetail={handleViewDetail}
                  />
                )}
                {viewMode === 'month' && (
                  <MonthView
                    appointments={appointments}
                    technicians={technicians}
                    selectedTech={selectedTech}
                    currentDate={currentDate}
                    onViewDetail={handleViewDetail}
                    onDateChange={setCurrentDate}
                  />
                )}
              </>
            )}
          </div>
        </div>
      )}

      {!isAll && (
        <main className={styles.main}>
          <AppointmentList
            appointments={filteredAppointments}
            activeTab={activeTab}
            loading={loading}
            error={error}
            onRetry={loadAppointments}
            onViewDetail={handleViewDetail}
            onNewBooking={() => setShowBooking(true)}
          />
        </main>
      )}

      {isAll && !error && loading && (
        <main className={styles.main}>
          <AppointmentList
            appointments={filteredAppointments}
            activeTab={activeTab}
            loading={loading}
            error={error}
            onRetry={loadAppointments}
            onViewDetail={handleViewDetail}
            onNewBooking={() => setShowBooking(true)}
          />
        </main>
      )}

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
