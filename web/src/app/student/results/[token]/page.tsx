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
  const [materi, setMateri] = useState<{id: number; nama: string} | null>(null);
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
      const resultData = response.data as TestResultResponse;
      setResult(resultData);
      
      // Fetch materi data
      await fetchMateri(resultData.sessionInfo.mataPelajaran.id, resultData.sessionInfo.tingkat.id);
    } catch (error) {
      console.error('Error fetching result:', error);
      toast({ title: 'Error loading results', status: 'error' });
    } finally {
      setLoading(false);
    }
  };

  const fetchMateri = async (mataPelajaranId: number, tingkatId: number) => {
    try {
      const response = await axios.get('http://localhost:8080/v1/topics');
      const materiList = response.data.materi || [];
      
      // Find materi that matches the session's mataPelajaran and tingkat
      const matchingMateri = materiList.find((m: any) => 
        m.mataPelajaran.id === mataPelajaranId && m.tingkat.id === tingkatId
      );
      
      if (matchingMateri) {
        setMateri({ id: matchingMateri.id, nama: matchingMateri.nama });
      }
    } catch (error) {
      console.error('Error fetching materi:', error);
      // Don't show error toast for materi fetch failure
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
        {/* Header Box - Orange Design */}
        <Card width="full" bg="orange.50" borderWidth="2px" borderColor="orange.200" borderRadius="xl">
          <CardBody>
            <VStack spacing={6}>
              <HStack justify="center" spacing={4}>
                <Box bg="orange.500" p={4} borderRadius="md" color="white" fontWeight="bold" fontSize="lg">
                  CBT
                </Box>
                <VStack align="start" spacing={0}>
                  <Text fontWeight="bold" fontSize="xl" color="orange.700">
                    {sessionInfo.mataPelajaran.nama.toUpperCase()} {sessionInfo.tingkat.nama} SD KELAS {sessionInfo.tingkat.nama === '1' ? 'I' : sessionInfo.tingkat.nama === '2' ? 'II' : sessionInfo.tingkat.nama === '3' ? 'III' : 'IV'}{materi ? ` - ${materi.nama.toUpperCase()}` : ''}
                  </Text>
                  <Text fontSize="sm" color="gray.600">
                    {sessionInfo.namaPeserta || 'Hasil Tes Anda'}
                  </Text>
                </VStack>
              </HStack>
            </VStack>
          </CardBody>
        </Card>

        {/* Score Card - Big Number */}
        <Card width="full" borderWidth="2px" borderColor="gray.200" borderRadius="xl">
          <CardBody py={8}>
            <VStack spacing={6}>
              <Box textAlign="center">
                <Badge colorScheme="orange" fontSize="md" px={4} py={2} borderRadius="md" mb={4}>
                  Nilai CBT
                </Badge>
                <Text fontSize="sm" color="gray.600" mb={2}>
                  Total nilai kamu adalah
                </Text>
                <Text fontSize="8xl" fontWeight="bold" color="orange.500" lineHeight="1">
                  {scorePercentage.toFixed(2)}
                </Text>
                <Text fontSize="md" color="gray.600" mt={4}>
                  Selamat kamu mendapatkan nilai yang bagus! Tingkatkan terus belajar kamu, agar meraih angka yang lebih baik lagi!
                </Text>
              </Box>

              {/* Buttons */}
              <HStack spacing={4} width="full" justify="center" mt={4}>
                <Button
                  variant="outline"
                  colorScheme="orange"
                  size="md"
                  onClick={() => setShowReview(!showReview)}
                >
                  Bagikan Ulang
                </Button>
                <Button
                  colorScheme="orange"
                  size="md"
                  onClick={() => setShowReview(!showReview)}
                >
                  Pembahasan Kunci
                </Button>
              </HStack>
            </VStack>
          </CardBody>
        </Card>

        {/* Review Section - Collapsible */}
        {showReview && (
          <Card width="full" borderWidth="2px" borderColor="blue.200" borderRadius="xl" bg="blue.50">
            <CardBody>
              <VStack spacing={6} align="stretch">
                <Box textAlign="center">
                  <Heading size="md" color="blue.700">Pembahasan Soal</Heading>
                </Box>

                {/* Stats */}
                <Box bg="white" p={4} borderRadius="md">
                  <VStack spacing={4}>
                    <SimpleGrid columns={materi ? 4 : 3} spacing={4} w="full">
                      <Stat textAlign="center">
                        <StatLabel fontSize="sm" color="gray.600">Nama Siswa</StatLabel>
                        <StatNumber fontSize="md" color="gray.800">{sessionInfo.namaPeserta}</StatNumber>
                      </Stat>
                      <Stat textAlign="center">
                        <StatLabel fontSize="sm" color="gray.600">Mata Pelajaran</StatLabel>
                        <StatNumber fontSize="md" color="gray.800">{sessionInfo.mataPelajaran.nama}</StatNumber>
                      </Stat>
                      <Stat textAlign="center">
                        <StatLabel fontSize="sm" color="gray.600">Kelas</StatLabel>
                        <StatNumber fontSize="md" color="gray.800">{sessionInfo.tingkat.nama}</StatNumber>
                      </Stat>
                      {materi && (
                        <Stat textAlign="center">
                          <StatLabel fontSize="sm" color="gray.600">Materi</StatLabel>
                          <StatNumber fontSize="md" color="gray.800">{materi.nama}</StatNumber>
                        </Stat>
                      )}
                    </SimpleGrid>
                    <Button colorScheme="orange" size="sm" width="fit-content">
                      Bagikan
                    </Button>
                  </VStack>
                </Box>

                {/* Two Column Layout: Daftar Soal + Question Detail */}
                <SimpleGrid columns={{ base: 1, lg: 2 }} spacing={6} alignItems="start">
                  {/* Left Column: Daftar Soal */}
                  <Box bg="white" p={6} borderRadius="md" height="fit-content" position="sticky" top="20px">
                    <Heading size="sm" mb={4}>Daftar Soal</Heading>
                    <SimpleGrid columns={5} spacing={2}>
                      {result.detailJawaban.map((jawaban) => {
                        let colorScheme = 'red';
                        if (jawaban.isCorrect) {
                          colorScheme = 'green';
                        } else if (!jawaban.jawabanDipilih) {
                          colorScheme = 'gray';
                        }

                        const isSelected = currentQuestionIndex === result.detailJawaban.findIndex(j => j.nomorUrut === jawaban.nomorUrut);

                        return (
                          <Button
                            key={jawaban.nomorUrut}
                            onClick={() => {
                              setCurrentQuestionIndex(result.detailJawaban.findIndex(j => j.nomorUrut === jawaban.nomorUrut));
                            }}
                            size="md"
                            colorScheme={colorScheme}
                            variant="solid"
                            borderRadius="md"
                            borderWidth={isSelected ? '3px' : '0px'}
                            borderColor={isSelected ? 'blue.500' : 'transparent'}
                            _hover={{ transform: 'scale(1.05)' }}
                            transition="all 0.2s"
                          >
                            {jawaban.nomorUrut}
                          </Button>
                        );
                      })}
                    </SimpleGrid>
                    <VStack spacing={2} fontSize="xs" align="start" mt={4} px={2}>
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
                    </VStack>
                  </Box>

                  {/* Right Column: Question Detail */}
                  {(() => {
                    const currentJawaban = result.detailJawaban[currentQuestionIndex];
                    return (
                      <Card bg="white" borderRadius="md">
                        <CardBody>
                          <VStack spacing={6} align="stretch">
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
                                px={3}
                                py={1}
                              >
                                {!currentJawaban.jawabanDipilih
                                  ? 'Tidak Menjawab'
                                  : currentJawaban.isCorrect
                                  ? 'Benar'
                                  : 'Jawaban kamu salah'}
                              </Badge>
                            </HStack>

                            <Text fontSize="lg" fontWeight="medium">
                              {currentJawaban.pertanyaan}
                            </Text>

                            {/* Images */}
                            {currentJawaban.gambar && Array.isArray(currentJawaban.gambar) && currentJawaban.gambar.length > 0 && (
                              <Box>
                                <Text fontSize="sm" color="gray.600" mb={2}>
                                  Perhatikan gambar dibawah ini
                                </Text>
                                <SimpleGrid columns={{ base: 1, md: 2 }} spacing={4}>
                                  {currentJawaban.gambar
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
                                </SimpleGrid>
                              </Box>
                            )}

                            {/* Options */}
                            <VStack spacing={3} align="stretch">
                              <Text fontSize="sm" color="gray.600" mb={-2}>
                                Salah satu manfaat dari gunung bagi manusia adalah ....
                              </Text>
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
                                    p={3}
                                    borderWidth={borderWidth}
                                    borderColor={borderColor}
                                    borderRadius="md"
                                    bg={bgColor}
                                  >
                                    <HStack justify="space-between" align="start">
                                      <Text fontWeight={isCorrectAnswer || isUserAnswer ? 'semibold' : 'normal'} flex="1">
                                        {option}. {optionText}
                                      </Text>
                                      {isUserAnswer && !isCorrectAnswer && (
                                        <Badge colorScheme="red" ml={2}>Jawaban kamu salah</Badge>
                                      )}
                                    </HStack>
                                  </Box>
                                );
                              })}
                            </VStack>

                            {/* Kunci Jawaban Label */}
                            <Box>
                              <Text fontSize="sm" fontWeight="bold" color="green.700">
                                Kunci Jawaban: {currentJawaban.jawabanBenar}
                              </Text>
                            </Box>

                            {/* Pembahasan */}
                            {currentJawaban.pembahasan && currentJawaban.pembahasan.trim() ? (
                              <Box p={4} bg="gray.50" borderRadius="md" borderLeft="4px solid" borderLeftColor="blue.400">
                                <Text fontWeight="bold" mb={2} color="gray.700">Pembahasan :</Text>
                                <Text color="gray.700" whiteSpace="pre-wrap" lineHeight="1.6">
                                  {currentJawaban.pembahasan}
                                </Text>
                              </Box>
                            ) : (
                              <Box p={4} bg="gray.50" borderRadius="md" borderLeft="4px solid" borderLeftColor="gray.400">
                                <Text fontWeight="bold" mb={2} color="gray.600">Pembahasan :</Text>
                                <Text color="gray.500" fontStyle="italic">
                                  Pembahasan tidak tersedia untuk soal ini.
                                </Text>
                              </Box>
                            )}
                          </VStack>
                        </CardBody>
                      </Card>
                    );
                  })()}
                </SimpleGrid>
              </VStack>
            </CardBody>
          </Card>
        )}

        <VStack spacing={4}>
          <Link href="/student/history">
            <Button colorScheme="orange" size="lg" variant="outline">
              Lihat Riwayat Saya
            </Button>
          </Link>
          <Link href="/student">
            <Button variant="outline" size="lg">
              Kembali ke Beranda
            </Button>
          </Link>
        </VStack>
      </VStack>
    </Container>
  );
}