'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import {
  Box,
  Button,
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
} from '@chakra-ui/react';
import axios from 'axios';

interface HistoryItem {
  id: number;
  sessionToken: string;
  mataPelajaran: {
    id: number;
    nama: string;
  };
  tingkat: {
    id: number;
    nama: string;
  };
  waktuMulai: string;
  waktuSelesai: string;
  durasiPengerjaanDetik: number;
  nilaiAkhir: number;
  jumlahBenar: number;
  totalSoal: number;
  status: string;
  namaMateri?: string;
}

const API_BASE = 'http://localhost:8080/v1/history/student';

export default function HistoryPage() {
  const toast = useToast();
  const router = useRouter();
  const [history, setHistory] = useState<HistoryItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchHistory();
  }, []);

  const fetchHistory = async () => {
    try {
      const response = await axios.get(API_BASE);
      setHistory(response.data.history);
    } catch (error) {
      console.error('Error fetching history:', error);
      toast({ title: 'Error loading history', status: 'error' });
    } finally {
      setLoading(false);
    }
  };

  const formatDateTime = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('id-ID', { day: '2-digit', month: 'long', year: 'numeric' }) + ' - ' + date.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  };

  if (loading) {
    return (
      <Container maxW="container.xl" py={10}>
        <Text>Memuat riwayat...</Text>
      </Container>
    );
  }

  return (
    <Container maxW="container.xl" py={10}>
      <VStack spacing={6} align="stretch">
        <Box bg="blue.50" py={6} px={4} borderRadius="md" textAlign="center">
          <Heading as="h1" size="lg" color="blue.700">
            HISTORI NILAI {history.length > 0 && history[0].mataPelajaran ? `${history[0].mataPelajaran.nama.toUpperCase()} ${history[0].tingkat.nama} SD KELAS ${history[0].tingkat.nama === '1' ? 'IV' : history[0].tingkat.nama}` : ''}
          </Heading>
        </Box>

        <HStack justify="space-between" align="center">
          <Button
            variant="link"
            color="gray.600"
            leftIcon={<Text>←</Text>}
            onClick={() => router.push('/student')}
          >
            Kembali
          </Button>
          <Select maxW="200px" size="sm" placeholder="Pilihan bab">
            <option>Semua</option>
          </Select>
        </HStack>

        {history.length === 0 ? (
          <Card>
            <CardBody>
              <Text textAlign="center">Belum ada riwayat tes tersedia.</Text>
            </CardBody>
          </Card>
        ) : (
          <SimpleGrid columns={{ base: 1, md: 2 }} spacing={6}>
            {history.map((item) => (
              <Card
                key={item.sessionToken}
                bg="orange.50"
                borderWidth="2px"
                borderColor="orange.200"
                borderRadius="xl"
                overflow="hidden"
                _hover={{ shadow: 'lg' }}
                cursor="pointer"
                onClick={() => router.push(`/student/results/${item.sessionToken}`)}
              >
                <CardBody>
                  <VStack spacing={4} align="stretch">
                    <HStack justify="flex-start">
                      <Badge colorScheme="orange" px={3} py={1} borderRadius="md" fontSize="sm">
                        Nilai CBT
                      </Badge>
                    </HStack>

                    <Text fontSize="md" fontWeight="medium" color="gray.700">
                      {item.namaMateri || item.mataPelajaran.nama}
                    </Text>

                    <Box textAlign="center" py={4}>
                      <Text fontSize="5xl" fontWeight="bold" color="orange.500">
                        {item.nilaiAkhir.toFixed(2)}
                      </Text>
                    </Box>

                    <VStack spacing={2} align="stretch" fontSize="sm" color="gray.600">
                      <HStack>
                        <Text>⏰ Mulai :</Text>
                        <Text>{formatDateTime(item.waktuMulai)}</Text>
                      </HStack>
                      <HStack>
                        <Text>⏰ Selesai :</Text>
                        <Text>{formatDateTime(item.waktuSelesai)}</Text>
                      </HStack>
                    </VStack>
                  </VStack>
                </CardBody>
              </Card>
            ))}
          </SimpleGrid>
        )}

        <Link href="/student">
          <Button variant="outline" size="lg" width="full" mt={4}>
            Kembali ke Beranda
          </Button>
        </Link>
      </VStack>
    </Container>
  );
}