import { useState } from 'react';
import type { Appointment } from '../types';
import { cancelAppointment } from '../api';
import styles from './DetailModal.module.css';

interface DetailModalProps {
  appointment: Appointment;
  onClose: () => void;
  onCancelled: () => void;
}

export default function DetailModal({ appointment, onClose, onCancelled }: DetailModalProps) {
  const [showConfirm, setShowConfirm] = useState(false);
  const [cancelling, setCancelling] = useState(false);
  const [error, setError] = useState('');
  const isCancelled = appointment.status === 'cancelled';

  const handleCancel = async () => {
    setCancelling(true);
    setError('');
    try {
      await cancelAppointment(appointment.id);
      onCancelled();
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Cancellation failed';
      setError(message);
    } finally {
      setCancelling(false);
    }
  };

  const startDate = new Date(appointment.scheduled_start);
  const endDate = new Date(appointment.scheduled_end);

  const dateStr = startDate.toLocaleDateString('en-US', {
    weekday: 'long',
    month: 'long',
    day: 'numeric',
    year: 'numeric',
  });

  const timeStr = `${startDate.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })} - ${endDate.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}`;

  const vehicleName = `${appointment.vehicle_make ?? ''} ${appointment.vehicle_model ?? ''}`.trim();
  const createdAtStr = new Date(appointment.created_at).toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <div className={styles.header}>
          <h2 className={styles.title}>Appointment Detail</h2>
          <button className={styles.closeBtn} onClick={onClose}>X</button>
        </div>

        <div className={styles.body}>
          <span className={`${styles.statusBadge} ${isCancelled ? styles.badgeCancelled : styles.badgeConfirmed}`}>
            STATUS: {appointment.status.toUpperCase()}
          </span>

          <div className={styles.details}>
            <DetailRow label="Vehicle" value={vehicleName} />
            <DetailRow label="Service" value={appointment.service_type_name ?? '-'} />
            <DetailRow label="Date" value={dateStr} />
            <DetailRow label="Time" value={timeStr} />
            <DetailRow label="Customer" value={appointment.customer_name ?? '-'} />
            <DetailRow label="Technician" value={appointment.technician_name ?? '-'} />
            <DetailRow label="Bay" value={appointment.service_bay_name ?? '-'} />
            <DetailRow label="Booked at" value={createdAtStr} />
            {appointment.notes && appointment.notes.trim() && (
              <div className={styles.row} style={{ flexDirection: 'column', alignItems: 'flex-start' }}>
                <span className={styles.label}>Notes</span>
                <span className={styles.value} style={{ whiteSpace: 'pre-wrap', lineHeight: 1.5, marginTop: '2px' }}>
                  {appointment.notes}
                </span>
              </div>
            )}
          </div>
        </div>

        {error && <p className={styles.error}>{error}</p>}

        <div className={styles.actions}>
          {!isCancelled && !showConfirm && (
            <button className={styles.cancelBtn} onClick={() => setShowConfirm(true)}>
              Cancel Appointment
            </button>
          )}
          {showConfirm && (
            <div className={styles.confirmBox}>
              <p className={styles.confirmText}>Are you sure you want to cancel this appointment?</p>
              <div className={styles.confirmActions}>
                <button
                  className={styles.noBtn}
                  onClick={() => setShowConfirm(false)}
                  disabled={cancelling}
                >
                  No, keep it
                </button>
                <button
                  className={styles.yesBtn}
                  onClick={handleCancel}
                  disabled={cancelling}
                >
                  {cancelling ? 'Cancelling...' : 'Yes, cancel'}
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className={styles.row}>
      <span className={styles.label}>{label}</span>
      <span className={styles.value}>{value}</span>
    </div>
  );
}
