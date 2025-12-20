'use client';

import { useState, useEffect, useMemo } from 'react';
import {
  Box,
  VStack,
  Heading,
  Container,
  useToast,
  Card,
  CardBody,
  Text,
  Badge,
  HStack,
  SimpleGrid,
  Select,
  Input,
  Accordion,
  AccordionItem,
  AccordionButton,
  AccordionPanel,
  AccordionIcon,
  Spinner,
} from '@chakra-ui/react';
import { useAuth } from '../../auth-context';

interface TestSession {
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

interface ApiResponse {
  testSessions: TestSession[];
  pagination: {
    totalCount: number;
    totalPages: number;
    currentPage: number;
    pageSize: number;
  };
}

const API_BASE = process.env.NEXT_PUBLIC_API_BASE + '/v1/admin/sessions';

export default function HistoryTab() {
  const toast = useToast();
  const { token } = useAuth();
  const [sessions, setSessions] = useState<TestSession[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchPeserta, setSearchPeserta] = useState<string>('');
  const [selectedSubject, setSelectedSubject] = useState<string>('Semua');
  const [selectedLevel, setSelectedLevel] = useState<string>('Semua');

  useEffect(() => {
    fetchSessions();
  }, []);

  const fetchSessions = async () => {
    try {
      const response = await fetch(API_BASE, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      const data: ApiResponse = await response.json();
      setSessions(data.testSessions || []);
    } catch (error) {
      console.error('Error fetching sessions:', error);
      toast({ title: 'Error loading sessions', status: 'error' });
      setSessions([]);
    } finally {
      setLoading(false);
    }
  };

  const groupedSessions = useMemo(() => {
    const groups: Record<string, Record<string, Record<string, TestSession[]>>> = {};
    if (!Array.isArray(sessions)) return groups;
    sessions.forEach(item => {
      const peserta = item.namaPeserta || 'Unknown';
      const subj = item.mataPelajaran?.nama || 'Unknown';
      const lvl = item.tingkat?.nama || 'Unknown';
      if (!groups[peserta]) groups[peserta] = {};
      if (!groups[peserta][subj]) groups[peserta][subj] = {};
      if (!groups[peserta][subj][lvl]) groups[peserta][subj][lvl] = [];
      groups[peserta][subj][lvl].push(item);
    });
    return groups;
  }, [sessions]);

  const pesertas = useMemo(() => {
    const allPesertas = Object.keys(groupedSessions);
    if (searchPeserta.trim() === '') return allPesertas;
    return allPesertas.filter(p => p.toLowerCase().includes(searchPeserta.toLowerCase()));
  }, [groupedSessions, searchPeserta]);

  const subjects = useMemo(() => {
    if (searchPeserta.trim() === '') {
      const allSubjects = new Set<string>();
      Object.values(groupedSessions).forEach(peserta => Object.keys(peserta).forEach(subj => allSubjects.add(subj)));
      return ['Semua', ...Array.from(allSubjects)];
    } else {
      const matchedPesertas = pesertas;
      const allSubjects = new Set<string>();
      matchedPesertas.forEach(p => {
        Object.keys(groupedSessions[p] || {}).forEach(subj => allSubjects.add(subj));
      });
      return ['Semua', ...Array.from(allSubjects)];
    }
  }, [groupedSessions, searchPeserta, pesertas]);

  const levels = useMemo(() => {
    if (selectedSubject === 'Semua') {
      const allLevels = new Set<string>();
      pesertas.forEach(p => {
        Object.values(groupedSessions[p] || {}).forEach(subj =>
          Object.keys(subj).forEach(lvl => allLevels.add(lvl))
        );
      });
      return ['Semua', ...Array.from(allLevels)];
    } else {
      const allLevels = new Set<string>();
      pesertas.forEach(p => {
        Object.keys(groupedSessions[p]?.[selectedSubject] || {}).forEach(lvl => allLevels.add(lvl));
      });
      return ['Semua', ...Array.from(allLevels)];
    }
  }, [groupedSessions, selectedSubject, pesertas]);

  const filteredGroups = useMemo(() => {
    const filtered: Record<string, Record<string, Record<string, TestSession[]>>> = {};
    pesertas.forEach(peserta => {
      filtered[peserta] = {};
      Object.keys(groupedSessions[peserta]).forEach(subj => {
        if (selectedSubject !== 'Semua' && subj !== selectedSubject) return;
        filtered[peserta][subj] = {};
        Object.keys(groupedSessions[peserta][subj]).forEach(lvl => {
          if (selectedLevel !== 'Semua' && lvl !== selectedLevel) return;
          filtered[peserta][subj][lvl] = groupedSessions[peserta][subj][lvl];
        });
        if (Object.keys(filtered[peserta][subj]).length === 0) delete filtered[peserta][subj];
      });
      if (Object.keys(filtered[peserta]).length === 0) delete filtered[peserta];
    });
    return filtered;
  }, [groupedSessions, pesertas, selectedSubject, selectedLevel]);

  const formatDateTime = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('id-ID', { day: '2-digit', month: 'long', year: 'numeric' }) + ' - ' + date.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  };

  const formatDuration = (waktuMulai: string, waktuSelesai: string | null) => {
    if (!waktuSelesai) return 'Belum selesai';
    const start = new Date(waktuMulai);
    const end = new Date(waktuSelesai);
    const diffMs = end.getTime() - start.getTime();
    const diffSec = Math.floor(diffMs / 1000);
    const hours = Math.floor(diffSec / 3600);
    const minutes = Math.floor((diffSec % 3600) / 60);
    const secs = diffSec % 60;
    if (hours > 0) return `${hours}h ${minutes}m ${secs}s`;
    if (minutes > 0) return `${minutes}m ${secs}s`;
    return `${secs}s`;
  };

  if (loading) {
    return (
      <Container maxW="container.xl" py={10}>
        <Box textAlign="center">
          <Spinner size="xl" color="blue.500" mb={4} />
          <Text>Memuat riwayat sesi...</Text>
        </Box>
      </Container>
    );
  }

  return (
    <Container maxW="container.xl" py={10}>
      <VStack spacing={6} align="stretch">
        <Box bg="blue.50" py={6} px={4} borderRadius="md" textAlign="center">
          <Heading as="h1" size="lg" color="blue.700">
            RIWAYAT SESI SISWA
          </Heading>
        </Box>

        <HStack justify="space-between" align="center" spacing={4} flexWrap="wrap">
          <HStack spacing={3} flex={1} maxW="600px">
            <Input
              size="sm"
              placeholder="Cari nama peserta..."
              value={searchPeserta}
              onChange={(e) => setSearchPeserta(e.target.value)}
              borderColor="gray.300"
              _focus={{ borderColor: 'blue.500', boxShadow: '0 0 0 1px #3182ce' }}
            />
            <Select
              size="sm"
              value={selectedSubject}
              onChange={(e) => setSelectedSubject(e.target.value)}
              placeholder="Pilih Mata Pelajaran"
            >
              {subjects.map(subj => (
                <option key={subj} value={subj}>{subj === 'Semua' ? 'Semua Mata Pelajaran' : subj}</option>
              ))}
            </Select>
            <Select
              size="sm"
              value={selectedLevel}
              onChange={(e) => setSelectedLevel(e.target.value)}
              placeholder="Pilih Tingkat"
            >
              {levels.map(lvl => (
                <option key={lvl} value={lvl}>{lvl === 'Semua' ? 'Semua Tingkat' : `Tingkat ${lvl}`}</option>
              ))}
            </Select>
          </HStack>
        </HStack>

        {Object.keys(filteredGroups).length === 0 ? (
          <Card>
            <CardBody>
              <Text textAlign="center">Belum ada riwayat sesi tersedia untuk filter yang dipilih.</Text>
            </CardBody>
          </Card>
        ) : (
          <Accordion allowMultiple>
            {Object.keys(filteredGroups).map(peserta => (
              <AccordionItem key={peserta}>
                <AccordionButton>
                  <Box flex="1" textAlign="left" fontWeight="bold" fontSize="md">
                    Peserta: {peserta}
                  </Box>
                  <AccordionIcon />
                </AccordionButton>
                <AccordionPanel pb={4}>
                  <Accordion allowMultiple>
                    {Object.keys(filteredGroups[peserta]).map(subj => (
                      <AccordionItem key={subj} ml={4}>
                        <AccordionButton>
                          <Box flex="1" textAlign="left" fontWeight="bold" fontSize="sm">
                            Mata Pelajaran: {subj}
                          </Box>
                          <AccordionIcon />
                        </AccordionButton>
                        <AccordionPanel pb={4}>
                          {Object.keys(filteredGroups[peserta][subj]).map(lvl => (
                            <Box key={lvl} mb={6}>
                              <Heading size="sm" mb={4} color="gray.700">
                                Tingkat {lvl}
                              </Heading>
                              <SimpleGrid columns={{ base: 1, md: 2 }} spacing={4}>
                                {filteredGroups[peserta][subj][lvl].map((item) => (
                                  <Card
                                    key={item.sessionToken}
                                    bg="orange.50"
                                    borderWidth="2px"
                                    borderColor="orange.200"
                                    borderRadius="xl"
                                    overflow="hidden"
                                    _hover={{ shadow: 'lg' }}
                                  >
                                    <CardBody>
                                      <VStack spacing={4} align="stretch">
                                        <HStack justify="space-between">
                                          <Badge colorScheme="orange" px={3} py={1} borderRadius="md" fontSize="xs">
                                            Nilai CBT
                                          </Badge>
                                          <Text fontSize="xs" color="gray.500">
                                            {item.totalSoal ? `${item.totalSoal} soal` : 'No Questions'}
                                          </Text>
                                        </HStack>

                                        <Box textAlign="center" py={4}>
                                          <Text fontSize="5xl" fontWeight="bold" color="orange.500">
                                            {item.nilaiAkhir ? item.nilaiAkhir.toFixed(2) : '0.00'}
                                          </Text>
                                          <Text fontSize="xs" color="gray.600">
                                            {item.jumlahBenar || 0}/{item.totalSoal || 0} benar
                                          </Text>
                                        </Box>

                                        <VStack spacing={2} align="stretch" fontSize="xs" color="gray.600">
                                          <HStack>
                                            <Text>Mulai:</Text>
                                            <Text fontSize="xs">{formatDateTime(item.waktuMulai)}</Text>
                                          </HStack>
                                          <HStack>
                                            <Text>Selesai:</Text>
                                            <Text fontSize="xs">{item.waktuSelesai ? formatDateTime(item.waktuSelesai) : 'Belum selesai'}</Text>
                                          </HStack>
                                          <HStack>
                                            <Text>Durasi:</Text>
                                            <Text fontWeight="medium" fontSize="xs">{formatDuration(item.waktuMulai, item.waktuSelesai)}</Text>
                                          </HStack>
                                          <HStack>
                                            <Text>Status:</Text>
                                            <Badge colorScheme={item.status === 'COMPLETED' ? 'green' : 'yellow'} fontSize="xs">
                                              {item.status}
                                            </Badge>
                                          </HStack>
                                        </VStack>
                                      </VStack>
                                    </CardBody>
                                  </Card>
                                ))}
                              </SimpleGrid>
                            </Box>
                          ))}
                        </AccordionPanel>
                      </AccordionItem>
                    ))}
                  </Accordion>
                </AccordionPanel>
              </AccordionItem>
            ))}
          </Accordion>
        )}
      </VStack>
    </Container>
  );
}