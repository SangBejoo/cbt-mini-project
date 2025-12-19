'use client';

import { useState, useEffect, useRef, useCallback, useMemo } from 'react';
import { debounce } from 'lodash';
import {
  Box,
  Button,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  IconButton,
  useDisclosure,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  FormControl,
  FormLabel,
  Input,
  Select,
  Textarea,
  useToast,
  HStack,
  VStack,
  Text,
  Badge,
  Image as ChakraImage,
  Heading,
  SimpleGrid,
  Divider,
  Stepper,
  Step,
  StepIndicator,
  StepStatus,
  StepTitle,
  StepSeparator,
  Accordion,
  AccordionItem,
  AccordionButton,
  AccordionPanel,
  AccordionIcon,
} from '@chakra-ui/react';
import { EditIcon, DeleteIcon, AddIcon } from '@chakra-ui/icons';
import { useAuth } from '../../auth-context';

// --- Interfaces ---
interface Level {
  id: number;
  nama: string;
}

interface Subject {
  id: number;
  nama: string;
}

interface Topic {
  id: number;
  mataPelajaran: { id: number; nama: string };
  tingkat: { id: number; nama: string };
  nama: string;
}

interface Question {
  id: number;
  materi: {
    id: number;
    mataPelajaran: { id: number; nama: string };
    tingkat: { id: number; nama: string };
    nama: string;
  };
  pertanyaan: string;
  opsiA: string;
  opsiB: string;
  opsiC: string;
  opsiD: string;
  jawabanBenar: string;
  pembahasan?: string;
  gambar: {
    id: number;
    namaFile: string;
    filePath: string;
    fileSize: number;
    mimeType: string;
    urutan: number;
    keterangan: string;
    createdAt: string;
  }[];
}

const mapLetterToEnum = (val: string) => {
  switch ((val || '').trim().toUpperCase()) {
    case 'A': return 'A';
    case 'B': return 'B';
    case 'C': return 'C';
    case 'D': return 'D';
    default: return 'JAWABAN_INVALID';
  }
};

const mapEnumToLetter = (val: string | number) => {
  const raw = `${val || ''}`.trim().toUpperCase();
  if (raw === 'A' || raw === '1') return 'A';
  if (raw === 'B' || raw === '2') return 'B';
  if (raw === 'C' || raw === '3') return 'C';
  if (raw === 'D' || raw === '4') return 'D';
  if (raw.endsWith('_A')) return 'A';
  if (raw.endsWith('_B')) return 'B';
  if (raw.endsWith('_C')) return 'C';
  if (raw.endsWith('_D')) return 'D';
  return 'A';
};

// --- API Helpers ---
const API_BASE = process.env.NEXT_PUBLIC_API_BASE + '/v1';

const steps = [
  { title: 'Pilih', description: 'Materi & Tingkat' },
  { title: 'Isi', description: 'Soal & Jawaban' },
  { title: 'Gambar', description: 'Upload Gambar' },
];

