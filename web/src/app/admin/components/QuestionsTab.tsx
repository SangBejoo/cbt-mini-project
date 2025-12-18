'use client';

import { useState, useEffect, useRef } from 'react';
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
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  Card,
  CardBody,
  CardHeader,
  Heading,
  SimpleGrid,
  Flex,
  Spacer
} from '@chakra-ui/react';
import { EditIcon, DeleteIcon, AddIcon, AttachmentIcon } from '@chakra-ui/icons';

// --- Interfaces ---

interface Level {
  id: number;
  name: string;
}

interface Subject {
  id: number;
  name: string;
  level_id: number;
  level_name?: string;
}

interface Topic {
  id: number;
  mataPelajaran: { id: number; nama: string };
  tingkat: { id: number; nama: string };
  nama: string;
}

interface QuestionImage {
  id: number;
  image_url: string;
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

// --- API Helpers ---
const API_BASE = 'http://localhost:8080/v1';

export default function QuestionsTab() {
  const toast = useToast();

  // --- State ---
  const [levels, setLevels] = useState<Level[]>([]);
  const [subjects, setSubjects] = useState<Subject[]>([]);
  const [topics, setTopics] = useState<Topic[]>([]);
  const [questions, setQuestions] = useState<Question[]>([]);

  // Loading states
  const [isLoading, setIsLoading] = useState(false);

  // --- Fetch Data ---
  const fetchLevels = async () => {
    try {
      const res = await fetch(`${API_BASE}/levels`);
      if (res.ok) {
        const data = await res.json();
        setLevels(data.tingkat || []);
      }
    } catch (error) {
      console.error("Failed to fetch levels", error);
    }
  };

  const fetchSubjects = async () => {
    try {
      const res = await fetch(`${API_BASE}/subjects`);
      if (res.ok) {
        const data = await res.json();
        setSubjects(data.mataPelajaran || []);
      }
    } catch (error) {
      console.error("Failed to fetch subjects", error);
    }
  };

  const fetchTopics = async () => {
    try {
      const res = await fetch(`${API_BASE}/topics`);
      if (res.ok) {
        const data = await res.json();
        setTopics(data.materi || []);
      }
    } catch (error) {
      console.error("Failed to fetch topics", error);
    }
  };

  const fetchQuestions = async () => {
    try {
      const res = await fetch(`${API_BASE}/questions`);
      if (res.ok) {
        const data = await res.json();
        setQuestions(data.soal || []);
      }
    } catch (error) {
      console.error("Failed to fetch questions", error);
    }
  };

  useEffect(() => {
    fetchLevels();
    fetchSubjects();
    fetchTopics();
    fetchQuestions();
  }, []);

  // --- Modals & Form State ---
  
  // Level Modal
  const { isOpen: isLevelOpen, onOpen: onLevelOpen, onClose: onLevelClose } = useDisclosure();
  const [currentLevel, setCurrentLevel] = useState<Partial<Level>>({});
  
  // Subject Modal
  const { isOpen: isSubjectOpen, onOpen: onSubjectOpen, onClose: onSubjectClose } = useDisclosure();
  const [currentSubject, setCurrentSubject] = useState<Partial<Subject>>({});

  // Topic Modal
  const { isOpen: isTopicOpen, onOpen: onTopicOpen, onClose: onTopicClose } = useDisclosure();
  const [currentTopic, setCurrentTopic] = useState<Partial<Topic>>({});

  // Question Modal
  const { isOpen: isQuestionOpen, onOpen: onQuestionOpen, onClose: onQuestionClose } = useDisclosure();
  const [currentQuestion, setCurrentQuestion] = useState<Partial<Question>>({});
  const [selectedFiles, setSelectedFiles] = useState<FileList | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Delete Confirmation Modal
  const { isOpen: isDeleteOpen, onOpen: onDeleteOpen, onClose: onDeleteClose } = useDisclosure();
  const [currentDeleteId, setCurrentDeleteId] = useState<number | null>(null);

  // --- Handlers: Level ---

  const handleSaveLevel = async () => {
    if (!currentLevel.name) {
      toast({ title: 'Name required', status: 'error' });
      return;
    }
    const method = currentLevel.id ? 'PUT' : 'POST';
    const url = currentLevel.id ? `${API_BASE}/levels/${currentLevel.id}` : `${API_BASE}/levels`;

    try {
      const res = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(currentLevel),
      });
      if (res.ok) {
        toast({ title: 'Level saved', status: 'success' });
        fetchLevels();
        onLevelClose();
      } else {
        toast({ title: 'Failed to save level', status: 'error' });
      }
    } catch (e) {
      toast({ title: 'Error saving level', status: 'error' });
    }
  };

  const handleDeleteLevel = async (id: number) => {
    if (!confirm('Are you sure? This might delete related subjects.')) return;
    try {
      const res = await fetch(`${API_BASE}/levels/${id}`, { method: 'DELETE' });
      if (res.ok) {
        toast({ title: 'Level deleted', status: 'success' });
        fetchLevels();
      } else {
        toast({ title: 'Failed to delete', status: 'error' });
      }
    } catch (e) {
      toast({ title: 'Error deleting', status: 'error' });
    }
  };

  // --- Handlers: Subject ---

  const handleSaveSubject = async () => {
    if (!currentSubject.name || !currentSubject.level_id) {
      toast({ title: 'Name and Level required', status: 'error' });
      return;
    }
    const method = currentSubject.id ? 'PUT' : 'POST';
    const url = currentSubject.id ? `${API_BASE}/subjects/${currentSubject.id}` : `${API_BASE}/subjects`;

    try {
      const res = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(currentSubject),
      });
      if (res.ok) {
        toast({ title: 'Subject saved', status: 'success' });
        fetchSubjects();
        onSubjectClose();
      } else {
        toast({ title: 'Failed to save subject', status: 'error' });
      }
    } catch (e) {
      toast({ title: 'Error saving subject', status: 'error' });
    }
  };

  const handleDeleteSubject = async (id: number) => {
    if (!confirm('Are you sure?')) return;
    try {
      const res = await fetch(`${API_BASE}/subjects/${id}`, { method: 'DELETE' });
      if (res.ok) {
        toast({ title: 'Subject deleted', status: 'success' });
        fetchSubjects();
      } else {
        toast({ title: 'Failed to delete', status: 'error' });
      }
    } catch (e) {
      toast({ title: 'Error deleting', status: 'error' });
    }
  };

  // --- Handlers: Topic ---

  const handleSaveTopic = async () => {
    if (!currentTopic.name || !currentTopic.subject_id) {
      toast({ title: 'Name and Subject required', status: 'error' });
      return;
    }
    const method = currentTopic.id ? 'PUT' : 'POST';
    const url = currentTopic.id ? `${API_BASE}/topics/${currentTopic.id}` : `${API_BASE}/topics`;

    try {
      const res = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(currentTopic),
      });
      if (res.ok) {
        toast({ title: 'Topic saved', status: 'success' });
        fetchTopics();
        onTopicClose();
      } else {
        toast({ title: 'Failed to save topic', status: 'error' });
      }
    } catch (e) {
      toast({ title: 'Error saving topic', status: 'error' });
    }
  };

  const handleDeleteTopic = async (id: number) => {
    if (!confirm('Are you sure?')) return;
    try {
      const res = await fetch(`${API_BASE}/topics/${id}`, { method: 'DELETE' });
      if (res.ok) {
        toast({ title: 'Topic deleted', status: 'success' });
        fetchTopics();
      } else {
        toast({ title: 'Failed to delete', status: 'error' });
      }
    } catch (e) {
      toast({ title: 'Error deleting', status: 'error' });
    }
  };

  // --- Handlers: Question ---

  const handleSaveQuestion = async () => {
    if (!currentQuestion.pertanyaan || !currentQuestion.opsiA || !currentQuestion.opsiB || !currentQuestion.opsiC || !currentQuestion.opsiD || !currentQuestion.jawabanBenar || !currentQuestion.materi?.id) {
      toast({ title: 'All fields required', status: 'error' });
      return;
    }

    const data: any = {
      idMateri: currentQuestion.materi?.id,
      idTingkat: currentQuestion.materi?.tingkat?.id,
      pertanyaan: currentQuestion.pertanyaan,
      opsiA: currentQuestion.opsiA,
      opsiB: currentQuestion.opsiB,
      opsiC: currentQuestion.opsiC,
      opsiD: currentQuestion.opsiD,
      jawabanBenar: currentQuestion.jawabanBenar,
      imageBytes: []
    };

    if (selectedFiles) {
      for (let i = 0; i < selectedFiles.length; i++) {
        const file = selectedFiles[i];
        const base64 = await new Promise<string>((resolve) => {
          const reader = new FileReader();
          reader.onload = () => resolve(reader.result as string);
          reader.readAsDataURL(file);
        });
        data.imageBytes.push(base64.split(',')[1]); // remove data:image/...;base64,
      }
    }

    const method = currentQuestion.id ? 'PUT' : 'POST';
    const url = currentQuestion.id ? `${API_BASE}/questions/${currentQuestion.id}` : `${API_BASE}/questions`;

    try {
      const res = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data), 
      });

      if (res.ok) {
        toast({ title: 'Question saved', status: 'success' });
        fetchQuestions();
        onQuestionClose();
        setSelectedFiles(null);
      } else {
        const errorText = await res.text();
        toast({ title: `Failed to save question: ${errorText}`, status: 'error' });
      }
    } catch (e) {
      toast({ title: 'Error saving question', status: 'error' });
    }
  };

  const handleDeleteQuestion = async () => {
    if (!currentDeleteId) return;
    try {
      const res = await fetch(`${API_BASE}/questions/${currentDeleteId}`, { method: 'DELETE' });
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
    if (!confirm('Delete this image?')) return;
    try {
      const res = await fetch(`${API_BASE}/questions/${questionId}/images/${imageId}`, {
        method: 'DELETE',
      });
      if (res.ok) {
        toast({ title: 'Image deleted', status: 'success' });
        // Refresh questions to update the list
        fetchQuestions();
        
        // Also update the current modal state if open
        if (currentQuestion && currentQuestion.id === questionId) {
           const updatedImages = currentQuestion.images?.filter(img => img.id !== imageId);
           setCurrentQuestion({ ...currentQuestion, images: updatedImages });
        }
      } else {
        toast({ title: 'Failed to delete image', status: 'error' });
      }
    } catch (e) {
      toast({ title: 'Error deleting image', status: 'error' });
    }
  };

  // --- Render Helpers ---

  const getLevelName = (id: number) => levels.find(l => l.id === id)?.name || id;
  const getSubjectName = (id: number) => subjects.find(s => s.id === id)?.name || id;
  const getTopicName = (id: number) => topics.find(t => t.id === id)?.nama || id;

  const groupedQuestions = questions.reduce((acc, q) => {
    const subjectId = q.materi.mataPelajaran.id;
    if (!acc[subjectId]) {
      acc[subjectId] = {
        subject: q.materi.mataPelajaran,
        questions: []
      };
    }
    acc[subjectId].questions.push(q);
    return acc;
  }, {} as Record<number, { subject: { id: number; nama: string }, questions: Question[] }>);

  return (
    <Box>
      <Button leftIcon={<AddIcon />} colorScheme="blue" onClick={() => { setCurrentQuestion({ pertanyaan: '', opsiA: '', opsiB: '', opsiC: '', opsiD: '', jawabanBenar: 'A' }); setSelectedFiles(null); onQuestionOpen(); }} mb={4}>
        Add Question
      </Button>
      {Object.values(groupedQuestions).map((group) => (
        <Box key={group.subject.id} mb={8}>
          <Heading size="md" mb={4}>{group.subject.nama}</Heading>
          <Table variant="simple">
            <Thead>
              <Tr>
                <Th>Question</Th>
                <Th>Topic</Th>
                <Th>Level</Th>
                <Th>Correct Answer</Th>
                <Th>Images</Th>
                <Th>Actions</Th>
              </Tr>
            </Thead>
            <Tbody>
              {group.questions.map((q) => (
                <Tr key={q.id}>
                  <Td maxW="300px" isTruncated>{q.pertanyaan}</Td>
                  <Td>{q.materi.nama}</Td>
                  <Td>{q.materi.tingkat.nama}</Td>
                  <Td>{q.jawabanBenar}</Td>
                  <Td>{q.gambar.length}</Td>
                  <Td>
                    <IconButton
                      aria-label="Edit"
                      icon={<EditIcon />}
                      size="sm"
                      mr={2}
                      onClick={() => { setCurrentQuestion(q); setSelectedFiles(null); onQuestionOpen(); }}
                    />
                    <IconButton
                      aria-label="Delete"
                      icon={<DeleteIcon />}
                      size="sm"
                      colorScheme="red"
                      onClick={() => { setCurrentDeleteId(q.id); onDeleteOpen(); }}
                    />
                  </Td>
                </Tr>
              ))}
            </Tbody>
          </Table>
        </Box>
      ))}

      {/* Question Modal */}
      <Modal isOpen={isQuestionOpen} onClose={onQuestionClose} size="xl">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>{currentQuestion.id ? 'Edit Question' : 'Add Question'}</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4}>
              <FormControl>
                <FormLabel>Topic</FormLabel>
                <Select
                  placeholder="Select Topic"
                  value={currentQuestion.materi?.id || ''}
                  onChange={(e) => {
                    const topicId = parseInt(e.target.value);
                    const topic = topics.find(t => t.id === topicId);
                    setCurrentQuestion({ ...currentQuestion, materi: topic });
                  }}
                >
                  {topics.map(t => <option key={t.id} value={t.id}>{t.nama} ({t.mataPelajaran.nama})</option>)}
                </Select>
              </FormControl>

              <FormControl>
                <FormLabel>Question</FormLabel>
                <Textarea
                  value={currentQuestion.pertanyaan || ''}
                  onChange={(e) => setCurrentQuestion({ ...currentQuestion, pertanyaan: e.target.value })}
                  rows={3}
                />
              </FormControl>

              <FormControl>
                <FormLabel>Option A</FormLabel>
                <Input
                  value={currentQuestion.opsiA || ''}
                  onChange={(e) => setCurrentQuestion({ ...currentQuestion, opsiA: e.target.value })}
                />
              </FormControl>

              <FormControl>
                <FormLabel>Option B</FormLabel>
                <Input
                  value={currentQuestion.opsiB || ''}
                  onChange={(e) => setCurrentQuestion({ ...currentQuestion, opsiB: e.target.value })}
                />
              </FormControl>

              <FormControl>
                <FormLabel>Option C</FormLabel>
                <Input
                  value={currentQuestion.opsiC || ''}
                  onChange={(e) => setCurrentQuestion({ ...currentQuestion, opsiC: e.target.value })}
                />
              </FormControl>

              <FormControl>
                <FormLabel>Option D</FormLabel>
                <Input
                  value={currentQuestion.opsiD || ''}
                  onChange={(e) => setCurrentQuestion({ ...currentQuestion, opsiD: e.target.value })}
                />
              </FormControl>

              <FormControl>
                <FormLabel>Correct Answer</FormLabel>
                <Select
                  value={currentQuestion.jawabanBenar || ''}
                  onChange={(e) => setCurrentQuestion({ ...currentQuestion, jawabanBenar: e.target.value })}
                >
                  <option value="A">A</option>
                  <option value="B">B</option>
                  <option value="C">C</option>
                  <option value="D">D</option>
                </Select>
              </FormControl>

              <FormControl>
                <FormLabel>Images</FormLabel>
                {/* Existing Images */}
                {currentQuestion.gambar && currentQuestion.gambar.length > 0 && (
                  <SimpleGrid columns={3} spacing={2} mb={3}>
                    {currentQuestion.gambar.map((img) => (
                      <Box key={img.id} position="relative" borderWidth="1px" borderRadius="md" overflow="hidden">
                        <ChakraImage src={`http://localhost:8080/${img.filePath.replace(/\\/g, '/')}`} alt="Question Image" boxSize="100px" objectFit="cover" />
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
                )}
                
                {/* Upload New Images */}
                <Input
                  type="file"
                  multiple
                  accept="image/*"
                  ref={fileInputRef}
                  onChange={(e) => setSelectedFiles(e.target.files)}
                  pt={1}
                />
                <Text fontSize="sm" color="gray.500" mt={1}>
                  Supported formats: JPG, PNG, GIF. Max 5MB.
                </Text>
              </FormControl>
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button colorScheme="blue" mr={3} onClick={handleSaveQuestion}>Save</Button>
            <Button onClick={onQuestionClose}>Cancel</Button>
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
            Apakah Anda yakin ingin menghapus soal ini?
          </ModalBody>
          <ModalFooter>
            <Button colorScheme="red" mr={3} onClick={handleDeleteQuestion}>Ya, Hapus</Button>
            <Button onClick={onDeleteClose}>Batal</Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </Box>
  );
}