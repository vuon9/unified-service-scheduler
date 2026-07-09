import type { Appointment } from '../types';
import styles from './AppointmentCard.module.css';

interface AppointmentCardProps {
  appointment: Appointment;
  onViewDetail: (appointment: Appointment) => void;
}

export default function AppointmentCard({ appointment, onViewDetail }: AppointmentCardProps) {
  const isCancelled = appointment.status === 'cancelled';

  const startDate = new Date(appointment.scheduled_start);
  const endDate = new Date(appointment.scheduled_end);

  const dateStr = startDate.toLocaleDateString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });

  const timeStr = `${startDate.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })} - ${endDate.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}`;

  const vehicleName = `${appointment.vehicle_make ?? ''} ${appointment.vehicle_model ?? ''}`.trim();

  return (
    <div className={`${styles.card} ${isCancelled ? styles.cancelled : ''}`}>
      <div className={styles.header}>
        <span className={styles.date}>{dateStr}</span>
        <span className={styles.time}>{timeStr}</span>
        <span className={`${styles.badge} ${isCancelled ? styles.badgeCancelled : styles.badgeConfirmed}`}>
          {appointment.status.toUpperCase()}
        </span>
      </div>
      <div className={styles.body}>
        <div className={styles.info}>
          <span className={styles.label}>Vehicle</span>
          <span className={styles.value}>{vehicleName}</span>
        </div>
        <div className={styles.info}>
          <span className={styles.label}>Service</span>
          <span className={styles.value}>{appointment.service_type_name}</span>
        </div>
        <div className={styles.info}>
          <span className={styles.label}>Technician</span>
          <span className={styles.value}>{appointment.technician_name}</span>
        </div>
        <div className={styles.info}>
          <span className={styles.label}>Bay</span>
          <span className={styles.value}>{appointment.service_bay_name}</span>
        </div>
      </div>
      <button className={styles.viewBtn} onClick={() => onViewDetail(appointment)}>
        View Details
      </button>
    </div>
  );
}
