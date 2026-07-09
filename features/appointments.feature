Feature: Unified Service Scheduler

  Background:
    Given the seed data is loaded

  # ============================================================
  # Happy Path
  # ============================================================

  Scenario: T-01 - Simple booking with all resources available
    Given no appointments exist
    When customer "c1" books service "s1" for vehicle "v1" at dealership "d1" starting "2026-07-15T09:00:00+07:00"
    Then the booking status should be "confirmed"
    And the response should include a technician
    And the response should include a service bay
    And the scheduled end time should be "2026-07-15T10:00:00+07:00"
    And the appointment should be persisted

  Scenario: T-02 - Booking with limited qualified technicians
    Given no appointments exist
    When customer "c1" books service "s2" for vehicle "v1" at dealership "d1" starting "2026-07-15T09:00:00+07:00"
    Then the booking status should be "confirmed"
    And the assigned technician should be one of "t1, t3"
    And the response should include a service bay

  Scenario: T-03 - Cancel appointment frees resources for rebooking
    Given an appointment exists for customer "c1" with service "s1" for vehicle "v1" at dealership "d1" from "2026-07-15T09:00:00+07:00" to "2026-07-15T10:00:00+07:00"
    When customer "c1" cancels that appointment
    Then the cancellation status should be "cancelled"
    When customer "c2" books service "s1" for vehicle "v2" at dealership "d1" starting "2026-07-15T09:00:00+07:00"
    Then the booking status should be "confirmed"

  Scenario: T-04 - List appointments by customer
    Given an appointment exists for customer "c1" with service "s1" for vehicle "v1" at dealership "d1" from "2026-07-15T09:00:00+07:00" to "2026-07-15T10:00:00+07:00"
    And an appointment exists for customer "c1" with service "s2" for vehicle "v1" at dealership "d1" from "2026-07-16T09:00:00+07:00" to "2026-07-16T11:00:00+07:00"
    When I list appointments for customer "c1"
    Then the list should contain exactly "2" appointments
    And each appointment should include full details

  Scenario: T-05 - List appointments by date range
    Given an appointment exists for customer "c1" with service "s1" for vehicle "v1" at dealership "d1" from "2026-07-15T09:00:00+07:00" to "2026-07-15T10:00:00+07:00"
    And an appointment exists for customer "c2" with service "s1" for vehicle "v2" at dealership "d1" from "2026-07-16T09:00:00+07:00" to "2026-07-16T10:00:00+07:00"
    When I list appointments from "2026-07-15" to "2026-07-15"
    Then the list should contain only appointments on "2026-07-15"

  Scenario: T-06 - List appointments by dealership
    Given an appointment exists for customer "c1" with service "s1" for vehicle "v1" at dealership "d1" from "2026-07-15T09:00:00+07:00" to "2026-07-15T10:00:00+07:00"
    When I list appointments for dealership "d1"
    Then the list should contain all appointments at dealership "d1"

  # ============================================================
  # Conflict Cases
  # ============================================================

  Scenario: T-07 - Same time slot, system auto-selects other free bay
    Given an appointment exists for customer "c1" with service "s1" for vehicle "v1" at dealership "d1" from "2026-07-15T09:00:00+07:00" to "2026-07-15T10:00:00+07:00" using bay "b1"
    When customer "c2" books service "s1" for vehicle "v2" at dealership "d1" starting "2026-07-15T09:00:00+07:00"
    Then the booking status should be "confirmed"
    And the assigned service bay should be "b2"

  Scenario: T-08 - All bays occupied, no availability
    Given service bay "b1" is occupied from "2026-07-15T09:00:00+07:00" to "2026-07-15T10:00:00+07:00"
    And service bay "b2" is occupied from "2026-07-15T09:00:00+07:00" to "2026-07-15T10:00:00+07:00"
    When customer "c1" books service "s1" for vehicle "v1" at dealership "d1" starting "2026-07-15T09:00:00+07:00"
    Then the booking should fail with reason "no_service_bay_available"

  Scenario: T-09 - Tech occupied but other qualified tech is free
    Given an appointment exists for customer "c1" with service "s1" for vehicle "v1" at dealership "d1" from "2026-07-15T09:00:00+07:00" to "2026-07-15T10:00:00+07:00" using technician "t1"
    When customer "c2" books service "s1" for vehicle "v2" at dealership "d1" starting "2026-07-15T09:00:00+07:00"
    Then the booking status should be "confirmed"
    And the assigned technician should be "t2"

  Scenario: T-10 - All qualified technicians occupied
    Given technician "t1" is occupied from "2026-07-15T09:00:00+07:00" to "2026-07-15T10:00:00+07:00"
    And technician "t2" is occupied from "2026-07-15T09:00:00+07:00" to "2026-07-15T10:00:00+07:00"
    When customer "c1" books service "s1" for vehicle "v1" at dealership "d1" starting "2026-07-15T09:00:00+07:00"
    Then the booking should fail with reason "no_qualified_technician"

  Scenario: T-11 - Technician free but not qualified for the service
    Given technician "t1" is occupied from "2026-07-15T09:00:00+07:00" to "2026-07-15T11:00:00+07:00"
    And technician "t3" is occupied from "2026-07-15T09:00:00+07:00" to "2026-07-15T11:00:00+07:00"
    When customer "c1" books service "s2" for vehicle "v1" at dealership "d1" starting "2026-07-15T09:00:00+07:00"
    Then the booking should fail with reason "no_qualified_technician"
    And technician "t2" should not be assigned

  Scenario: T-12 - Bay free but only qualified technician occupied
    Given an appointment exists for customer "c1" with service "s3" for vehicle "v1" at dealership "d1" from "2026-07-15T09:00:00+07:00" to "2026-07-15T10:30:00+07:00" using technician "t2"
    When customer "c2" books service "s3" for vehicle "v2" at dealership "d1" starting "2026-07-15T10:00:00+07:00"
    Then the booking should fail with reason "no_qualified_technician"

  # ============================================================
  # Boundary Cases
  # ============================================================

  Scenario: T-13 - Adjacent appointments with no overlap
    Given an appointment exists for customer "c1" with service "s1" for vehicle "v1" at dealership "d1" from "2026-07-15T09:00:00+07:00" to "2026-07-15T10:00:00+07:00"
    When customer "c2" books service "s1" for vehicle "v2" at dealership "d1" starting "2026-07-15T10:00:00+07:00"
    Then the booking status should be "confirmed"

  Scenario: T-14 - One minute overlap causes conflict
    Given an appointment exists for customer "c1" with service "s1" for vehicle "v1" at dealership "d1" from "2026-07-15T09:00:00+07:00" to "2026-07-15T10:00:00+07:00"
    When customer "c2" books service "s1" for vehicle "v2" at dealership "d1" starting "2026-07-15T09:59:00+07:00"
    Then the booking should fail with reason "unavailable"

  Scenario: T-15 - Partial overlap with existing appointment
    Given an appointment exists for customer "c1" with service "s2" for vehicle "v1" at dealership "d1" from "2026-07-15T10:00:00+07:00" to "2026-07-15T12:00:00+07:00"
    When customer "c2" books service "s1" for vehicle "v2" at dealership "d1" starting "2026-07-15T11:00:00+07:00"
    Then the booking should fail with reason "unavailable"

  # ============================================================
  # Validation Cases
  # ============================================================

  Scenario Outline: T-16, T-18, T-19, T-20 - Validation of booking parameters
    Given no appointments exist
    When customer "<customer>" books service "<service>" for vehicle "<vehicle>" at dealership "<dealership>" starting "<start>"
    Then the booking should fail with reason "<reason>"

    Examples:
      | customer | service | vehicle | dealership | start                      | reason                             |
      | c1       | s1      | v1      | d1         | 2020-01-01T09:00:00+07:00 | "past_start_time"                  |
      | c1       | s1      | vx      | d1         | 2026-07-15T09:00:00+07:00 | "vehicle_not_found"                |
      | c1       | sx      | v1      | d1         | 2026-07-15T09:00:00+07:00 | "service_type_not_found"           |
      | c1       | s1      | v1      | dx         | 2026-07-15T09:00:00+07:00 | "dealership_not_found"             |

  Scenario: T-17 - Customer does not own the vehicle
    Given customer "c1" has vehicle "v1"
    When customer "c2" books service "s1" for vehicle "v1" at dealership "d1" starting "2026-07-15T09:00:00+07:00"
    Then the booking should fail with reason "customer_does_not_own_vehicle"

  Scenario: T-21 - Cancel non-existent appointment
    Given appointment "999" does not exist
    When customer "c1" cancels appointment "999"
    Then the booking should fail with reason "appointment_not_found"

  Scenario: T-22 - Cancel already-cancelled appointment
    Given appointment "100" is already cancelled
    When customer "c1" cancels appointment "100"
    Then the booking should fail with reason "appointment_already_cancelled"

  Scenario: T-23 - Cancel a confirmed appointment
    Given an appointment with id "200" and status "confirmed" exists
    When customer "c1" cancels appointment "200"
    Then the appointment status should be "cancelled"
    And resources should be freed for that time slot

  # ============================================================
  # Race Condition
  # ============================================================

  # NOTE: This scenario requires parallel execution in the step definition.
  # The When step must spawn two concurrent booking requests and collect both results.
  Scenario: T-24 - Two customers book same resources concurrently
    Given only 1 qualified technician and 1 service bay are free from "2026-07-15T09:00:00+07:00" to "2026-07-15T10:00:00+07:00"
    When customer "c1" and customer "c2" simultaneously book service "s1" for their vehicles at dealership "d1" starting "2026-07-15T09:00:00+07:00"
    Then exactly 1 booking should succeed with status "confirmed" and 1 should fail with reason "unavailable"
