import type { Appointment, TabFilter } from '../types';
import AppointmentCard from './AppointmentCard';
import styles from './AppointmentList.module.css';

interface AppointmentListProps {
  appointments: Appointment[];
  activeTab: TabFilter;
  loading: boolean;
  error: string | null;
  onRetry: () => void;
  onViewDetail: (appointment: Appointment) => void;
  onNewBooking: () => void;
}

export default function AppointmentList({
  appointments,
  activeTab,
  loading,
  error,
  onRetry,
  onViewDetail,
  onNewBooking,
}: AppointmentListProps) {
  if (loading) {
    return (
      <div className={styles.stateContainer}>
        <div className={styles.spinner} />
        <p className={styles.stateText}>Loading appointments...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className={styles.stateContainer}>
        <p className={styles.errorText}>Could not load appointments</p>
        <p className={styles.errorDetail}>{error}</p>
        <button className={styles.retryBtn} onClick={onRetry}>
          Retry
        </button>
      </div>
    );
  }

  if (appointments.length === 0) {
    if (activeTab !== 'all') {
      return (
        <div className={styles.stateContainer}>
          <p className={styles.stateText}>No appointments match this filter</p>
        </div>
      );
    }
    return (
      <div className={styles.stateContainer}>
        <p className={styles.stateText}>No appointments yet. Book your first one!</p>
        <button className={styles.ctaBtn} onClick={onNewBooking}>
          + New Booking
        </button>
      </div>
    );
  }

  return (
    <div className={styles.list}>
      {appointments.map((appt) => (
        <AppointmentCard key={appt.id} appointment={appt} onViewDetail={onViewDetail} />
      ))}
    </div>
  );
}
