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
  const [currentQuestionIndex, setCurrentQuestionIndex] = useState(0);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [mounted, setMounted] = useState(false);
  const { isOpen, onOpen, onClose } = useDisclosure();
  const [showConfirmModal, setShowConfirmModal] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

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
        if (q.jawabanDipilih && q.jawabanDipilih !== 'JAWABAN_INVALID') {
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
      // Refresh data to update answered count
      fetchAllQuestions();
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

  const confirmFinish = async () => {
    setSubmitting(true);
    try {
      await axios.post(`${API_BASE}/${token}/complete`);
      toast({ title: 'Tes selesai!', status: 'success' });
      router.push(`/student/results/${token}`);
    } catch (error) {
      console.error('Error completing test:', error);
      toast({ title: 'Error menyelesaikan tes', status: 'error' });
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
      fetchAllQuestions();
    } catch (error) {
      console.error('Error clearing answer:', error);
      toast({ title: 'Error membatalkan jawaban', status: 'error' });
    }
  };

  return (
    <Container maxW="container.xl" py={6}>
      <Flex gap={6} direction={{ base: 'column', lg: 'row' }}>
        {/* Main Question Area */}
        <Box flex="1">
          <Card bg="blue.50" borderWidth="2px" borderColor="blue.200" mb={4}>
            <CardBody>
              <HStack spacing={4}>
                <Box bg="orange.400" p={3} borderRadius="md">
                  <Text fontSize="2xl">📚</Text>
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
                    Daftar Soal 📋
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
                              src={img.filePath ? `http://localhost:8080/${img.filePath.replace(/\\/g, '/')}` : ''}
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

                <RadioGroup
                  value={answers[currentQuestion.nomorUrut] || ''}
                  onChange={(value) => handleAnswerChange(currentQuestion.nomorUrut, value)}
                >
                  <VStack spacing={3} align="stretch">
                    <Box
                      p={3}
                      borderWidth="1px"
                      borderRadius="md"
                      cursor="pointer"
                      _hover={{ bg: 'gray.50' }}
                      bg={answers[currentQuestion.nomorUrut] === 'A' ? 'orange.50' : 'white'}
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
                    >
                      <Radio value="D">D. {currentQuestion.opsiD}</Radio>
                    </Box>
                  </VStack>
                </RadioGroup>

                <HStack justify="space-between" pt={4}>
                  <Button
                    leftIcon={<Text>◀</Text>}
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
                      Selesai ✅
                    </Button>
                  ) : (
                    <Button
                      rightIcon={<Text>▶</Text>}
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