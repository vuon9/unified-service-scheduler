import type {
  Appointment,
  AppointmentListResponse,
  Vehicle,
  ServiceType,
  AvailabilityRequest,
  AvailabilityResponse,
  CreateAppointmentRequest,
} from './types';

const API_BASE = '/api/v1';

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const url = `${API_BASE}${path}`;
  const res = await fetch(url, {
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    ...options,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(body.error || body.message || `Request failed: ${res.status}`);
  }

  return res.json();
}

export function fetchAppointments(params?: { status?: string; customer_id?: string; dealership_id?: string }): Promise<Appointment[]> {
  const searchParams = new URLSearchParams();
  if (params?.status) searchParams.set('status', params.status);
  if (params?.customer_id) searchParams.set('customer_id', params.customer_id);
  if (params?.dealership_id) searchParams.set('dealership_id', params.dealership_id);
  const query = searchParams.toString() ? `?${searchParams.toString()}` : '';
  return apiFetch<AppointmentListResponse>(`/appointments${query}`).then(r => r.appointments ?? []);
}

export function fetchAppointment(id: string): Promise<Appointment> {
  return apiFetch<Appointment>(`/appointments/${id}`);
}

export function createAppointment(data: CreateAppointmentRequest): Promise<Appointment> {
  return apiFetch<Appointment>('/appointments', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export function cancelAppointment(id: string): Promise<{ id: string; status: string }> {
  return apiFetch<{ id: string; status: string }>(`/appointments/${id}/cancel`, {
    method: 'POST',
  });
}

export function checkAvailability(data: AvailabilityRequest): Promise<AvailabilityResponse> {
  return apiFetch<AvailabilityResponse>('/availability', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export function fetchVehicles(): Promise<Vehicle[]> {
  return apiFetch<Vehicle[]>('/vehicles');
}

export function fetchServiceTypes(): Promise<ServiceType[]> {
  return apiFetch<ServiceType[]>('/service-types');
}

export function fetchTechnicians(): Promise<Technician[]> {
  return apiFetch<Technician[]>('/technicians');
}

export function fetchServiceBays(): Promise<ServiceBay[]> {
  return apiFetch<ServiceBay[]>('/service-bays');
}
