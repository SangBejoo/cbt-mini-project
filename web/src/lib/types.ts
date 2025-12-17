export interface Materi {
  id: number;
  mataPelajaran: {
    id: number;
    nama: string;
  };
  nama: string;
  tingkatan: number;
}

export interface Soal {
  id: number;
  materi: Materi;
  pertanyaan: string;
  opsiA: string;
  opsiB: string;
  opsiC: string;
  opsiD: string;
  jawabanBenar: JawabanOption;
}

export enum JawabanOption {
  JAWABAN_INVALID = 0,
  A = 1,
  B = 2,
  C = 3,
  D = 4,
}

export interface CreateSoalRequest {
  id_materi: number;
  pertanyaan: string;
  opsiA: string;
  opsiB: string;
  opsiC: string;
  opsiD: string;
  jawaban_benar: JawabanOption;
}

export interface UpdateSoalRequest extends CreateSoalRequest {
  id: number;
}

export interface ListSoalResponse {
  soal: Soal[];
  pagination: {
    page: number;
    limit: number;
    total: number;
  };
}

export interface ListMateriResponse {
  materi: Materi[];
  pagination: {
    page: number;
    limit: number;
    total: number;
  };
}