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
  SimpleGrid,
  HStack,
  useDisclosure,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  ModalCloseButton,
  RadioGroup,
  Radio,
  Image,
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
    pembahasan?: string;
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
  const { isOpen, onOpen, onClose } = useDisclosure();
  const [selectedQuestion, setSelectedQuestion] = useState<any>(null);
  const [currentQuestionIndex, setCurrentQuestionIndex] = useState(0);
  const [showReview, setShowReview] = useState(false);

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

  const openQuestionDetail = (question: any) => {
    setSelectedQuestion(question);
    onOpen();
  };

  const goToQuestion = (index: number) => {
    setCurrentQuestionIndex(index);
  };

  const goToNextQuestion = () => {
    if (currentQuestionIndex < result!.detailJawaban.length - 1) {
      setCurrentQuestionIndex(currentQuestionIndex + 1);
    }
  };

  const goToPreviousQuestion = () => {
    if (currentQuestionIndex > 0) {
      setCurrentQuestionIndex(currentQuestionIndex - 1);
    }
  };

  if (loading) {
    return (
      <Container maxW="container.lg" py={10}>
        <VStack spacing={6}>
          <Heading size="lg">Loading Test Results...</Heading>
          <Box p={8} bg="blue.50" borderRadius="lg" w="full" textAlign="center">
            <Text fontSize="lg" color="blue.600">Please wait while we fetch your results</Text>
            <Text fontSize="sm" color="gray.600" mt={2}>This may take a few moments...</Text>
          </Box>
        </VStack>
      </Container>
    );
  }

  if (!result?.sessionInfo) {
    return (
      <Container maxW="container.lg" py={10}>
        <VStack spacing={6}>
          <Heading size="lg" color="red.500">Results Not Available</Heading>
          <Box p={8} bg="red.50" borderRadius="lg" w="full" textAlign="center">
            <Text fontSize="lg" color="red.600">Unable to load test results</Text>
            <Text fontSize="sm" color="gray.600" mt={2}>Please check your session token or try again later</Text>
            <Link href="/student">
              <Button mt={4} colorScheme="blue">Back to Home</Button>
            </Link>
          </Box>
        </VStack>
      </Container>
    );
  }

  const sessionInfo = result.sessionInfo;
  const scorePercentage = sessionInfo.nilaiAkhir || 0;
  const isPassed = scorePercentage >= 70; // Assuming 70% pass mark

  // Calculate actual duration from start and end time
  const startTime = new Date(sessionInfo.waktuMulai);
  const endTime = new Date(sessionInfo.waktuSelesai);
  const actualDurationMinutes = Math.round((endTime.getTime() - startTime.getTime()) / (1000 * 60));

  return (
    <Container maxW="container.lg" py={10}>
      <VStack spacing={8}>
        <Box textAlign="center">
          <Heading as="h1" size="2xl" color="blue.600" mb={2}>
            📊 Test Results
          </Heading>
          <Text fontSize="lg" color="gray.600">
            Review your performance and learn from the questions
          </Text>
        </Box>

        <Card width="full" shadow="lg" borderRadius="xl">
          <CardBody>
            <VStack spacing={6}>
              <Box textAlign="center" p={6} bg={isPassed ? 'green.50' : 'red.50'} borderRadius="lg" w="full">
                <Text fontSize="4xl" fontWeight="bold" color={isPassed ? 'green.600' : 'red.600'} mb={2}>
                  {scorePercentage.toFixed(1)}%
                </Text>
                <Badge colorScheme={isPassed ? 'green' : 'red'} fontSize="lg" px={4} py={2} borderRadius="full">
                  {isPassed ? '✅ PASSED' : '❌ FAILED'}
                </Badge>
                <Text fontSize="sm" color="gray.600" mt={2}>
                  Passing score: 70%
                </Text>
              </Box>

              <SimpleGrid columns={{ base: 2, md: 4 }} spacing={6} w="full">
                <Stat textAlign="center">
                  <StatLabel fontSize="sm" color="gray.600">👤 Participant</StatLabel>
                  <StatNumber fontSize="lg" color="blue.600">{sessionInfo.namaPeserta}</StatNumber>
                </Stat>
                <Stat textAlign="center">
                  <StatLabel fontSize="sm" color="gray.600">📚 Subject</StatLabel>
                  <StatNumber fontSize="lg" color="purple.600">{sessionInfo.mataPelajaran.nama}</StatNumber>
                </Stat>
                <Stat textAlign="center">
                  <StatLabel fontSize="sm" color="gray.600">📊 Score</StatLabel>
                  <StatNumber fontSize="lg" color={isPassed ? 'green.600' : 'red.600'}>
                    {sessionInfo.jumlahBenar}/{sessionInfo.totalSoal}
                  </StatNumber>
                </Stat>
                <Stat textAlign="center">
                  <StatLabel fontSize="sm" color="gray.600">⏱️ Duration</StatLabel>
                  <StatNumber fontSize="lg" color="orange.600">{actualDurationMinutes} min</StatNumber>
                </Stat>
              </SimpleGrid>

              <Box width="full" p={4} bg="gray.50" borderRadius="lg">
                <Text fontWeight="medium" mb={3} color="gray.700">📅 Session Details:</Text>
                <SimpleGrid columns={{ base: 1, md: 2 }} spacing={3}>
                  <Text>🏫 Level: <Badge colorScheme="purple">{sessionInfo.tingkat.nama}</Badge></Text>
                  <Text>📊 Status: <Badge colorScheme={sessionInfo.status === 'COMPLETED' ? 'green' : 'yellow'}>{sessionInfo.status}</Badge></Text>
                  <Text>🕐 Started: {new Date(sessionInfo.waktuMulai).toLocaleString()}</Text>
                  <Text>🏁 Completed: {new Date(sessionInfo.waktuSelesai).toLocaleString()}</Text>
                </SimpleGrid>
              </Box>
            </VStack>
          </CardBody>
        </Card>

        {/* Question Review Section */}
        <Card width="full" shadow="lg" borderRadius="xl">
          <CardBody>
            <VStack spacing={6} align="stretch">
              <Box textAlign="center">
                <Heading size="lg" color="blue.600">🔍 Question Review</Heading>
                <Text fontSize="sm" color="gray.600" mt={1}>
                  Click on question numbers to review answers and explanations
                </Text>
              </Box>
              <SimpleGrid columns={{ base: 4, md: 6, lg: 8 }} spacing={2}>
                {result.detailJawaban.map((jawaban) => {
                  let colorScheme = 'gray';
                  let statusText = 'Tidak Menjawab';

                  if (jawaban.jawabanDipilih) {
                    if (jawaban.isCorrect) {
                      colorScheme = 'green';
                      statusText = 'Benar';
                    } else {
                      colorScheme = 'red';
                      statusText = 'Salah';
                    }
                  }

                  return (
                    <Button
                      key={jawaban.nomorUrut}
                      onClick={() => openQuestionDetail(jawaban)}
                      size="sm"
                      colorScheme={colorScheme}
                      variant="solid"
                      title={statusText}
                      borderRadius="full"
                      fontSize="md"
                      fontWeight="bold"
                    >
                      {jawaban.nomorUrut}
                    </Button>
                  );
                })}
              </SimpleGrid>
              <HStack spacing={6} fontSize="sm" justify="center">
                <HStack>
                  <Box w="12px" h="12px" bg="green.500" borderRadius="sm" />
                  <Text>✅ Correct</Text>
                </HStack>
                <HStack>
                  <Box w="12px" h="12px" bg="red.500" borderRadius="sm" />
                  <Text>❌ Wrong</Text>
                </HStack>
                <HStack>
                  <Box w="12px" h="12px" bg="gray.500" borderRadius="sm" />
                  <Text>⚪ Not Answered</Text>
                </HStack>
              </HStack>
              <Button
                colorScheme="blue"
                width="full"
                onClick={() => setShowReview(true)}
                mt={4}
              >
                Lihat Pembahasan Lengkap
              </Button>
            </VStack>
          </CardBody>
        </Card>

        {/* Detailed Question Review */}
        {showReview && (
          <Card width="full">
            <CardBody>
              <VStack spacing={6} align="stretch">
                <HStack justify="space-between">
                  <Heading size="md">Pembahasan Soal</Heading>
                  <Button size="sm" variant="outline" onClick={() => setShowReview(false)}>
                    Sembunyikan
                  </Button>
                </HStack>

                {/* Question Navigation */}
                <Box>
                  <Text fontWeight="medium" mb={2}>Daftar Soal</Text>
                  <SimpleGrid columns={{ base: 8, md: 10, lg: 12 }} spacing={2}>
                    {result.detailJawaban.map((jawaban, index) => {
                      let colorScheme = 'gray';
                      if (jawaban.jawabanDipilih) {
                        colorScheme = jawaban.isCorrect ? 'green' : 'red';
                      }
                      return (
                        <Button
                          key={jawaban.nomorUrut}
                          onClick={() => goToQuestion(index)}
                          size="sm"
                          colorScheme={colorScheme}
                          variant={currentQuestionIndex === index ? 'solid' : 'outline'}
                          borderWidth={currentQuestionIndex === index ? '2px' : '1px'}
                        >
                          {jawaban.nomorUrut}
                        </Button>
                      );
                    })}
                  </SimpleGrid>
                </Box>

                {/* Current Question Detail */}
                {(() => {
                  const currentJawaban = result.detailJawaban[currentQuestionIndex];
                  return (
                    <Card bg="gray.50">
                      <CardBody>
                        <VStack spacing={4} align="stretch">
                          <HStack justify="space-between">
                            <Badge colorScheme="blue" fontSize="md" px={3} py={1}>
                              Soal No. {currentJawaban.nomorUrut}
                            </Badge>
                            <Badge
                              colorScheme={
                                !currentJawaban.jawabanDipilih
                                  ? 'gray'
                                  : currentJawaban.isCorrect
                                  ? 'green'
                                  : 'red'
                              }
                              fontSize="md"
                            >
                              {!currentJawaban.jawabanDipilih
                                ? 'Tidak Menjawab'
                                : currentJawaban.isCorrect
                                ? 'Benar ✓'
                                : 'Salah ✗'}
                            </Badge>
                          </HStack>

                          <Text fontSize="lg" fontWeight="medium">
                            {currentJawaban.pertanyaan}
                          </Text>

                          {/* Gambar Soal */}
                          {currentJawaban.gambar && Array.isArray(currentJawaban.gambar) && currentJawaban.gambar.length > 0 && (
                            <Box mt={4}>
                              <Text fontSize="sm" color="gray.600" mb={3} fontWeight="medium">
                                🖼️ Gambar Pendukung
                              </Text>
                              <SimpleGrid columns={{ base: 1, md: 2 }} spacing={3}>
                                {currentJawaban.gambar
                                  .sort((a, b) => a.urutan - b.urutan)
                                  .map((img) => (
                                    <Box key={img.id} borderWidth="2px" borderRadius="lg" p={3} bg="white" borderColor="gray.200" shadow="sm">
                                      <Image
                                        src={img.filePath ? `http://localhost:8080/${img.filePath.replace(/\\/g, '/')}` : ''}
                                        alt={img.keterangan || 'Gambar soal'}
                                        maxH="250px"
                                        objectFit="contain"
                                        mx="auto"
                                        borderRadius="md"
                                      />
                                      {img.keterangan && (
                                        <Text fontSize="sm" color="gray.700" mt={2} textAlign="center" fontStyle="italic">
                                          {img.keterangan}
                                        </Text>
                                      )}
                                      <Text fontSize="xs" color="gray.500" mt={1} textAlign="center">
                                        Gambar {img.urutan}
                                      </Text>
                                    </Box>
                                  ))}
                              </SimpleGrid>
                            </Box>
                          )}

                          {/* Options */}
                          <VStack spacing={3} align="stretch">
                            {['A', 'B', 'C', 'D'].map((option) => {
                              const isCorrectAnswer = currentJawaban.jawabanBenar === option;
                              const isUserAnswer = currentJawaban.jawabanDipilih === option;
                              const optionText = currentJawaban[`opsi${option}` as keyof typeof currentJawaban];

                              let bgColor = 'white';
                              let borderColor = 'gray.200';
                              let borderWidth = '1px';

                              if (isCorrectAnswer) {
                                bgColor = 'green.50';
                                borderColor = 'green.400';
                                borderWidth = '2px';
                              } else if (isUserAnswer && !isCorrectAnswer) {
                                bgColor = 'red.50';
                                borderColor = 'red.400';
                                borderWidth = '2px';
                              }

                              return (
                                <Box
                                  key={option}
                                  p={4}
                                  borderWidth={borderWidth}
                                  borderColor={borderColor}
                                  borderRadius="md"
                                  bg={bgColor}
                                >
                                  <HStack justify="space-between">
                                    <Text fontWeight={isCorrectAnswer || isUserAnswer ? 'bold' : 'normal'}>
                                      {option}. {optionText}
                                    </Text>
                                    <HStack spacing={2}>
                                      {isCorrectAnswer && (
                                        <Badge colorScheme="green">Jawaban Benar</Badge>
                                      )}
                                      {isUserAnswer && !isCorrectAnswer && (
                                        <Badge colorScheme="red">Jawaban Anda</Badge>
                                      )}
                                    </HStack>
                                  </HStack>
                                </Box>
                              );
                            })}
                          </VStack>

                          {/* Pembahasan */}
                          {currentJawaban.pembahasan && currentJawaban.pembahasan.trim() && (
                            <Box mt={6} p={6} bg="blue.50" borderRadius="lg" border="2px solid" borderColor="blue.200" position="relative">
                              <Box
                                position="absolute"
                                top="-12px"
                                left="20px"
                                bg="blue.500"
                                color="white"
                                px={3}
                                py={1}
                                borderRadius="md"
                                fontSize="sm"
                                fontWeight="bold"
                              >
                                📚 Pembahasan
                              </Box>
                              <Text color="blue.800" whiteSpace="pre-wrap" lineHeight="1.6" mt={2}>
                                {currentJawaban.pembahasan}
                              </Text>
                            </Box>
                          )}

                          {/* Navigation Buttons */}
                          <HStack justify="space-between" pt={4}>
                            <Button
                              leftIcon={<Text>◀</Text>}
                              onClick={goToPreviousQuestion}
                              isDisabled={currentQuestionIndex === 0}
                              colorScheme="blue"
                              variant="outline"
                            >
                              Sebelumnya
                            </Button>
                            <Text fontSize="sm" color="gray.600">
                              {currentQuestionIndex + 1} / {result.detailJawaban.length}
                            </Text>
                            <Button
                              rightIcon={<Text>▶</Text>}
                              onClick={goToNextQuestion}
                              isDisabled={currentQuestionIndex === result.detailJawaban.length - 1}
                              colorScheme="blue"
                            >
                              Selanjutnya
                            </Button>
                          </HStack>
                        </VStack>
                      </CardBody>
                    </Card>
                  );
                })()}
              </VStack>
            </CardBody>
          </Card>
        )}

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

      {/* Question Detail Modal */}
      <Modal isOpen={isOpen} onClose={onClose} size="4xl" scrollBehavior="inside">
        <ModalOverlay backdropFilter="blur(4px)" />
        <ModalContent borderRadius="2xl" shadow="2xl">
          <ModalHeader bg="blue.50" borderRadius="2xl 2xl 0 0" pb={4}>
            <HStack justify="space-between" align="center">
              <Box>
                <Text fontSize="xl" fontWeight="bold" color="blue.700">
                  📝 Question {selectedQuestion?.nomorUrut}
                </Text>
                <Text fontSize="sm" color="gray.600" mt={1}>
                  Review your answer and learn from the explanation
                </Text>
              </Box>
              <Badge
                size="lg"
                colorScheme={
                  !selectedQuestion?.jawabanDipilih
                    ? 'gray'
                    : selectedQuestion?.isCorrect
                    ? 'green'
                    : 'red'
                }
                fontSize="md"
                px={4}
                py={2}
                borderRadius="full"
              >
                {!selectedQuestion?.jawabanDipilih
                  ? '⚪ Not Answered'
                  : selectedQuestion?.isCorrect
                  ? '✅ Correct'
                  : '❌ Wrong'}
              </Badge>
            </HStack>
          </ModalHeader>
          <ModalCloseButton />
          <ModalBody p={6}>
            {selectedQuestion && (
              <VStack spacing={4} align="stretch">
                <Text fontSize="lg" fontWeight="medium">
                  {selectedQuestion.pertanyaan}
                </Text>

                <RadioGroup
                  value={selectedQuestion.jawabanDipilih || ''}
                  isReadOnly
                >
                  <VStack spacing={3} align="stretch">
                    <Box
                      p={3}
                      borderWidth="1px"
                      borderRadius="md"
                      bg={
                        selectedQuestion.jawabanBenar === 'A'
                          ? 'green.50'
                          : selectedQuestion.jawabanDipilih === 'A'
                          ? 'red.50'
                          : 'white'
                      }
                      borderColor={
                        selectedQuestion.jawabanBenar === 'A'
                          ? 'green.300'
                          : selectedQuestion.jawabanDipilih === 'A'
                          ? 'red.300'
                          : 'gray.200'
                      }
                    >
                      <Radio value="A" isReadOnly>
                        A. {selectedQuestion.opsiA}
                        {selectedQuestion.jawabanBenar === 'A' && (
                          <Badge ml={2} colorScheme="green">Jawaban Benar</Badge>
                        )}
                        {selectedQuestion.jawabanDipilih === 'A' && selectedQuestion.jawabanBenar !== 'A' && (
                          <Badge ml={2} colorScheme="red">Jawaban Anda</Badge>
                        )}
                      </Radio>
                    </Box>
                    <Box
                      p={3}
                      borderWidth="1px"
                      borderRadius="md"
                      bg={
                        selectedQuestion.jawabanBenar === 'B'
                          ? 'green.50'
                          : selectedQuestion.jawabanDipilih === 'B'
                          ? 'red.50'
                          : 'white'
                      }
                      borderColor={
                        selectedQuestion.jawabanBenar === 'B'
                          ? 'green.300'
                          : selectedQuestion.jawabanDipilih === 'B'
                          ? 'red.300'
                          : 'gray.200'
                      }
                    >
                      <Radio value="B" isReadOnly>
                        B. {selectedQuestion.opsiB}
                        {selectedQuestion.jawabanBenar === 'B' && (
                          <Badge ml={2} colorScheme="green">Jawaban Benar</Badge>
                        )}
                        {selectedQuestion.jawabanDipilih === 'B' && selectedQuestion.jawabanBenar !== 'B' && (
                          <Badge ml={2} colorScheme="red">Jawaban Anda</Badge>
                        )}
                      </Radio>
                    </Box>
                    <Box
                      p={3}
                      borderWidth="1px"
                      borderRadius="md"
                      bg={
                        selectedQuestion.jawabanBenar === 'C'
                          ? 'green.50'
                          : selectedQuestion.jawabanDipilih === 'C'
                          ? 'red.50'
                          : 'white'
                      }
                      borderColor={
                        selectedQuestion.jawabanBenar === 'C'
                          ? 'green.300'
                          : selectedQuestion.jawabanDipilih === 'C'
                          ? 'red.300'
                          : 'gray.200'
                      }
                    >
                      <Radio value="C" isReadOnly>
                        C. {selectedQuestion.opsiC}
                        {selectedQuestion.jawabanBenar === 'C' && (
                          <Badge ml={2} colorScheme="green">Jawaban Benar</Badge>
                        )}
                        {selectedQuestion.jawabanDipilih === 'C' && selectedQuestion.jawabanBenar !== 'C' && (
                          <Badge ml={2} colorScheme="red">Jawaban Anda</Badge>
                        )}
                      </Radio>
                    </Box>
                    <Box
                      p={3}
                      borderWidth="1px"
                      borderRadius="md"
                      bg={
                        selectedQuestion.jawabanBenar === 'D'
                          ? 'green.50'
                          : selectedQuestion.jawabanDipilih === 'D'
                          ? 'red.50'
                          : 'white'
                      }
                      borderColor={
                        selectedQuestion.jawabanBenar === 'D'
                          ? 'green.300'
                          : selectedQuestion.jawabanDipilih === 'D'
                          ? 'red.300'
                          : 'gray.200'
                      }
                    >
                      <Radio value="D" isReadOnly>
                        D. {selectedQuestion.opsiD}
                        {selectedQuestion.jawabanBenar === 'D' && (
                          <Badge ml={2} colorScheme="green">Jawaban Benar</Badge>
                        )}
                        {selectedQuestion.jawabanDipilih === 'D' && selectedQuestion.jawabanBenar !== 'D' && (
                          <Badge ml={2} colorScheme="red">Jawaban Anda</Badge>
                        )}
                      </Radio>
                    </Box>
                  </VStack>
                </RadioGroup>
              </VStack>
            )}
          </ModalBody>
          <ModalFooter bg="gray.50" borderRadius="0 0 2xl 2xl">
            <HStack spacing={4} w="full" justify="space-between">
              <Button
                leftIcon={<Text>◀</Text>}
                onClick={() => {
                  const currentIndex = result.detailJawaban.findIndex(q => q.nomorUrut === selectedQuestion?.nomorUrut);
                  if (currentIndex > 0) {
                    setSelectedQuestion(result.detailJawaban[currentIndex - 1]);
                  }
                }}
                isDisabled={result.detailJawaban.findIndex(q => q.nomorUrut === selectedQuestion?.nomorUrut) === 0}
                colorScheme="blue"
                variant="outline"
              >
                Previous
              </Button>
              <Text fontSize="sm" color="gray.600" alignSelf="center">
                Question {selectedQuestion?.nomorUrut} of {result.detailJawaban.length}
              </Text>
              <HStack>
                <Button
                  rightIcon={<Text>▶</Text>}
                  onClick={() => {
                    const currentIndex = result.detailJawaban.findIndex(q => q.nomorUrut === selectedQuestion?.nomorUrut);
                    if (currentIndex < result.detailJawaban.length - 1) {
                      setSelectedQuestion(result.detailJawaban[currentIndex + 1]);
                    }
                  }}
                  isDisabled={result.detailJawaban.findIndex(q => q.nomorUrut === selectedQuestion?.nomorUrut) === result.detailJawaban.length - 1}
                  colorScheme="blue"
                >
                  Next
                </Button>
                <Button onClick={onClose} colorScheme="gray">
                  Close
                </Button>
              </HStack>
            </HStack>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </Container>
  );
}