import { useState, useCallback, useEffect, useMemo } from 'react';
import { useToast } from '@chakra-ui/react';
import apiClient from '../services/api';

export interface TestSession {
  id: number;
  sessionToken: string;
  user: {
    id: number;
    email: string;
    nama: string;
    role: string;
    isActive: boolean;
  } | null;
  namaPeserta: string;
  tingkat: {
    id: number;
    nama: string;
  };
  mataPelajaran: {
    id: number;
    nama: string;
  };
  waktuMulai: string;
  waktuSelesai: string | null;
  batasWaktu: string;
  durasiMenit: number;
  nilaiAkhir: number | null;
  jumlahBenar: number | null;
  totalSoal: number | null;
  status: string;
}

export interface SessionsResponse {
  testSessions: TestSession[];
  pagination: {
    totalCount: number;
    totalPages: number;
    currentPage: number;
    pageSize: number;
  };
}

export interface GroupedSessions {
  [peserta: string]: {
    [subject: string]: {
      [level: string]: TestSession[];
    };
  };
}

export function useSessions() {
  const [sessions, setSessions] = useState<TestSession[]>([]);
  const [loading, setLoading] = useState(true);
  const toast = useToast();

  const fetchSessions = useCallback(async () => {
    setLoading(true);
    try {
      const response = await apiClient.get<SessionsResponse>(
        '/admin/sessions'
      );
      setSessions(response.data.testSessions || []);
    } catch (error: any) {
      const message = error.response?.data?.message || error.message || 'Gagal memuat sesi';
      toast({
        title: 'Error',
        description: message,
        status: 'error',
        duration: 4000,
        isClosable: true,
      });
      setSessions([]);
    } finally {
      setLoading(false);
    }
  }, [toast]);

  useEffect(() => {
    fetchSessions();
  }, [fetchSessions]);

  // Group sessions by peserta > subject > level
  const groupedSessions = useMemo(() => {
    const groups: GroupedSessions = {};
    if (!Array.isArray(sessions)) return groups;

    sessions.forEach((item) => {
      const peserta = item.namaPeserta || 'Unknown';
      const subject = item.mataPelajaran?.nama || 'Unknown';
      const level = item.tingkat?.nama || 'Unknown';

      if (!groups[peserta]) groups[peserta] = {};
      if (!groups[peserta][subject]) groups[peserta][subject] = {};
      if (!groups[peserta][subject][level]) groups[peserta][subject][level] = [];

      groups[peserta][subject][level].push(item);
    });

    return groups;
  }, [sessions]);

  // Get filtered peserta list
  const getPesertas = useCallback(
    (searchQuery: string = '') => {
      const allPesertas = Object.keys(groupedSessions);
      if (!searchQuery.trim()) return allPesertas;
      return allPesertas.filter((p) =>
        p.toLowerCase().includes(searchQuery.toLowerCase())
      );
    },
    [groupedSessions]
  );

  // Get filtered subjects for peserta
  const getSubjects = useCallback(
    (pesertas: string[], selectedSubject: string = 'Semua') => {
      if (selectedSubject === 'Semua') {
        const allSubjects = new Set<string>();
        pesertas.forEach((p) => {
          Object.keys(groupedSessions[p] || {}).forEach((subj) =>
            allSubjects.add(subj)
          );
        });
        return ['Semua', ...Array.from(allSubjects)];
      }
      return ['Semua', selectedSubject];
    },
    [groupedSessions]
  );

  // Get filtered levels
  const getLevels = useCallback(
    (
      pesertas: string[],
      selectedSubject: string = 'Semua',
      selectedLevel: string = 'Semua'
    ) => {
      if (selectedLevel === 'Semua') {
        const allLevels = new Set<string>();
        if (selectedSubject === 'Semua') {
          pesertas.forEach((p) => {
            Object.values(groupedSessions[p] || {}).forEach((subj) =>
              Object.keys(subj).forEach((lvl) => allLevels.add(lvl))
            );
          });
        } else {
          pesertas.forEach((p) => {
            Object.keys(
              groupedSessions[p]?.[selectedSubject] || {}
            ).forEach((lvl) => allLevels.add(lvl));
          });
        }
        return ['Semua', ...Array.from(allLevels)];
      }
      return ['Semua', selectedLevel];
    },
    [groupedSessions]
  );

  // Get filtered groups
  const getFilteredGroups = useCallback(
    (
      pesertas: string[],
      selectedSubject: string = 'Semua',
      selectedLevel: string = 'Semua'
    ): GroupedSessions => {
      const filtered: GroupedSessions = {};

      pesertas.forEach((peserta) => {
        filtered[peserta] = {};
        Object.keys(groupedSessions[peserta]).forEach((subj) => {
          if (selectedSubject !== 'Semua' && subj !== selectedSubject) return;
          filtered[peserta][subj] = {};
          Object.keys(groupedSessions[peserta][subj]).forEach((lvl) => {
            if (selectedLevel !== 'Semua' && lvl !== selectedLevel) return;
            filtered[peserta][subj][lvl] =
              groupedSessions[peserta][subj][lvl];
          });
          if (Object.keys(filtered[peserta][subj]).length === 0)
            delete filtered[peserta][subj];
        });
        if (Object.keys(filtered[peserta]).length === 0)
          delete filtered[peserta];
      });

      return filtered;
    },
    [groupedSessions]
  );

  return {
    sessions,
    loading,
    groupedSessions,
    fetchSessions,
    getPesertas,
    getSubjects,
    getLevels,
    getFilteredGroups,
  };
}
