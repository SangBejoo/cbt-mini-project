'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
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
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Badge,
  HStack,
} from '@chakra-ui/react';
import axios from 'axios';

interface HistoryItem {
  id: number;
  session_token: string;
  mata_pelajaran: {
    id: number;
    nama: string;
  };
  tingkat: {
    id: number;
    nama: string;
  };
  waktu_mulai: string;
  waktu_selesai: string;
  durasi_pengerjaan_detik: number;
  nilai_akhir: number;
  jumlah_benar: number;
  total_soal: number;
  status: string;
}

const API_BASE = 'http://localhost:8080/v1/history/student';

export default function HistoryPage() {
  const toast = useToast();
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

  if (loading) {
    return (
      <Container maxW="container.xl" py={10}>
        <Text>Loading history...</Text>
      </Container>
    );
  }

  return (
    <Container maxW="container.xl" py={10}>
      <VStack spacing={6}>
        <Heading as="h1" size="xl" textAlign="center">
          Test History
        </Heading>

        {history.length === 0 ? (
          <Card width="full">
            <CardBody>
              <Text textAlign="center">No test history available.</Text>
            </CardBody>
          </Card>
        ) : (
          <Card width="full">
            <CardBody>
              <Table variant="simple">
                <Thead>
                  <Tr>
                    <Th>Subject</Th>
                    <Th>Level</Th>
                    <Th>Topic</Th>
                    <Th>Score</Th>
                    <Th>Correct</Th>
                    <Th>Status</Th>
                    <Th>Completed</Th>
                    <Th>Actions</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {history.map((item) => (
                    <Tr key={item.sessionToken}>
                      <Td>{item.mataPelajaran?.nama || '-'}</Td>
                      <Td>{item.tingkat?.nama || '-'}</Td>
                      <Td>-</Td>
                      <Td>
                        <HStack>
                          <Text fontWeight="bold" color={(item.nilaiAkhir || 0) >= 70 ? 'green.500' : 'red.500'}>
                            {(item.nilaiAkhir || 0).toFixed(1)}%
                          </Text>
                          <Badge colorScheme={(item.nilaiAkhir || 0) >= 70 ? 'green' : 'red'}>
                            {(item.nilaiAkhir || 0) >= 70 ? 'PASS' : 'FAIL'}
                          </Badge>
                        </HStack>
                      </Td>
                      <Td>{(item.jumlahBenar || 0)}/{(item.totalSoal || 0)}</Td>
                      <Td>
                        <Badge colorScheme={item.status === 'completed' ? 'green' : 'yellow'}>
                          {item.status}
                        </Badge>
                      </Td>
                      <Td>{item.waktuSelesai ? new Date(item.waktuSelesai).toLocaleDateString() : '-'}</Td>
                      <Td>
                        <Link href={`/student/results/${item.sessionToken}`}>
                          <Button size="sm" colorScheme="blue">
                            View Details
                          </Button>
                        </Link>
                      </Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            </CardBody>
          </Card>
        )}

        <Link href="/student">
          <Button variant="outline" size="lg">
            Back to Home
          </Button>
        </Link>
      </VStack>
    </Container>
  );
}