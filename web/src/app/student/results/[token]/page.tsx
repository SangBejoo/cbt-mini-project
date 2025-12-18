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

  // Calculate actual duration from start and end time
  const startTime = new Date(sessionInfo.waktuMulai);
  const endTime = new Date(sessionInfo.waktuSelesai);
  const actualDurationMinutes = Math.round((endTime.getTime() - startTime.getTime()) / (1000 * 60));

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
                  <StatNumber>{actualDurationMinutes} minutes</StatNumber>
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

        {/* Question Review Section */}
        <Card width="full">
          <CardBody>
            <VStack spacing={4} align="stretch">
              <Heading size="md">Review Soal</Heading>
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
                    >
                      {jawaban.nomorUrut}
                    </Button>
                  );
                })}
              </SimpleGrid>
              <HStack spacing={4} fontSize="sm" justify="center">
                <HStack>
                  <Box w="12px" h="12px" bg="green.500" borderRadius="sm" />
                  <Text>Benar</Text>
                </HStack>
                <HStack>
                  <Box w="12px" h="12px" bg="red.500" borderRadius="sm" />
                  <Text>Salah</Text>
                </HStack>
                <HStack>
                  <Box w="12px" h="12px" bg="gray.500" borderRadius="sm" />
                  <Text>Tidak Menjawab</Text>
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
                            <Box>
                              <Text fontSize="sm" color="gray.600" mb={2}>
                                Perhatikan gambar dibawah ini
                              </Text>
                              <VStack spacing={3}>
                                {currentJawaban.gambar
                                  .sort((a, b) => a.urutan - b.urutan)
                                  .map((img) => (
                                    <Box key={img.id} borderWidth="1px" borderRadius="md" p={2} bg="white">
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
                          {currentJawaban.pembahasan && (
                            <Box mt={6} p={4} bg="blue.50" borderRadius="md" border="1px solid" borderColor="blue.200">
                              <Text fontWeight="bold" color="blue.800" mb={2}>
                                Pembahasan:
                              </Text>
                              <Text color="blue.700" whiteSpace="pre-wrap">
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
      <Modal isOpen={isOpen} onClose={onClose} size="xl">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>
            Soal No. {selectedQuestion?.nomorUrut}
            <Badge
              ml={2}
              colorScheme={
                !selectedQuestion?.jawabanDipilih
                  ? 'gray'
                  : selectedQuestion?.isCorrect
                  ? 'green'
                  : 'red'
              }
            >
              {!selectedQuestion?.jawabanDipilih
                ? 'Tidak Menjawab'
                : selectedQuestion?.isCorrect
                ? 'Benar'
                : 'Salah'}
            </Badge>
          </ModalHeader>
          <ModalCloseButton />
          <ModalBody>
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
          <ModalFooter>
            <Button onClick={onClose}>Tutup</Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </Container>
  );
}