export default function QuestionsTab() {
  const toast = useToast();
  const { token } = useAuth();

  // --- API Helpers ---
  const authFetch = (url: string, options: RequestInit = {}) => {
    return fetch(url, {
      ...options,
      headers: {
        ...options.headers,
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    });
  };

  // --- State ---
  const [levels, setLevels] = useState<Level[]>([]);
  const [subjects, setSubjects] = useState<Subject[]>([]);
  const [topics, setTopics] = useState<Topic[]>([]);
  const [questions, setQuestions] = useState<Question[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Filter state
  const [selectedLevel, setSelectedLevel] = useState<string>('');
  const [selectedSubject, setSelectedSubject] = useState<string>('');
  const [searchQuery, setSearchQuery] = useState<string>('');
  
  // Debounced filter values to prevent excessive re-renders
  const [debouncedSearchQuery, setDebouncedSearchQuery] = useState<string>('');

  // Pagination for levels (if >5 levels)
  const levelsPerPage = 5;
  const [currentLevelPage, setCurrentLevelPage] = useState(1);

  // Multi-step modal state
  const { isOpen: isQuestionOpen, onOpen: onQuestionOpen, onClose: onQuestionClose } = useDisclosure();
  const [currentStep, setCurrentStep] = useState(0);
  const [currentQuestion, setCurrentQuestion] = useState<Partial<Question>>({});
  const [selectedFiles, setSelectedFiles] = useState<FileList | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  
  // Local form state to avoid re-render on every keystroke
  const [formValues, setFormValues] = useState({
    pertanyaan: '',
    opsiA: '',
    opsiB: '',
    opsiC: '',
    opsiD: '',
    jawabanBenar: 'A',
    pembahasan: '',
  });

  // Delete confirmation
  const { isOpen: isDeleteOpen, onOpen: onDeleteOpen, onClose: onDeleteClose } = useDisclosure();
  const [currentDeleteId, setCurrentDeleteId] = useState<number | null>(null);

  // --- Debounced Search Query Update ---
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearchQuery(searchQuery);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  // --- Direct Update for Form Inputs (No Debounce on State, Debounce on Save) ---
  const updateFormValue = useCallback((field: string, value: string) => {
    setFormValues(prev => ({ ...prev, [field]: value }));
  }, []);

  // --- Fetch Data ---
  const fetchLevels = async () => {
    try {
      const res = await authFetch(`${API_BASE}/levels`);
      if (res.ok) {
        const data = await res.json();
        setLevels(data.tingkat || []);
        setCurrentLevelPage(1);
      }
    } catch (error) {
      console.error('Failed to fetch levels', error);
    }
  };

  const fetchSubjects = async () => {
    try {
      const res = await authFetch(`${API_BASE}/subjects`);
      if (res.ok) {
        const data = await res.json();
        setSubjects(data.mataPelajaran || []);
      }
    } catch (error) {
      console.error('Failed to fetch subjects', error);
    }
  };

  const fetchTopics = async () => {
    try {
      const res = await authFetch(`${API_BASE}/topics`);
      if (res.ok) {
        const data = await res.json();
        setTopics(data.materi || []);
      }
    } catch (error) {
      console.error('Failed to fetch topics', error);
    }
  };

  const fetchQuestions = async () => {
    try {
      const res = await authFetch(`${API_BASE}/questions`);
      if (res.ok) {
        const data = await res.json();
        const normalized = (data.soal || []).map((q: Question) => ({
          ...q,
          jawabanBenar: mapEnumToLetter((q as any).jawabanBenar),
        }));
        setQuestions(normalized);
      }
    } catch (error) {
      console.error('Failed to fetch questions', error);
    }
  };

  useEffect(() => {
    const loadData = async () => {
      await Promise.all([fetchLevels(), fetchSubjects(), fetchTopics(), fetchQuestions()]);
      setIsLoading(false);
    };
    loadData();
  }, []);

  // --- Handlers ---

  const handleOpenNewQuestion = () => {
    setCurrentQuestion({
      pertanyaan: '',
      opsiA: '',
      opsiB: '',
      opsiC: '',
      opsiD: '',
      jawabanBenar: 'A',
      pembahasan: '',
    });
    setFormValues({
      pertanyaan: '',
      opsiA: '',
      opsiB: '',
      opsiC: '',
      opsiD: '',
      jawabanBenar: 'A',
      pembahasan: '',
    });
    setSelectedFiles(null);
    setCurrentStep(0);
    onQuestionOpen();
  };

  const handleEditQuestion = (question: Question) => {
    setCurrentQuestion(question);
    setFormValues({
      pertanyaan: question.pertanyaan || '',
      opsiA: question.opsiA || '',
      opsiB: question.opsiB || '',
      opsiC: question.opsiC || '',
      opsiD: question.opsiD || '',
      jawabanBenar: question.jawabanBenar || 'A',
      pembahasan: question.pembahasan || '',
    });
    setSelectedFiles(null);
    setCurrentStep(0);
    onQuestionOpen();
  };

  const handleNextStep = () => {
    // Validation for each step
    if (currentStep === 0) {
      if (!currentQuestion.materi?.id) {
        toast({ title: 'Pilih Materi terlebih dahulu', status: 'error' });
        return;
      }
    } else if (currentStep === 1) {
      if (
        !formValues.pertanyaan ||
        !formValues.opsiA ||
        !formValues.opsiB ||
        !formValues.opsiC ||
        !formValues.opsiD ||
        !formValues.jawabanBenar
      ) {
        toast({ title: 'Semua field harus diisi', status: 'error' });
        return;
      }
    }

    setCurrentStep(currentStep + 1);
  };

  const handlePrevStep = () => {
    setCurrentStep(currentStep - 1);
  };

  const handleSaveQuestion = async () => {
    // Merge form values into current question
    const mergedQuestion = {
      ...currentQuestion,
      ...formValues,
    };

    if (!mergedQuestion.pertanyaan || !mergedQuestion.opsiA || !mergedQuestion.opsiB || !mergedQuestion.opsiC || !mergedQuestion.opsiD || !mergedQuestion.jawabanBenar || !mergedQuestion.materi?.id) {
      toast({ title: 'Semua field harus diisi', status: 'error' });
      return;
    }

    const data: any = {
      idMateri: mergedQuestion.materi?.id,
      idTingkat: mergedQuestion.materi?.tingkat?.id,
      pertanyaan: mergedQuestion.pertanyaan,
      opsiA: mergedQuestion.opsiA,
      opsiB: mergedQuestion.opsiB,
      opsiC: mergedQuestion.opsiC,
      opsiD: mergedQuestion.opsiD,
      jawabanBenar: mapLetterToEnum(mergedQuestion.jawabanBenar || 'A'),
      pembahasan: mergedQuestion.pembahasan || '',
      imageBytes: [],
    };

    if (selectedFiles) {
      for (let i = 0; i < selectedFiles.length; i++) {
        const file = selectedFiles[i];
        const base64 = await new Promise<string>((resolve) => {
          const reader = new FileReader();
          reader.onload = () => resolve(reader.result as string);
          reader.readAsDataURL(file);
        });
        data.imageBytes.push(base64.split(',')[1]);
      }
    }

    const method = mergedQuestion.id ? 'PUT' : 'POST';
    const url = mergedQuestion.id ? `${API_BASE}/questions/${mergedQuestion.id}` : `${API_BASE}/questions`;

    try {
      const res = await authFetch(url, {
        method,
        body: JSON.stringify(data),
      });

      if (res.ok) {
        toast({ title: 'Soal berhasil disimpan', status: 'success' });
        fetchQuestions();
        onQuestionClose();
        setCurrentStep(0);
        setSelectedFiles(null);
        setFormValues({
          pertanyaan: '',
          opsiA: '',
          opsiB: '',
          opsiC: '',
          opsiD: '',
          jawabanBenar: 'A',
          pembahasan: '',
        });
      } else {
        const errorText = await res.text();
        toast({ title: `Gagal menyimpan soal: ${errorText}`, status: 'error' });
      }
    } catch (e) {
      toast({ title: 'Error menyimpan soal', status: 'error' });
    }
  };

  const handleDeleteQuestion = async () => {
    if (!currentDeleteId) return;
    try {
      const res = await authFetch(`${API_BASE}/questions/${currentDeleteId}`, { method: 'DELETE' });
      if (res.ok) {
        toast({ title: 'Soal berhasil dihapus', status: 'success' });
        fetchQuestions();
        onDeleteClose();
        setCurrentDeleteId(null);
      } else {
        toast({ title: 'Gagal menghapus soal', status: 'error' });
      }
    } catch (e) {
      toast({ title: 'Error menghapus soal', status: 'error' });
    }
  };

  const handleDeleteImage = async (questionId: number, imageId: number) => {
    if (!confirm('Hapus gambar ini?')) return;
    try {
      const res = await authFetch(`${API_BASE}/questions/${questionId}/images/${imageId}`, {
        method: 'DELETE',
      });
      if (res.ok) {
        toast({ title: 'Gambar berhasil dihapus', status: 'success' });
        fetchQuestions();
        if (currentQuestion && currentQuestion.id === questionId) {
          const updatedImages = currentQuestion.gambar?.filter((img) => img.id !== imageId);
          setCurrentQuestion({ ...currentQuestion, gambar: updatedImages });
        }
      } else {
        toast({ title: 'Gagal menghapus gambar', status: 'error' });
      }
    } catch (e) {
      toast({ title: 'Error menghapus gambar', status: 'error' });
    }
  };

  // --- Render Helpers ---

  const filteredQuestions = useMemo(() => {
    return questions.filter((q) => {
      const matchesLevel = selectedLevel === '' || q.materi.tingkat.id.toString() === selectedLevel;
      const matchesSubject = selectedSubject === '' || q.materi.mataPelajaran.id.toString() === selectedSubject;
      const matchesSearch = debouncedSearchQuery === '' ||
        q.materi.nama.toLowerCase().includes(debouncedSearchQuery.toLowerCase()) ||
        q.materi.mataPelajaran.nama.toLowerCase().includes(debouncedSearchQuery.toLowerCase());
      return matchesLevel && matchesSubject && matchesSearch;
    });
  }, [questions, selectedLevel, selectedSubject, debouncedSearchQuery]);

  const nestedData = useMemo(() => {
    return filteredQuestions.reduce(
      (acc, q) => {
        const levelId = q.materi.tingkat.id;
        const subjectId = q.materi.mataPelajaran.id;
        const topicId = q.materi.id;

        if (!acc[levelId]) {
          acc[levelId] = {
            level: q.materi.tingkat,
            subjects: {},
          };
        }

        if (!acc[levelId].subjects[subjectId]) {
          acc[levelId].subjects[subjectId] = {
            subject: q.materi.mataPelajaran,
            topics: {},
          };
        }

        if (!acc[levelId].subjects[subjectId].topics[topicId]) {
          acc[levelId].subjects[subjectId].topics[topicId] = {
            topic: q.materi,
            questions: [],
          };
        }

        acc[levelId].subjects[subjectId].topics[topicId].questions.push(q);
        return acc;
      },
      {} as Record<number, {
        level: { id: number; nama: string };
        subjects: Record<number, {
          subject: { id: number; nama: string };
          topics: Record<number, {
            topic: { id: number; nama: string; mataPelajaran: { id: number; nama: string }; tingkat: { id: number; nama: string } };
            questions: Question[];
          }>;
        }>;
      }>
    );
  }, [filteredQuestions]);

  const renderContent = useMemo(() => {
    if (isLoading) return <Text>Loading...</Text>;

    const totalLevelPages = Math.ceil(levels.length / levelsPerPage);
    const displayedLevels = levels.slice((currentLevelPage - 1) * levelsPerPage, currentLevelPage * levelsPerPage);
    const filteredNestedData = Object.values(nestedData).filter(levelData =>
      displayedLevels.some(level => level.id === levelData.level.id)
    );

    return (
      <>
        {levels.length > levelsPerPage && (
          <HStack justify="center" mb={4}>
            <Button
              onClick={() => setCurrentLevelPage(Math.max(1, currentLevelPage - 1))}
              isDisabled={currentLevelPage === 1}
            >
              Previous
            </Button>
            <Text>
              Halaman {currentLevelPage} dari {totalLevelPages}
            </Text>
            <Button
              onClick={() => setCurrentLevelPage(Math.min(totalLevelPages, currentLevelPage + 1))}
              isDisabled={currentLevelPage === totalLevelPages}
            >
              Next
            </Button>
          </HStack>
        )}
        {filteredNestedData.map((levelData) => (
          <Accordion allowToggle key={levelData.level.id} mb={8}>
            <AccordionItem>
              <AccordionButton>
                <Box flex="1" textAlign="left">
                  <Heading size="md">{levelData.level.nama}</Heading>
                </Box>
                <AccordionIcon />
              </AccordionButton>
              <AccordionPanel pb={4}>
                <Accordion allowToggle>
                  {Object.values(levelData.subjects).map((subjectData) => (
                    <AccordionItem key={subjectData.subject.id}>
                      <AccordionButton>
                        <Box flex="1" textAlign="left">
                          <Text fontWeight="bold">{subjectData.subject.nama}</Text>
                        </Box>
                        <AccordionIcon />
                      </AccordionButton>
                      <AccordionPanel pb={4}>
                        <Accordion allowToggle>
                          {Object.values(subjectData.topics).map((topicData) => (
                            <AccordionItem key={topicData.topic.id}>
                              <AccordionButton>
                                <Box flex="1" textAlign="left">
                                  <Text fontWeight="semibold">{topicData.topic.nama}</Text>
                                </Box>
                                <AccordionIcon />
                              </AccordionButton>
                              <AccordionPanel pb={4}>
                                <Table variant="simple" size="sm">
                                  <Thead>
                                    <Tr>
                                      <Th>Soal</Th>
                                      <Th>Jawaban Benar</Th>
                                      <Th>Gambar</Th>
                                      <Th>Aksi</Th>
                                    </Tr>
                                  </Thead>
                                  <Tbody>
                                    {topicData.questions.map((q) => (
                                      <Tr key={q.id}>
                                        <Td maxW="300px" isTruncated>
                                          {q.pertanyaan}
                                        </Td>
                                        <Td>
                                          <Badge colorScheme="green">{q.jawabanBenar}</Badge>
                                        </Td>
                                        <Td>
                                          <Badge colorScheme="blue">{q.gambar.length}</Badge>
                                        </Td>
                                        <Td>
                                          <IconButton
                                            aria-label="Edit"
                                            icon={<EditIcon />}
                                            size="sm"
                                            mr={2}
                                            onClick={() => handleEditQuestion(q)}
                                          />
                                          <IconButton
                                            aria-label="Delete"
                                            icon={<DeleteIcon />}
                                            size="sm"
                                            colorScheme="red"
                                            onClick={() => {
                                              setCurrentDeleteId(q.id);
                                              onDeleteOpen();
                                            }}
                                          />
                                        </Td>
                                      </Tr>
                                    ))}
                                  </Tbody>
                                </Table>
                              </AccordionPanel>
                            </AccordionItem>
                          ))}
                        </Accordion>
                      </AccordionPanel>
                    </AccordionItem>
                  ))}
                </Accordion>
              </AccordionPanel>
            </AccordionItem>
          </Accordion>
        ))}
      </>
    );
  }, [nestedData, levels, currentLevelPage, isLoading]);

  return (
    <Box>
      <Button leftIcon={<AddIcon />} colorScheme="blue" onClick={handleOpenNewQuestion} mb={4}>
        Tambah Soal
      </Button>

      {/* Filter Section */}
      <HStack spacing={4} mb={6}>
        <FormControl>
          <FormLabel>Filter Tingkat</FormLabel>
          <Select
            placeholder="Semua Tingkat"
            value={selectedLevel}
            onChange={(e) => setSelectedLevel(e.target.value)}
          >
            {levels.map((level) => (
              <option key={level.id} value={level.id.toString()}>
                {level.nama}
              </option>
            ))}
          </Select>
        </FormControl>
        <FormControl>
          <FormLabel>Filter Mata Pelajaran</FormLabel>
          <Select
            placeholder="Semua Mata Pelajaran"
            value={selectedSubject}
            onChange={(e) => setSelectedSubject(e.target.value)}
          >
            {subjects.map((subject) => (
              <option key={subject.id} value={subject.id.toString()}>
                {subject.nama}
              </option>
            ))}
          </Select>
        </FormControl>
        <FormControl>
          <FormLabel>Pencarian</FormLabel>
          <Input
            placeholder="Cari berdasarkan materi atau mata pelajaran"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </FormControl>
      </HStack>

      {renderContent}

      {/* Multi-Step Question Modal */}
      <Modal isOpen={isQuestionOpen} onClose={onQuestionClose} size="2xl" closeOnEsc={false} closeOnOverlayClick={false}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>{currentQuestion.id ? 'Edit Soal' : 'Tambah Soal Baru'}</ModalHeader>
          <ModalCloseButton isDisabled={currentStep > 0} />

          <Stepper size="sm" index={currentStep} mb={6} colorScheme="blue" px={6} pt={4}>
            {steps.map((step, index) => (
              <Step key={index}>
                <StepIndicator>
                  <StepStatus complete={`✓`} incomplete={index + 1} active={index + 1} />
                </StepIndicator>
                <Box flexShrink="0">
                  <StepTitle fontSize="sm">{step.title}</StepTitle>
                  <Text fontSize="xs" color="gray.500">{step.description}</Text>
                </Box>
                <StepSeparator />
              </Step>
            ))}
          </Stepper>

          <ModalBody pb={6}>
            {/* STEP 1: Select Topic & Level */}
            {currentStep === 0 && (
              <VStack spacing={4} align="stretch">
                <Box>
                  <Heading size="sm" mb={4}>
                    Langkah 1: Pilih Materi & Tingkatan
                  </Heading>
                  <Divider mb={4} />
                </Box>

                <FormControl isRequired>
                  <FormLabel fontWeight="bold">Pilih Materi</FormLabel>
                  <Select
                    placeholder="-- Pilih Materi --"
                    value={currentQuestion.materi?.id || ''}
                    onChange={(e) => {
                      const topicId = parseInt(e.target.value);
                      const topic = topics.find((t) => t.id === topicId);
                      setCurrentQuestion({ ...currentQuestion, materi: topic });
                    }}
                    size="lg"
                    focusBorderColor="blue.400"
                  >
                    {topics.map((t) => (
                      <option key={t.id} value={t.id}>
                        {t.nama} - {t.mataPelajaran.nama} ({t.tingkat.nama})
                      </option>
                    ))}
                  </Select>
                </FormControl>

                {currentQuestion.materi && (
                  <Box bg="blue.50" p={4} borderRadius="md" borderLeft="4px solid" borderLeftColor="blue.400">
                    <VStack align="start" spacing={2}>
                      <Text>
                        <strong>Mata Pelajaran:</strong> {currentQuestion.materi.mataPelajaran.nama}
                      </Text>
                      <Text>
                        <strong>Tingkatan:</strong> <Badge colorScheme="blue">{currentQuestion.materi.tingkat.nama}</Badge>
                      </Text>
                      <Text>
                        <strong>Materi:</strong> {currentQuestion.materi.nama}
                      </Text>
                    </VStack>
                  </Box>
                )}
              </VStack>
            )}

            {/* STEP 2: Question & Answers */}
            {currentStep === 1 && (
              <VStack spacing={4} align="stretch">
                <Box>
                  <Heading size="sm" mb={4}>
                    Langkah 2: Soal, Opsi Jawaban & Pembahasan
                  </Heading>
                  <Divider mb={4} />
                </Box>

                <FormControl isRequired>
                  <FormLabel fontWeight="bold">Soal/Pertanyaan</FormLabel>
                  <Textarea
                    value={formValues.pertanyaan}
                    onChange={(e) => updateFormValue('pertanyaan', e.target.value)}
                    placeholder="Masukkan pertanyaan soal di sini..."
                    rows={4}
                    size="lg"
                    focusBorderColor="blue.400"
                  />
                </FormControl>

                <Divider my={2} />

                <FormControl isRequired>
                  <FormLabel fontWeight="bold">A. Opsi A</FormLabel>
                  <Input
                    value={formValues.opsiA}
                    onChange={(e) => updateFormValue('opsiA', e.target.value)}
                    placeholder="Masukkan pilihan A..."
                    size="lg"
                    focusBorderColor="blue.400"
                  />
                </FormControl>

                <FormControl isRequired>
                  <FormLabel fontWeight="bold">B. Opsi B</FormLabel>
                  <Input
                    value={formValues.opsiB}
                    onChange={(e) => updateFormValue('opsiB', e.target.value)}
                    placeholder="Masukkan pilihan B..."
                    size="lg"
                    focusBorderColor="blue.400"
                  />
                </FormControl>

                <FormControl isRequired>
                  <FormLabel fontWeight="bold">C. Opsi C</FormLabel>
                  <Input
                    value={formValues.opsiC}
                    onChange={(e) => updateFormValue('opsiC', e.target.value)}
                    placeholder="Masukkan pilihan C..."
                    size="lg"
                    focusBorderColor="blue.400"
                  />
                </FormControl>

                <FormControl isRequired>
                  <FormLabel fontWeight="bold">D. Opsi D</FormLabel>
                  <Input
                    value={formValues.opsiD}
                    onChange={(e) => updateFormValue('opsiD', e.target.value)}
                    placeholder="Masukkan pilihan D..."
                    size="lg"
                    focusBorderColor="blue.400"
                  />
                </FormControl>

                <Divider my={2} />

                <FormControl isRequired>
                  <FormLabel fontWeight="bold">Jawaban Benar</FormLabel>
                  <Select
                    value={formValues.jawabanBenar || 'A'}
                    onChange={(e) => updateFormValue('jawabanBenar', e.target.value)}
                    size="lg"
                    focusBorderColor="green.400"
                    bg="green.50"
                  >
                    <option value="A">A</option>
                    <option value="B">B</option>
                    <option value="C">C</option>
                    <option value="D">D</option>
                  </Select>
                </FormControl>

                <FormControl>
                  <FormLabel fontWeight="bold">Pembahasan (Opsional)</FormLabel>
                  <Textarea
                    value={formValues.pembahasan || ''}
                    onChange={(e) => updateFormValue('pembahasan', e.target.value)}
                    placeholder="Masukkan penjelasan untuk membantu siswa memahami jawaban yang benar..."
                    rows={3}
                    size="lg"
                    focusBorderColor="blue.400"
                  />
                  <Text fontSize="xs" color="gray.500" mt={1}>
                    Pembahasan membantu siswa belajar lebih dalam
                  </Text>
                </FormControl>
              </VStack>
            )}

            {/* STEP 3: Images */}
            {currentStep === 2 && (
              <VStack spacing={4} align="start" w="100%">
                <Box w="100%">
                  <Heading size="sm" mb={4}>
                    Langkah 3: Upload Gambar (Opsional)
                  </Heading>
                  <Divider mb={4} />
                </Box>

                {/* Existing Images */}
                {currentQuestion.gambar && currentQuestion.gambar.length > 0 && (
                  <Box w="100%">
                    <Text fontWeight="bold" mb={3}>
                      Gambar yang Ada ({currentQuestion.gambar.length})
                    </Text>
                    <SimpleGrid columns={[2, 3, 4]} spacing={3}>
                      {currentQuestion.gambar.map((img) => (
                        <Box
                          key={img.id}
                          position="relative"
                          borderWidth="1px"
                          borderRadius="md"
                          overflow="hidden"
                          _hover={{ shadow: 'md' }}
                          bg="gray.50"
                        >
                          <ChakraImage
                            src={`${process.env.NEXT_PUBLIC_API_BASE}/${img.filePath.replace(/\\/g, '/')}`}
                            alt="Question Image"
                            boxSize="100px"
                            objectFit="cover"
                          />
                          <IconButton
                            aria-label="Delete Image"
                            icon={<DeleteIcon />}
                            size="xs"
                            colorScheme="red"
                            position="absolute"
                            top={1}
                            right={1}
                            onClick={() => handleDeleteImage(currentQuestion.id!, img.id)}
                          />
                        </Box>
                      ))}
                    </SimpleGrid>
                    <Divider my={4} />
                  </Box>
                )}

                {/* Upload New Images */}
                <FormControl w="100%">
                  <FormLabel fontWeight="bold">Upload Gambar Baru</FormLabel>
                  <Input
                    type="file"
                    multiple
                    accept="image/*"
                    ref={fileInputRef}
                    onChange={(e) => setSelectedFiles(e.target.files)}
                    size="lg"
                    focusBorderColor="blue.400"
                  />
                  <Text fontSize="xs" color="gray.500" mt={2}>
                    Format: JPG, PNG, GIF. Maksimal 5MB per file.
                  </Text>
                </FormControl>

                {selectedFiles && selectedFiles.length > 0 && (
                  <Box bg="green.50" p={3} borderRadius="md" w="100%" borderLeft="4px solid" borderLeftColor="green.400">
                    <Text fontWeight="bold" color="green.700">
                      {selectedFiles.length} file siap di-upload
                    </Text>
                  </Box>
                )}

                <Box bg="gray.50" p={3} borderRadius="md" w="100%">
                  <Text fontSize="sm" color="gray.600">
                    Gambar opsional. Klik &quot;Simpan Soal&quot; untuk menyelesaikan pembuatan soal.
                  </Text>
                </Box>
              </VStack>
            )}
          </ModalBody>

          <ModalFooter>
            <HStack spacing={2}>
              {currentStep > 0 && (
                <Button variant="outline" onClick={handlePrevStep}>
                  ← Kembali
                </Button>
              )}

              {currentStep < steps.length - 1 && (
                <Button colorScheme="blue" onClick={handleNextStep}>
                  Lanjut →
                </Button>
              )}

              {currentStep === steps.length - 1 && (
                <Button colorScheme="green" onClick={handleSaveQuestion}>
                  Simpan Soal
                </Button>
              )}

              <Button
                variant="ghost"
                onClick={() => {
                  onQuestionClose();
                  setCurrentStep(0);
                  setSelectedFiles(null);
                }}
              >
                Batal
              </Button>
            </HStack>
          </ModalFooter>
        </ModalContent>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal isOpen={isDeleteOpen} onClose={onDeleteClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Konfirmasi Hapus</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <Text>Apakah Anda yakin ingin menghapus soal ini? Tindakan ini tidak dapat dibatalkan.</Text>
          </ModalBody>
          <ModalFooter>
            <Button colorScheme="red" mr={3} onClick={handleDeleteQuestion}>
              Ya, Hapus
            </Button>
            <Button onClick={onDeleteClose}>Batal</Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </Box>
  );
}
