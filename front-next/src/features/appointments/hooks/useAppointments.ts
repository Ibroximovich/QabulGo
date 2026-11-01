import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { appointmentApi } from '../api/appointmentApi';
import {
  CreateAppointmentRequest,
  PatientAppointmentFilter,
  ClinicAppointmentFilter,
} from '../types';

export function useCreateAppointment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      payload,
      idempotencyKey,
    }: {
      payload: CreateAppointmentRequest;
      idempotencyKey?: string;
    }) => {
      const res = await appointmentApi.bookAppointment(payload, idempotencyKey);
      return res.data;
    },
    onSuccess: (data) => {
      // Invalidate availability for this clinic/doctor so the slot disappears
      queryClient.invalidateQueries({ queryKey: ['publicAvailability'] });
      // Invalidate patient's appointments list
      queryClient.invalidateQueries({ queryKey: ['myAppointments'] });
      // Invalidate clinic appointments if relevant
      queryClient.invalidateQueries({ queryKey: ['clinicAppointments'] });
    },
  });
}

export function useMyAppointments(filter?: PatientAppointmentFilter) {
  return useQuery({
    queryKey: ['myAppointments', filter],
    queryFn: async () => {
      const res = await appointmentApi.getMyAppointments(filter);
      return res.data;
    },
  });
}

export function useMyAppointmentDetail(id: string) {
  return useQuery({
    queryKey: ['myAppointment', id],
    queryFn: async () => {
      if (!id) return null;
      const res = await appointmentApi.getMyAppointmentById(id);
      return res.data;
    },
    enabled: Boolean(id),
  });
}

export function useCancelMyAppointment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ id, reason }: { id: string; reason?: string }) => {
      const res = await appointmentApi.cancelMyAppointment(id, reason);
      return res.data;
    },
    onSuccess: (data, variables) => {
      toast.success('Qabul muvaffaqiyatli bekor qilindi');
      queryClient.invalidateQueries({ queryKey: ['myAppointments'] });
      queryClient.invalidateQueries({ queryKey: ['myAppointment', variables.id] });
      queryClient.invalidateQueries({ queryKey: ['publicAvailability'] });
      queryClient.invalidateQueries({ queryKey: ['clinicAppointments'] });
    },
    onError: (err: any) => {
      const msg = err.response?.data?.message || 'Qabulni bekor qilishda xatolik yuz berdi';
      toast.error(msg);
    },
  });
}

export function useClinicAppointments(
  clinicId: string,
  filter?: ClinicAppointmentFilter,
  enabled = true
) {
  return useQuery({
    queryKey: ['clinicAppointments', clinicId, filter],
    queryFn: async () => {
      if (!clinicId) return null;
      const res = await appointmentApi.getClinicAppointments(clinicId, filter);
      return res.data;
    },
    enabled: Boolean(clinicId && enabled),
  });
}

export function useClinicAppointmentDetail(
  clinicId: string,
  appointmentId: string,
  enabled = true
) {
  return useQuery({
    queryKey: ['clinicAppointment', clinicId, appointmentId],
    queryFn: async () => {
      if (!clinicId || !appointmentId) return null;
      const res = await appointmentApi.getClinicAppointmentById(clinicId, appointmentId);
      return res.data;
    },
    enabled: Boolean(clinicId && appointmentId && enabled),
  });
}

export function useCancelClinicAppointment(clinicId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      appointmentId,
      reason,
    }: {
      appointmentId: string;
      reason?: string;
    }) => {
      const res = await appointmentApi.cancelClinicAppointment(clinicId, appointmentId, reason);
      return res.data;
    },
    onSuccess: (data, variables) => {
      toast.success('Qabul muvaffaqiyatli bekor qilindi');
      queryClient.invalidateQueries({ queryKey: ['clinicAppointments', clinicId] });
      queryClient.invalidateQueries({
        queryKey: ['clinicAppointment', clinicId, variables.appointmentId],
      });
      queryClient.invalidateQueries({ queryKey: ['publicAvailability'] });
    },
    onError: (err: any) => {
      const msg = err.response?.data?.message || 'Qabulni bekor qilishda xatolik yuz berdi';
      toast.error(msg);
    },
  });
}

export function useCompleteClinicAppointment(clinicId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (appointmentId: string) => {
      const res = await appointmentApi.completeClinicAppointment(clinicId, appointmentId);
      return res.data;
    },
    onSuccess: (data, appointmentId) => {
      toast.success('Qabul yakunlandi deb belgilandi');
      queryClient.invalidateQueries({ queryKey: ['clinicAppointments', clinicId] });
      queryClient.invalidateQueries({
        queryKey: ['clinicAppointment', clinicId, appointmentId],
      });
    },
    onError: (err: any) => {
      const msg = err.response?.data?.message || 'Amalni bajarishda xatolik yuz berdi';
      toast.error(msg);
    },
  });
}

export function useNoShowClinicAppointment(clinicId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (appointmentId: string) => {
      const res = await appointmentApi.noShowClinicAppointment(clinicId, appointmentId);
      return res.data;
    },
    onSuccess: (data, appointmentId) => {
      toast.success('Bemor kelmagan deb qayd etildi');
      queryClient.invalidateQueries({ queryKey: ['clinicAppointments', clinicId] });
      queryClient.invalidateQueries({
        queryKey: ['clinicAppointment', clinicId, appointmentId],
      });
    },
    onError: (err: any) => {
      const msg = err.response?.data?.message || 'Amalni bajarishda xatolik yuz berdi';
      toast.error(msg);
    },
  });
}
