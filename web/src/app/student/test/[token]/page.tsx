'use client';

import { useState, useEffect, useRef } from 'react';
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
  HStack,
  Badge,
  SimpleGrid,
  Flex,
  Image,
  useDisclosure,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  ModalCloseButton,
  Stat,
  StatLabel,
  StatNumber,
  StatGroup,
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
  gambar?: Array<{
    id: number;
    namaFile: string;
    filePath: string;
    fileSize: number;
    mimeType: string;
    urutan: number;
    keterangan?: string;
    createdAt: string;
  }>;
}

interface TestSessionData {
  session_token: string;
  soal: Question[];
  total_soal: number;
  current_nomor_urut: number;
  dijawab_count: number;
  is_answered_status: boolean[];
  batas_waktu?: string;
  batasWaktu?: string;
  BatasWaktu?: string;
  durasi_menit?: number;
  waktu_mulai?: string;
}

const API_BASE = process.env.NEXT_PUBLIC_API_BASE + '/v1/sessions';

export default function TestPage() {
  const params = useParams();
  const token = params.token as string;
  const router = useRouter();
  const toast = useToast();
  const hasFetchedRef = useRef(false);

  const [sessionData, setSessionData] = useState<TestSessionData | null>(null);
  const [answers, setAnswers] = useState<Record<number, string>>({});
  const [currentQuestionIndex, setCurrentQuestionIndex] = useState(0);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [mounted, setMounted] = useState(false);
  const { isOpen, onOpen, onClose } = useDisclosure();
  const [showConfirmModal, setShowConfirmModal] = useState(false);
  const [timeRemaining, setTimeRemaining] = useState<number | null>(null);
  const [isTimeUp, setIsTimeUp] = useState(false);
  const [isAutoSubmitting, setIsAutoSubmitting] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  // Load state from localStorage on mount
  useEffect(() => {
    if (mounted && token) {
      const savedState = localStorage.getItem(`test_session_${token}`);
      if (savedState) {
        try {
          const parsed = JSON.parse(savedState);
          setCurrentQuestionIndex(parsed.currentQuestionIndex || 0);
          setAnswers(parsed.answers || {});
        } catch (error) {
          console.error('Error loading saved state:', error);
        }
      }
    }
  }, [mounted, token]);

  // Save state to localStorage when answers or currentQuestionIndex changes
  useEffect(() => {
    if (mounted && token && sessionData) {
      localStorage.setItem(`test_session_${token}`, JSON.stringify({
        currentQuestionIndex,
        answers,
      }));
    }
  }, [mounted, token, currentQuestionIndex, answers, sessionData]);

  useEffect(() => {
    if (!token) {
      toast({ title: 'Invalid session token', status: 'error' });
      router.push('/student');
      return;
    }
    // Only fetch once to prevent double calls
    if (!hasFetchedRef.current) {
      hasFetchedRef.current = true;
      fetchAllQuestions();
    }
  }, [token]);

  // Countdown timer effect - simplified
  useEffect(() => {
    // Check multiple possible field names for batas_waktu (protobuf JSON naming variations)
    const batasWaktuValue = sessionData?.batas_waktu || sessionData?.batasWaktu || sessionData?.BatasWaktu;
    if (!batasWaktuValue) {
      return;
    }

    const timer = setInterval(() => {
      const now = new Date().getTime();
      const batasWaktu = new Date(batasWaktuValue).getTime();
      const remaining = Math.max(0, batasWaktu - now);

      setTimeRemaining(remaining);

      // Debug: log when time is running low
      if (remaining <= 10000 && remaining > 0) { // Last 10 seconds
      }

      if (remaining === 0 && !isAutoSubmitting && !submitting) {
        setIsTimeUp(true);
        setIsAutoSubmitting(true);
        clearInterval(timer);
        toast({ title: 'Waktu telah habis! Otomatis menyimpan...', status: 'warning' });
        setTimeout(() => {
          confirmFinish(true);
        }, 1000);
      }
    }, 1000);

    return () => clearInterval(timer);
  }, [sessionData?.batas_waktu, sessionData?.batasWaktu, sessionData?.BatasWaktu, submitting, isAutoSubmitting]);

  // Format remaining time for display
  const formatTimeRemaining = (ms: number | null) => {
    if (ms === null || ms === 0) return '00:00:00';
    const totalSeconds = Math.floor(ms / 1000);
    const hours = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    const seconds = totalSeconds % 60;
    return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
  };

  // Get timer color based on remaining time
  const getTimerColor = () => {
    if (!timeRemaining) return 'red';
    const totalSeconds = Math.floor(timeRemaining / 1000);
    if (totalSeconds <= 60) return 'red';
    if (totalSeconds <= 300) return 'orange'; // 5 minutes
    return 'green';
  };

  const fetchAllQuestions = async () => {
    try {
      // Ensure auth token is set before making request
      const authToken = localStorage.getItem('auth_token');
      if (authToken && !axios.defaults.headers.common['Authorization']) {
        axios.defaults.headers.common['Authorization'] = `Bearer ${authToken}`;
      }
      
      const response = await axios.get(`${API_BASE}/${token}/questions`);
      const data = response.data;
      setSessionData(data);
      
      // Load saved state from localStorage first
      const savedState = localStorage.getItem(`test_session_${token}`);
      let savedAnswers: Record<number, string> = {};
      let savedIndex = 0;
      
      if (savedState) {
        try {
          const parsed = JSON.parse(savedState);
          savedAnswers = parsed.answers || {};
          savedIndex = parsed.currentQuestionIndex || 0;
        } catch (e) {
          console.error('Error parsing saved state:', e);
        }
      }
      
      // Merge server answers with saved answers (prefer saved answers)
      const mergedAnswers: Record<number, string> = { ...savedAnswers };
      data.soal.forEach((q: Question) => {
        if (q.jawabanDipilih && q.jawabanDipilih !== 'JAWABAN_INVALID' && !mergedAnswers[q.nomorUrut]) {
          mergedAnswers[q.nomorUrut] = q.jawabanDipilih;
        }
      });
      
      setAnswers(mergedAnswers);
      setCurrentQuestionIndex(savedIndex);
      
      // Save merged state
      localStorage.setItem(`test_session_${token}`, JSON.stringify({
        currentQuestionIndex: savedIndex,
        answers: mergedAnswers,
      }));
    } catch (error) {
      console.error('Error fetching questions:', error);
      toast({ title: 'Error loading questions', status: 'error' });
    } finally {
      setLoading(false);
    }
  };

  const handleAnswerChange = async (questionId: number, answer: string) => {
    setAnswers({ ...answers, [questionId]: answer });
    // Submit answer immediately without refetching all
    try {
      // Ensure auth token is set before making request
      const authToken = localStorage.getItem('auth_token');
      if (authToken && !axios.defaults.headers.common['Authorization']) {
        axios.defaults.headers.common['Authorization'] = `Bearer ${authToken}`;
      }
      
      await axios.post(`${API_BASE}/${token}/answers`, {
        nomor_urut: questionId,
        jawaban_dipilih: answer,
      });
      // Update dijawab_count locally
      if (sessionData) {
        setSessionData({
          ...sessionData,
          dijawab_count: Object.keys({ ...answers, [questionId]: answer }).length,
        });
      }
    } catch (error) {
      console.error('Error submitting answer:', error);
      toast({ title: 'Error menyimpan jawaban', status: 'error' });
    }
  };

  const handleFinish = () => {
    setShowConfirmModal(true);
  };

  const handleConfirmFinish = () => {
    setShowConfirmModal(false);
    confirmFinish();
  };

  const handleCancelFinish = () => {
    setShowConfirmModal(false);
  };

  const confirmFinish = async (isAutoSubmit = false) => {
    setSubmitting(true);
    try {
      // Ensure auth token is set before making request
      const authToken = localStorage.getItem('auth_token');
      if (authToken && !axios.defaults.headers.common['Authorization']) {
        axios.defaults.headers.common['Authorization'] = `Bearer ${authToken}`;
      }
      
      await axios.post(`${API_BASE}/${token}/complete`);
      if (!isAutoSubmit) {
        toast({ title: 'Tes selesai!', status: 'success' });
      }
      // Clear localStorage after completing test
      localStorage.removeItem(`test_session_${token}`);
      if (!isAutoSubmit) {
        toast({ title: 'Tes selesai!', status: 'success' });
      }
      // Clear localStorage after completing test
      localStorage.removeItem(`test_session_${token}`);
      // Small delay before redirect
      setTimeout(() => {
        router.push(`/student/results/${token}`);
      }, 500);
    } catch (error) {
      console.error('Error completing test:', error);
      if (!isAutoSubmit) {
        toast({ title: 'Error menyelesaikan tes', status: 'error' });
      } else {
        // For auto-submit, try again after a delay
        setTimeout(() => {
          confirmFinish(true);
        }, 2000);
      }
    } finally {
      setSubmitting(false);
      onClose();
    }
  };

  const goToQuestion = (index: number) => {
    setCurrentQuestionIndex(index);
  };

  const goToNextQuestion = () => {
    if (currentQuestionIndex < sessionData!.soal.length - 1) {
      setCurrentQuestionIndex(currentQuestionIndex + 1);
    }
  };

  const goToPreviousQuestion = () => {
    if (currentQuestionIndex > 0) {
      setCurrentQuestionIndex(currentQuestionIndex - 1);
    }
  };

  if (!mounted) {
    return (
      <Container maxW="container.md" py={10} suppressHydrationWarning>
        <Text>Loading question...</Text>
      </Container>
    );
  }

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
        <Text>Tidak ada soal untuk tes ini.</Text>
        <Button onClick={() => router.push('/student')} mt={4}>
          Kembali
        </Button>
      </Container>
    );
  }

  const currentQuestion = sessionData.soal[currentQuestionIndex];
  const getQuestionStatus = (index: number) => {
    const nomorUrut = sessionData.soal[index].nomorUrut;
    if (answers[nomorUrut]) return 'answered';
    return 'unanswered';
  };

  const handleClearAnswer = async () => {
    try {
      await axios.post(`${API_BASE}/${token}/clear-answer`, {
        nomor_urut: currentQuestion.nomorUrut,
      });
      const newAnswers = { ...answers };
      delete newAnswers[currentQuestion.nomorUrut];
      setAnswers(newAnswers);
      // Update dijawab_count locally
      if (sessionData) {
        setSessionData({
          ...sessionData,
          dijawab_count: Object.keys(newAnswers).length,
        });
      }
    } catch (error) {
      console.error('Error clearing answer:', error);
      toast({ title: 'Error membatalkan jawaban', status: 'error' });
    }
  };

  return (
    <Container maxW="container.xl" py={6}>
      {/* Timer Display - Simple */}
      <Box textAlign="center" mb={4} p={4} bg={isTimeUp ? 'red.50' : 'blue.50'} borderRadius="md" borderWidth="2px" borderColor={isTimeUp ? 'red.200' : 'blue.200'}>
        <Text fontSize="sm" color="gray.600" mb={1}>Sisa Waktu</Text>
        <Text fontSize="2xl" fontFamily="mono" fontWeight="bold" color={getTimerColor()}>
          {isTimeUp ? '⏰ WAKTU HABIS!' : formatTimeRemaining(timeRemaining)}
        </Text>
      </Box>

      <Flex gap={6} direction={{ base: 'column', lg: 'row' }}>
        {/* Main Question Area */}
        <Box flex="1">
          <Card bg="orange.50" borderWidth="2px" borderColor="orange.200" mb={4}>
            <CardBody>
              <HStack spacing={4}>
                <Box bg="orange.500" p={3} borderRadius="md" color="white" fontWeight="bold" fontSize="lg">
                  CBT
                </Box>
                <VStack align="start" spacing={0}>
                  <Text fontWeight="bold" fontSize="lg">
                    {currentQuestion.materi.mataPelajaran.nama.toUpperCase()} {currentQuestion.materi.tingkat.nama} SD KELAS {currentQuestion.materi.tingkat.nama === '1' ? 'I' : currentQuestion.materi.tingkat.nama === '2' ? 'II' : currentQuestion.materi.tingkat.nama === '3' ? 'III' : 'IV'}
                  </Text>
                  <Text fontSize="sm" color="gray.600">
                    {currentQuestion.materi.nama}
                  </Text>
                </VStack>
                <Box ml="auto">
                  <Button
                    size="sm"
                    colorScheme="orange"
                    variant="outline"
                    onClick={onOpen}
                  >
                    Daftar Soal
                  </Button>
                </Box>
              </HStack>
            </CardBody>
          </Card>

          <Card>
            <CardBody>
              <VStack spacing={6} align="stretch">
                <Badge alignSelf="flex-start" colorScheme="blue" fontSize="md" px={3} py={1}>
                  Soal No. {currentQuestion.nomorUrut}
                </Badge>

                <Text fontSize="lg" fontWeight="medium">
                  {currentQuestion.pertanyaan}
                </Text>

                {currentQuestion.gambar && Array.isArray(currentQuestion.gambar) && currentQuestion.gambar.length > 0 && (
                  <Box>
                    <Text fontSize="sm" color="gray.600" mb={2}>
                      Perhatikan gambar dibawah ini
                    </Text>
                    <VStack spacing={3}>
                      {currentQuestion.gambar
                        .sort((a, b) => a.urutan - b.urutan)
                        .map((img) => (
                          <Box key={img.id} borderWidth="1px" borderRadius="md" p={2} bg="gray.50">
                            <Image
                              src={img.filePath ? `${process.env.NEXT_PUBLIC_API_BASE}/${img.filePath.replace(/\\/g, '/')}` : ''}
                              alt={img.keterangan || 'Gambar soal'}
                              maxH="300px"
                              objectFit="contain"
                              mx="auto"
                            />
                            {img.keterangan && (
                              <Text fontSize="sm" color="gray.600" mt={2} textAlign="center">
                                {img.keterangan}
                              </Text>
                            )}
                          </Box>
                        ))}
                    </VStack>
                  </Box>
                )}

                <RadioGroup value={answers[currentQuestion.nomorUrut] || ''}>
                  <VStack spacing={3} align="stretch">
                    <Box
                      p={3}
                      borderWidth="1px"
                      borderRadius="md"
                      cursor="pointer"
                      _hover={{ bg: 'gray.50' }}
                      bg={answers[currentQuestion.nomorUrut] === 'A' ? 'orange.50' : 'white'}
                      onClick={() => handleAnswerChange(currentQuestion.nomorUrut, 'A')}
                    >
                      <Radio value="A">A. {currentQuestion.opsiA}</Radio>
                    </Box>
                    <Box
                      p={3}
                      borderWidth="1px"
                      borderRadius="md"
                      cursor="pointer"
                      _hover={{ bg: 'gray.50' }}
                      bg={answers[currentQuestion.nomorUrut] === 'B' ? 'orange.50' : 'white'}
                      onClick={() => handleAnswerChange(currentQuestion.nomorUrut, 'B')}
                    >
                      <Radio value="B">B. {currentQuestion.opsiB}</Radio>
                    </Box>
                    <Box
                      p={3}
                      borderWidth="1px"
                      borderRadius="md"
                      cursor="pointer"
                      _hover={{ bg: 'gray.50' }}
                      bg={answers[currentQuestion.nomorUrut] === 'C' ? 'orange.50' : 'white'}
                      onClick={() => handleAnswerChange(currentQuestion.nomorUrut, 'C')}
                    >
                      <Radio value="C">C. {currentQuestion.opsiC}</Radio>
                    </Box>
                    <Box
                      p={3}
                      borderWidth="1px"
                      borderRadius="md"
                      cursor="pointer"
                      _hover={{ bg: 'gray.50' }}
                      bg={answers[currentQuestion.nomorUrut] === 'D' ? 'orange.50' : 'white'}
                      onClick={() => handleAnswerChange(currentQuestion.nomorUrut, 'D')}
                    >
                      <Radio value="D">D. {currentQuestion.opsiD}</Radio>
                    </Box>
                  </VStack>
                </RadioGroup>

                <HStack justify="space-between" pt={4}>
                  <Button
                    onClick={goToPreviousQuestion}
                    isDisabled={currentQuestionIndex === 0}
                    colorScheme="orange"
                    variant="outline"
                  >
                    Sebelum
                  </Button>
                  {answers[currentQuestion.nomorUrut] && (
                    <Button
                      colorScheme="red"
                      variant="outline"
                      onClick={handleClearAnswer}
                      size="sm"
                    >
                      Batalkan Jawaban
                    </Button>
                  )}
                  {currentQuestionIndex === sessionData.soal.length - 1 ? (
                    <Button
                      colorScheme="green"
                      onClick={handleFinish}
                      isLoading={submitting}
                    >
                      Selesai
                    </Button>
                  ) : (
                    <Button
                      onClick={goToNextQuestion}
                      colorScheme="orange"
                    >
                      Selanjutnya
                    </Button>
                  )}
                </HStack>
              </VStack>
            </CardBody>
          </Card>
        </Box>

        {/* Question Navigation Sidebar - Desktop Only */}
        <Box width={{ base: 'full', lg: '300px' }} display={{ base: 'none', lg: 'block' }}>
          <Card position="sticky" top="20px">
            <CardBody>
              <VStack spacing={4} align="stretch">
                <Heading size="md" textAlign="center">Daftar Soal</Heading>
                <SimpleGrid columns={5} spacing={2}>
                  {sessionData.soal.map((q, index) => {
                    const status = getQuestionStatus(index);
                    return (
                      <Button
                        key={q.id}
                        onClick={() => goToQuestion(index)}
                        size="sm"
                        colorScheme={
                          currentQuestionIndex === index
                            ? 'gray'
                            : status === 'answered'
                            ? 'green'
                            : 'gray'
                        }
                        variant={currentQuestionIndex === index ? 'solid' : 'solid'}
                      >
                        {q.nomorUrut}
                      </Button>
                    );
                  })}
                </SimpleGrid>
                <HStack spacing={2} fontSize="xs" justify="center">
                  <HStack>
                    <Box w="12px" h="12px" bg="green.500" borderRadius="sm" />
                    <Text>Dijawab</Text>
                  </HStack>
                  <HStack>
                    <Box w="12px" h="12px" bg="gray.500" borderRadius="sm" />
                    <Text>Belum Dijawab</Text>
                  </HStack>
                </HStack>
              </VStack>
            </CardBody>
          </Card>
        </Box>
      </Flex>

      {/* Question Navigation Modal - Mobile */}
      <Modal isOpen={isOpen} onClose={onClose} size="lg">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Daftar Soal</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <SimpleGrid columns={5} spacing={3}>
              {sessionData.soal.map((q, index) => {
                const status = getQuestionStatus(index);
                return (
                  <Button
                    key={q.id}
                    onClick={() => {
                      goToQuestion(index);
                      onClose();
                    }}
                    colorScheme={
                      currentQuestionIndex === index
                        ? 'gray'
                        : status === 'answered'
                        ? 'green'
                        : 'gray'
                    }
                  >
                    {q.nomorUrut}
                  </Button>
                );
              })}
            </SimpleGrid>
            <HStack spacing={3} fontSize="sm" justify="center" mt={4}>
              <HStack>
                <Box w="12px" h="12px" bg="green.500" borderRadius="sm" />
                <Text>Dijawab</Text>
              </HStack>
              <HStack>
                <Box w="12px" h="12px" bg="gray.500" borderRadius="sm" />
                <Text>Belum</Text>
              </HStack>
            </HStack>
          </ModalBody>
          <ModalFooter>
            <Button onClick={onClose}>Tutup</Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

      {/* Confirmation Modal */}
      <Modal isOpen={showConfirmModal} onClose={handleCancelFinish} size="xl">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Konfirmasi Selesai Tes</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={6} align="stretch">
              <Box textAlign="center">
                <Text fontSize="lg" fontWeight="medium">
                  Apakah Anda yakin ingin menyelesaikan tes?
                </Text>
                <Text fontSize="sm" color="gray.600" mt={2}>
                  Pastikan semua jawaban sudah benar sebelum mengumpulkan.
                </Text>
              </Box>

              <Card>
                <CardBody>
                  <VStack spacing={4}>
                    <StatGroup width="full">
                      <Stat>
                        <StatLabel>Total Soal</StatLabel>
                        <StatNumber>{sessionData?.soal.length || 0}</StatNumber>
                      </Stat>
                      <Stat>
                        <StatLabel>Sudah Dijawab</StatLabel>
                        <StatNumber color="green.500">
                          {Object.keys(answers).length}
                        </StatNumber>
                      </Stat>
                      <Stat>
                        <StatLabel>Belum Dijawab</StatLabel>
                        <StatNumber color="red.500">
                          {(sessionData?.soal.length || 0) - Object.keys(answers).length}
                        </StatNumber>
                      </Stat>
                    </StatGroup>
                  </VStack>
                </CardBody>
              </Card>

              <Box>
                <Text fontWeight="medium" mb={3}>Status Soal:</Text>
                <SimpleGrid columns={{ base: 6, md: 8, lg: 10 }} spacing={2}>
                  {sessionData?.soal.map((q, index) => {
                    const status = getQuestionStatus(index);
                    return (
                      <Button
                        key={q.id}
                        size="sm"
                        colorScheme={
                          status === 'answered' ? 'green' : 'gray'
                        }
                        variant="solid"
                        isDisabled
                        title={status === 'answered' ? 'Sudah dijawab' : 'Belum dijawab'}
                      >
                        {q.nomorUrut}
                      </Button>
                    );
                  })}
                </SimpleGrid>
                <HStack spacing={4} fontSize="sm" justify="center" mt={3}>
                  <HStack>
                    <Box w="12px" h="12px" bg="green.500" borderRadius="sm" />
                    <Text>Dijawab</Text>
                  </HStack>
                  <HStack>
                    <Box w="12px" h="12px" bg="gray.500" borderRadius="sm" />
                    <Text>Belum Dijawab</Text>
                  </HStack>
                </HStack>
              </Box>
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button variant="outline" onClick={handleCancelFinish} mr={3}>
              Batal
            </Button>
            <Button
              colorScheme="green"
              onClick={handleConfirmFinish}
              isLoading={submitting}
            >
              Ya, Selesai Tes
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </Container>
  );
}