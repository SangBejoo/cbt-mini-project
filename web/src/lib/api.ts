import { ListSoalResponse, Soal, CreateSoalRequest, UpdateSoalRequest, ListMateriResponse } from './types';

const API_BASE = 'http://localhost:8080/v1';

export async function fetchSoalList(page = 1, limit = 10): Promise<ListSoalResponse> {
  const res = await fetch(`${API_BASE}/questions?page=${page}&limit=${limit}`);
  if (!res.ok) throw new Error('Failed to fetch soal');
  return res.json();
}

export async function fetchSoal(id: number): Promise<Soal> {
  const res = await fetch(`${API_BASE}/questions/${id}`);
  if (!res.ok) throw new Error('Failed to fetch soal');
  const data = await res.json();
  return data.soal;
}

export async function createSoal(data: CreateSoalRequest): Promise<Soal> {
  const payload = {
    id_materi: data.id_materi,
    pertanyaan: data.pertanyaan,
    opsi_a: data.opsiA,
    opsi_b: data.opsiB,
    opsi_c: data.opsiC,
    opsi_d: data.opsiD,
    jawaban_benar: data.jawaban_benar,
  };
  const res = await fetch(`${API_BASE}/questions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  if (!res.ok) throw new Error('Failed to create soal');
  const result = await res.json();
  return result.soal;
}

export async function updateSoal(data: UpdateSoalRequest): Promise<Soal> {
  const payload = {
    id: data.id,
    id_materi: data.id_materi,
    pertanyaan: data.pertanyaan,
    opsi_a: data.opsiA,
    opsi_b: data.opsiB,
    opsi_c: data.opsiC,
    opsi_d: data.opsiD,
    jawaban_benar: data.jawaban_benar,
  };
  const res = await fetch(`${API_BASE}/questions/${data.id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  if (!res.ok) throw new Error('Failed to update soal');
  const result = await res.json();
  return result.soal;
}

export async function deleteSoal(id: number): Promise<void> {
  const res = await fetch(`${API_BASE}/questions/${id}`, {
    method: 'DELETE',
  });
  if (!res.ok) throw new Error('Failed to delete soal');
}

export async function fetchMateriList(page = 1, limit = 100): Promise<ListMateriResponse> {
  const res = await fetch(`${API_BASE}/topics?page=${page}&limit=${limit}`);
  if (!res.ok) throw new Error('Failed to fetch materi');
  return res.json();
}