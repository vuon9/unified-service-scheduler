import { useState, useEffect, useRef, useCallback } from 'react';
import type { Vehicle, ServiceType, AvailabilityResponse } from '../types';
import { fetchVehicles, fetchServiceTypes, checkAvailability, createAppointment } from '../api';
import styles from './BookingModal.module.css';

interface BookingModalProps {
  onClose: () => void;
  onBooked: () => void;
}

const DEFAULT_DEALERSHIP_ID = 'd1';

export default function BookingModal({ onClose, onBooked }: BookingModalProps) {
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [serviceTypes, setServiceTypes] = useState<ServiceType[]>([]);
  const [selectedVehicle, setSelectedVehicle] = useState('');
  const [selectedService, setSelectedService] = useState('');
  const [datetime, setDatetime] = useState('');
  const [notes, setNotes] = useState('');
  const [availability, setAvailability] = useState<AvailabilityResponse | null>(null);
  const [checkingAvail, setCheckingAvail] = useState(false);
  const [checkTouched, setCheckTouched] = useState(false);
  const [booking, setBooking] = useState(false);
  const [error, setError] = useState('');
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    Promise.all([fetchVehicles(), fetchServiceTypes()])
      .then(([v, s]) => {
        setVehicles(v);
        setServiceTypes(s);
      })
      .catch((err) => setError('Failed to load vehicles and services: ' + err.message));
  }, []);

  const selectedServiceType = serviceTypes.find((s) => s.id === selectedService);
  const selectedVehicleObj = vehicles.find((v) => v.id === selectedVehicle);

  const buildScheduledStart = useCallback((): string => {
    if (!datetime) return '';
    return new Date(datetime).toISOString();
  }, [datetime]);

  // Auto-check availability with debounce
  useEffect(() => {
    if (!selectedVehicle || !selectedService || !datetime) {
      setCheckTouched(false);
      setAvailability(null);
      return;
    }
    setCheckTouched(true);
    if (timerRef.current) clearTimeout(timerRef.current);
    const timer = setTimeout(async () => {
      setCheckingAvail(true);
      setError('');
      try {
        const scheduledStart = buildScheduledStart();
        const result = await checkAvailability({
          dealership_id: DEFAULT_DEALERSHIP_ID,
          service_type_id: selectedService,
          scheduled_start: scheduledStart,
        });
        setAvailability(result);
      } catch (err: unknown) {
        const message = err instanceof Error ? err.message : 'Failed to check availability';
        setError(message);
        setAvailability(null);
      } finally {
        setCheckingAvail(false);
      }
    }, 500);
    timerRef.current = timer;
    return () => {
      clearTimeout(timer);
      timerRef.current = null;
    };
  }, [selectedVehicle, selectedService, datetime, buildScheduledStart]);

  const handleBook = async () => {
    if (!availability?.available) return;
    setBooking(true);
    setError('');
    try {
      const scheduledStart = buildScheduledStart();
      await createAppointment({
        customer_id: selectedVehicleObj?.customer_id ?? '',
        vehicle_id: selectedVehicle,
        dealership_id: DEFAULT_DEALERSHIP_ID,
        service_type_id: selectedService,
        scheduled_start: scheduledStart,
        notes: notes || undefined,
      });
      onBooked();
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Booking failed';
      setError(message);
    } finally {
      setBooking(false);
    }
  };

  const canBook = availability?.available === true && !booking;

  // Compute the min datetime for the input
  const now = new Date();
  const minDatetime = now.toISOString().slice(0, 16);

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <div className={styles.header}>
          <h2 className={styles.title}>Book Appointment</h2>
          <button className={styles.closeBtn} onClick={onClose}>X</button>
        </div>

        <div className={styles.body}>
          <div className={styles.row}>
            <div className={styles.field}>
              <label className={styles.label}>Vehicle</label>
              <select
                className={styles.select}
                value={selectedVehicle}
                onChange={(e) => setSelectedVehicle(e.target.value)}
              >
                <option value="">-- Select --</option>
                {vehicles.map((v) => (
                  <option key={v.id} value={v.id}>
                    {v.make} {v.model} {v.year}
                  </option>
                ))}
              </select>
            </div>
            <div className={styles.field}>
              <label className={styles.label}>Service</label>
              <select
                className={styles.select}
                value={selectedService}
                onChange={(e) => setSelectedService(e.target.value)}
              >
                <option value="">-- Select --</option>
                {serviceTypes.map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name} ({s.duration_minutes} min)
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div className={styles.field}>
            <label className={styles.label}>Date &amp; Time</label>
            <input
              type="datetime-local"
              className={styles.input}
              value={datetime}
              min={minDatetime}
              onChange={(e) => setDatetime(e.target.value)}
            />
            {selectedServiceType && (
              <span className={styles.hint}>({selectedServiceType.duration_minutes} min)</span>
            )}
          </div>

          {/* Availability alert */}
          {checkTouched && (
            <div className={`${styles.alert} ${availability?.available ? styles.alertGreen : styles.alertRed}`}>
              {checkingAvail ? (
                <span className={styles.alertText}>Checking availability...</span>
              ) : availability === null ? (
                <span className={styles.alertText}>Complete all fields to check</span>
              ) : availability.available ? (
                <div className={styles.alertInner}>
                  <span className={styles.alertIcon}>✓</span>
                  <span className={styles.alertText}>Available!</span>
                  <span className={styles.alertSub}>
                    {availability.available_technicians?.map(t => t.name).join(', ')} &middot;{' '}
                    {availability.available_bays?.map(b => b.name).join(', ')}
                  </span>
                </div>
              ) : (
                <div className={styles.alertInner}>
                  <span className={styles.alertIcon}>✕</span>
                  <span className={styles.alertText}>No resources available for this time</span>
                </div>
              )}
            </div>
          )}

          {/* Notes */}
          <div className={styles.field}>
            <label className={styles.label}>Notes (optional)</label>
            <textarea
              className={styles.textarea}
              placeholder="e.g. customer preferences, special requests..."
              rows={2}
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
            />
          </div>

          {error && <p className={styles.error}>{error}</p>}
        </div>

        <div className={styles.actions}>
          <button className={styles.cancelBtn} onClick={onClose} disabled={booking}>
            Cancel
          </button>
          <button
            className={styles.bookBtn}
            onClick={handleBook}
            disabled={!canBook}
          >
            {booking ? 'Booking...' : 'Book Appointment'}
          </button>
        </div>
      </div>
    </div>
  );
}
