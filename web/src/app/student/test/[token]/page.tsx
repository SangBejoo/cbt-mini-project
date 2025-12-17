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
}

interface TestSessionData {
  session_token: string;
  soal: Question;
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
  const [currentQuestion, setCurrentQuestion] = useState(1); // Start with question 1
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!token) {
      toast({ title: 'Invalid session token', status: 'error' });
      router.push('/student');
      return;
    }
    fetchQuestion(currentQuestion);
  }, [token, currentQuestion]);

  const fetchQuestion = async (nomorUrut: number) => {
    try {
      const response = await axios.get(`${API_BASE}/${token}/questions?nomor_urut=${nomorUrut}`);
      const data = response.data;
      setSessionData(data);
      // Set answer if already answered
      if (data.soal && data.soal.jawabanDipilih) {
        setAnswers(prev => ({ ...prev, [data.soal.nomorUrut]: data.soal.jawabanDipilih }));
      }
    } catch (error) {
      console.error('Error fetching question:', error);
      toast({ title: 'Error loading question', status: 'error' });
    } finally {
      setLoading(false);
    }
  };

  const handleAnswerChange = (questionId: number, answer: string) => {
    setAnswers({ ...answers, [questionId]: answer });
  };

  const submitAnswer = async (questionId: number, answer: string) => {
    try {
      await axios.post(`${API_BASE}/${token}/answers`, {
        nomor_urut: questionId,
        jawaban_dipilih: answer,
      });
    } catch (error) {
      console.error('Error submitting answer:', error);
    }
  };

  const handleNext = async () => {
    if (sessionData?.soal && answers[sessionData.soal.nomorUrut]) {
      await submitAnswer(sessionData.soal.nomorUrut, answers[sessionData.soal.nomorUrut]);
    }

    if (currentQuestion < (sessionData?.total_soal || 1)) {
      setCurrentQuestion(currentQuestion + 1);
    }
  };

  const handlePrev = () => {
    if (currentQuestion > 1) {
      setCurrentQuestion(currentQuestion - 1);
    }
  };

  const handleFinish = async () => {
    // Submit current answer if any
    if (sessionData?.soal && answers[sessionData.soal.nomorUrut]) {
      await submitAnswer(sessionData.soal.nomorUrut, answers[sessionData.soal.nomorUrut]);
    }

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

  if (!sessionData?.soal) {
    return (
      <Container maxW="container.md" py={10}>
        <Text>No question available for this test.</Text>
        <Button onClick={() => router.push('/student')} mt={4}>
          Back to Home
        </Button>
      </Container>
    );
  }

  const currentQ = sessionData.soal;
  const progress = ((currentQuestion) / (sessionData.total_soal || 1)) * 100;

  return (
    <Container maxW="container.lg" py={10}>
      <VStack spacing={6}>
        <HStack width="full" justify="space-between">
          <Heading size="lg">Question {currentQuestion} of {sessionData.total_soal}</Heading>
          <Badge colorScheme="blue" fontSize="md">
            Answered: {sessionData.dijawab_count}/{sessionData.total_soal}
          </Badge>
        </HStack>

        <Progress value={progress} width="full" colorScheme="blue" />

        <Card width="full">
          <CardBody>
            <VStack spacing={6} align="stretch">
              <Text fontSize="lg" fontWeight="medium">
                {currentQ.pertanyaan}
              </Text>

              <RadioGroup
                value={answers[currentQ.nomorUrut] || ''}
                onChange={(value) => handleAnswerChange(currentQ.nomorUrut, value)}
              >
                <VStack spacing={3} align="stretch">
                  <Radio value="A">{currentQ.opsiA}</Radio>
                  <Radio value="B">{currentQ.opsiB}</Radio>
                  <Radio value="C">{currentQ.opsiC}</Radio>
                  <Radio value="D">{currentQ.opsiD}</Radio>
                </VStack>
              </RadioGroup>
            </VStack>
          </CardBody>
        </Card>

        <HStack spacing={4} width="full" justify="space-between">
          <Button
            onClick={handlePrev}
            isDisabled={currentQuestion === 1}
            variant="outline"
          >
            Previous
          </Button>

          {currentQuestion < (sessionData.total_soal || 1) ? (
            <Button onClick={handleNext} colorScheme="blue">
              Next
            </Button>
          ) : (
            <Button
              onClick={handleFinish}
              colorScheme="green"
              isLoading={submitting}
              loadingText="Completing test..."
            >
              Finish Test
            </Button>
          )}
        </HStack>
      </VStack>
    </Container>
  );
}