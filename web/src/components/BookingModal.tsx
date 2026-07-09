import { useState, useEffect } from 'react';
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
  const [date, setDate] = useState('');
  const [time, setTime] = useState('09:00');
  const [availability, setAvailability] = useState<AvailabilityResponse | null>(null);
  const [checkingAvail, setCheckingAvail] = useState(false);
  const [booking, setBooking] = useState(false);
  const [error, setError] = useState('');

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

  // Build ISO-8601 datetime with timezone for the API
  const buildScheduledStart = (): string => {
    if (!date || !time) return '';
    // Use local timezone offset to build proper ISO string
    const dt = new Date(`${date}T${time}:00`);
    return dt.toISOString();
  };

  const handleCheckAvailability = async () => {
    if (!selectedVehicle || !selectedService || !date || !time) {
      setError('Please fill in all fields');
      return;
    }
    setError('');
    setAvailability(null);
    setCheckingAvail(true);
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
    } finally {
      setCheckingAvail(false);
    }
  };

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
      });
      onBooked();
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Booking failed';
      setError(message);
    } finally {
      setBooking(false);
    }
  };

  const canCheck = selectedVehicle && selectedService && date && time;

  // Format technician names for display
  const techNames = availability?.available_technicians?.map(t => t.name).join(', ') ?? '';
  const bayNames = availability?.available_bays?.map(b => b.name).join(', ') ?? '';

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <div className={styles.header}>
          <h2 className={styles.title}>Book New Appointment</h2>
          <button className={styles.closeBtn} onClick={onClose}>X</button>
        </div>

        <div className={styles.steps}>
          {/* Step 1 */}
          <div className={styles.step}>
            <h3 className={styles.stepTitle}>Step 1: Select Vehicle &amp; Service</h3>
            <div className={styles.formRow}>
              <div className={styles.field}>
                <label className={styles.fieldLabel}>Vehicle</label>
                <select
                  className={styles.select}
                  value={selectedVehicle}
                  onChange={(e) => setSelectedVehicle(e.target.value)}
                >
                  <option value="">-- Select Vehicle --</option>
                  {vehicles.map((v) => (
                    <option key={v.id} value={v.id}>
                      {v.make} {v.model} {v.year}
                    </option>
                  ))}
                </select>
              </div>
              <div className={styles.field}>
                <label className={styles.fieldLabel}>Service</label>
                <select
                  className={styles.select}
                  value={selectedService}
                  onChange={(e) => setSelectedService(e.target.value)}
                >
                  <option value="">-- Select Service --</option>
                  {serviceTypes.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name} ({s.duration_minutes} min)
                    </option>
                  ))}
                </select>
              </div>
            </div>
          </div>

          {/* Step 2 */}
          <div className={styles.step}>
            <h3 className={styles.stepTitle}>Step 2: Pick Date &amp; Time</h3>
            <div className={styles.formRow}>
              <div className={styles.field}>
                <label className={styles.fieldLabel}>Date</label>
                <input
                  type="date"
                  className={styles.input}
                  value={date}
                  min={new Date().toISOString().split('T')[0]}
                  onChange={(e) => {
                    setDate(e.target.value);
                    setAvailability(null);
                  }}
                />
              </div>
              <div className={styles.field}>
                <label className={styles.fieldLabel}>Time</label>
                <input
                  type="time"
                  className={styles.input}
                  value={time}
                  onChange={(e) => {
                    setTime(e.target.value);
                    setAvailability(null);
                  }}
                />
              </div>
            </div>
            {selectedServiceType && (
              <p className={styles.durationHint}>
                Estimated duration: {selectedServiceType.duration_minutes} minutes
              </p>
            )}
          </div>

          {/* Step 3 */}
          <div className={styles.step}>
            <h3 className={styles.stepTitle}>Step 3: Check Availability</h3>
            <button
              className={styles.checkBtn}
              onClick={handleCheckAvailability}
              disabled={!canCheck || checkingAvail}
            >
              {checkingAvail ? 'Checking...' : 'Check Availability'}
            </button>

            {availability && (
              <div className={`${styles.result} ${availability.available ? styles.available : styles.unavailable}`}>
                {availability.available ? (
                  <>
                    <span className={styles.resultIcon}>[OK]</span>
                    <div className={styles.resultDetails}>
                      <p className={styles.resultText}>Available!</p>
                      {techNames && (
                        <p className={styles.resultInfo}>Technician: {techNames}</p>
                      )}
                      {bayNames && (
                        <p className={styles.resultInfo}>Bay: {bayNames}</p>
                      )}
                    </div>
                  </>
                ) : (
                  <>
                    <span className={styles.resultIcon}>[X]</span>
                    <div className={styles.resultDetails}>
                      <p className={styles.resultText}>No resources available for this time slot.</p>
                    </div>
                  </>
                )}
              </div>
            )}
          </div>
        </div>

        {error && <p className={styles.error}>{error}</p>}

        <div className={styles.actions}>
          <button className={styles.cancelBtn} onClick={onClose} disabled={booking}>
            Cancel
          </button>
          <button
            className={styles.bookBtn}
            onClick={handleBook}
            disabled={!availability?.available || booking}
          >
            {booking ? 'Booking...' : 'Book Appointment'}
          </button>
        </div>
      </div>
    </div>
  );
}
