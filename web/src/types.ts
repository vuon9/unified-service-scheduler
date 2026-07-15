export interface Appointment {
  id: string;
  customer_id: string;
  vehicle_id: string;
  dealership_id: string;
  service_type_id: string;
  technician_id: string;
  service_bay_id: string;
  scheduled_start: string;
  scheduled_end: string;
  status: 'confirmed' | 'cancelled' | 'completed';
  created_at: string;
  notes?: string;
  // Joined fields from backend
  customer_name?: string;
  vehicle_make?: string;
  vehicle_model?: string;
  service_type_name?: string;
  technician_name?: string;
  service_bay_name?: string;
}

export interface AppointmentListResponse {
  appointments: Appointment[];
  count: number;
}

export interface Vehicle {
  id: string;
  customer_id: string;
  vin: string;
  make: string;
  model: string;
  year: number;
}

export interface ServiceType {
  id: string;
  name: string;
  duration_minutes: number;
  description: string;
}

export interface Technician {
  id: string;
  dealership_id: string;
  name: string;
}

export interface ServiceBay {
  id: string;
  dealership_id: string;
  name: string;
}

export interface AvailabilityRequest {
  dealership_id: string;
  service_type_id: string;
  scheduled_start: string;
}

export interface AvailabilityResponse {
  available: boolean;
  available_technicians: Technician[];
  available_bays: ServiceBay[];
}

export interface CreateAppointmentRequest {
  customer_id: string;
  vehicle_id: string;
  dealership_id: string;
  service_type_id: string;
  scheduled_start: string;
  technician_id?: string;
  service_bay_id?: string;
  notes?: string;
}

export type TabFilter = 'confirmed' | 'cancelled';

export type ViewMode = 'timeline' | 'week' | 'month';


export type ToastType = 'success' | 'error';

export interface ToastMessage {
  id: number;
  type: ToastType;
  message: string;
}
