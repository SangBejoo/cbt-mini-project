'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import {
  Box,
  Button,
  RadioGroup,
  Radio,
  VStack,
  Heading,
  Container,
  useToast,
  Card,
  CardBody,
  Text,
  Progress,
  HStack,
  Badge,
} from '@chakra-ui/react';
import axios from 'axios';

interface Question {
  id: number;
  pertanyaan: string;
  opsiA: string;
  opsiB: string;
  opsiC: string;
  opsiD: string;
  nomorUrut: number;
  jawabanDipilih?: string;
  materi: {
    nama: string;
    mataPelajaran: {
      nama: string;
    };
    tingkat: {
      nama: string;
    };
  };
  image_path?: string;
}

interface TestSessionData {
  session_token: string;
  soal: Question[];
  total_soal: number;
  current_nomor_urut: number;
  dijawab_count: number;
  is_answered_status: boolean[];
  batas_waktu: string;
}

const API_BASE = 'http://localhost:8080/v1/sessions';

export default function TestPage() {
  const params = useParams();
  const token = params.token as string;
  const router = useRouter();
  const toast = useToast();

  const [sessionData, setSessionData] = useState<TestSessionData | null>(null);
  const [answers, setAnswers] = useState<Record<number, string>>({});
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!token) {
      toast({ title: 'Invalid session token', status: 'error' });
      router.push('/student');
      return;
    }
    fetchAllQuestions();
  }, [token]);

  const fetchAllQuestions = async () => {
    try {
      const response = await axios.get(`${API_BASE}/${token}/questions`);
      const data = response.data;
      setSessionData(data);
      // Set answers for all questions
      const initialAnswers: Record<number, string> = {};
      data.soal.forEach((q: Question) => {
        if (q.jawabanDipilih) {
          initialAnswers[q.nomorUrut] = q.jawabanDipilih;
        }
      });
      setAnswers(initialAnswers);
    } catch (error) {
      console.error('Error fetching questions:', error);
      toast({ title: 'Error loading questions', status: 'error' });
    } finally {
      setLoading(false);
    }
  };

  const handleAnswerChange = async (questionId: number, answer: string) => {
    setAnswers({ ...answers, [questionId]: answer });
    // Submit answer immediately
    try {
      await axios.post(`${API_BASE}/${token}/answers`, {
        nomor_urut: questionId,
        jawaban_dipilih: answer,
      });
      toast({ title: 'Answer saved', status: 'success', duration: 1000 });
      // Refresh data to update answered count
      fetchAllQuestions();
    } catch (error) {
      console.error('Error submitting answer:', error);
      toast({ title: 'Error saving answer', status: 'error' });
    }
  };

  const handleFinish = async () => {
    // Complete session
    setSubmitting(true);
    try {
      await axios.post(`${API_BASE}/${token}/complete`);
      toast({ title: 'Test completed!', status: 'success' });
      router.push(`/student/results/${token}`);
    } catch (error) {
      console.error('Error completing test:', error);
      toast({ title: 'Error completing test', status: 'error' });
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <Container maxW="container.md" py={10}>
        <Text>Loading question...</Text>
      </Container>
    );
  }

  if (!sessionData?.soal || sessionData.soal.length === 0) {
    return (
      <Container maxW="container.md" py={10}>
        <Text>No questions available for this test.</Text>
        <Button onClick={() => router.push('/student')} mt={4}>
          Back to Home
        </Button>
      </Container>
    );
  }

  const progress = (sessionData.dijawab_count / sessionData.total_soal) * 100;

  return (
    <Container maxW="container.xl" py={10}>
      <VStack spacing={6}>
        <HStack width="full" justify="space-between">
          <Heading size="lg">Test Session</Heading>
          <Badge colorScheme="blue" fontSize="md">
            Answered: {sessionData.dijawab_count}/{sessionData.total_soal}
          </Badge>
        </HStack>

        <Progress value={progress} width="full" colorScheme="blue" />

        <VStack spacing={6} width="full" align="stretch">
          {sessionData.soal.map((question, index) => (
            <Card key={question.id} width="full">
              <CardBody>
                <VStack spacing={4} align="stretch">
                  <HStack justify="space-between">
                    <Heading size="md">Question {question.nomorUrut}</Heading>
                    <Badge colorScheme={answers[question.nomorUrut] ? 'green' : 'gray'}>
                      {answers[question.nomorUrut] ? 'Answered' : 'Not Answered'}
                    </Badge>
                  </HStack>

                  <Text fontSize="sm" color="gray.600">
                    {question.materi.mataPelajaran.nama} - {question.materi.nama} ({question.materi.tingkat.nama})
                  </Text>

                  <Text fontSize="lg" fontWeight="medium">
                    {question.pertanyaan}
                  </Text>

                  {question.image_path && (
                    <Box>
                      <img
                        src={`http://localhost:8080/${question.image_path}`}
                        alt="Question"
                        style={{ maxWidth: '100%', maxHeight: '300px', objectFit: 'contain' }}
                      />
                    </Box>
                  )}

                  <RadioGroup
                    value={answers[question.nomorUrut] || ''}
                    onChange={(value) => handleAnswerChange(question.nomorUrut, value)}
                  >
                    <VStack spacing={3} align="stretch">
                      <Radio value="A">{question.opsiA}</Radio>
                      <Radio value="B">{question.opsiB}</Radio>
                      <Radio value="C">{question.opsiC}</Radio>
                      <Radio value="D">{question.opsiD}</Radio>
                    </VStack>
                  </RadioGroup>
                </VStack>
              </CardBody>
            </Card>
          ))}
        </VStack>

        <HStack spacing={4} width="full" justify="center">
          <Button
            onClick={handleFinish}
            colorScheme="green"
            size="lg"
            isLoading={submitting}
            loadingText="Completing test..."
          >
            Finish Test
          </Button>
        </HStack>
      </VStack>
    </Container>
  );
}