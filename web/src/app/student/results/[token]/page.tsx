'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
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
  Stat,
  StatLabel,
  StatNumber,
  StatGroup,
  Badge,
} from '@chakra-ui/react';
import axios from 'axios';

interface TestResultResponse {
  sessionInfo: {
    id: number;
    sessionToken: string;
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
    waktuSelesai: string;
    batasWaktu: string;
    durasiMenit: number;
    nilaiAkhir: number;
    jumlahBenar: number;
    totalSoal: number;
    status: string;
  };
  detailJawaban: Array<{
    nomorUrut: number;
    pertanyaan: string;
    opsiA: string;
    opsiB: string;
    opsiC: string;
    opsiD: string;
    jawabanDipilih: string;
    jawabanBenar: string;
    isCorrect: boolean;
  }>;
  tingkat: Array<{
    id: number;
    nama: string;
  }>;
}

const API_BASE = 'http://localhost:8080/v1/sessions';

export default function ResultsPage() {
  const params = useParams();
  const token = params.token as string;
  const router = useRouter();
  const toast = useToast();

  const [result, setResult] = useState<TestResultResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchResult();
  }, [token]);

  const fetchResult = async () => {
    try {
      const response = await axios.get(`${API_BASE}/${token}/result`);
      setResult(response.data as TestResultResponse);
    } catch (error) {
      console.error('Error fetching result:', error);
      toast({ title: 'Error loading results', status: 'error' });
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <Container maxW="container.md" py={10}>
        <Text>Loading results...</Text>
      </Container>
    );
  }

  if (!result?.sessionInfo) {
    return (
      <Container maxW="container.md" py={10}>
        <Text>Results not available.</Text>
        <Link href="/student">
          <Button mt={4}>Back to Home</Button>
        </Link>
      </Container>
    );
  }

  const sessionInfo = result.sessionInfo;
  const scorePercentage = sessionInfo.nilaiAkhir || 0;
  const isPassed = scorePercentage >= 70; // Assuming 70% pass mark

  return (
    <Container maxW="container.md" py={10}>
      <VStack spacing={6}>
        <Heading as="h1" size="xl" textAlign="center">
          Test Results
        </Heading>

        <Card width="full">
          <CardBody>
            <VStack spacing={6}>
              <Box textAlign="center">
                <Text fontSize="2xl" fontWeight="bold" color={isPassed ? 'green.500' : 'red.500'}>
                  {scorePercentage.toFixed(1)}%
                </Text>
                <Badge colorScheme={isPassed ? 'green' : 'red'} fontSize="md">
                  {isPassed ? 'PASSED' : 'FAILED'}
                </Badge>
              </Box>

              <StatGroup width="full">
                <Stat>
                  <StatLabel>Participant</StatLabel>
                  <StatNumber>{sessionInfo.namaPeserta}</StatNumber>
                </Stat>
                <Stat>
                  <StatLabel>Correct Answers</StatLabel>
                  <StatNumber>{sessionInfo.jumlahBenar}/{sessionInfo.totalSoal}</StatNumber>
                </Stat>
                <Stat>
                  <StatLabel>Subject</StatLabel>
                  <StatNumber>{sessionInfo.mataPelajaran.nama}</StatNumber>
                </Stat>
              </StatGroup>

              <StatGroup width="full">
                <Stat>
                  <StatLabel>Level</StatLabel>
                  <StatNumber>{sessionInfo.tingkat.nama}</StatNumber>
                </Stat>
                <Stat>
                  <StatLabel>Duration</StatLabel>
                  <StatNumber>{sessionInfo.durasiMenit} minutes</StatNumber>
                </Stat>
                <Stat>
                  <StatLabel>Status</StatLabel>
                  <StatNumber>
                    <Badge colorScheme={sessionInfo.status === 'COMPLETED' ? 'green' : 'yellow'}>
                      {sessionInfo.status}
                    </Badge>
                  </StatNumber>
                </Stat>
              </StatGroup>

              <Box width="full">
                <Text fontWeight="medium" mb={2}>Time Information:</Text>
                <Text>Started: {new Date(sessionInfo.waktuMulai).toLocaleString()}</Text>
                <Text>Completed: {new Date(sessionInfo.waktuSelesai).toLocaleString()}</Text>
              </Box>
            </VStack>
          </CardBody>
        </Card>

        <VStack spacing={4}>
          <Link href="/student/history">
            <Button colorScheme="blue" size="lg">
              View My History
            </Button>
          </Link>
          <Link href="/student">
            <Button variant="outline" size="lg">
              Back to Home
            </Button>
          </Link>
        </VStack>
      </VStack>
    </Container>
  );
}