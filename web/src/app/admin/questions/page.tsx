'use client';

import Link from 'next/link';
import { useState, useEffect } from 'react';
import {
  Box,
  Button,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
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
  useDisclosure,
  VStack,
  Heading,
  Container,
  useToast,
} from '@chakra-ui/react';
import axios from 'axios';

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
}

interface Topic {
  id: number;
  nama: string;
}

interface Level {
  id: number;
  nama: string;
}

const API_BASE = 'http://localhost:8080/v1/questions';
const TOPICS_API = 'http://localhost:8080/v1/topics';
const LEVELS_API = 'http://localhost:8080/v1/levels';

export default function QuestionsPage() {
  const [questions, setQuestions] = useState<Question[]>([]);
  const [topics, setTopics] = useState<Topic[]>([]);
  const [levels, setLevels] = useState<Level[]>([]);
  const [editingQuestion, setEditingQuestion] = useState<Question | null>(null);
  const [formData, setFormData] = useState({
    idMateri: '',
    idTingkat: '',
    pertanyaan: '',
    opsiA: '',
    opsiB: '',
    opsiC: '',
    opsiD: '',
    jawabanBenar: '',
  });
  const { isOpen, onOpen, onClose } = useDisclosure();
  const toast = useToast();

  useEffect(() => {
    fetchQuestions();
    fetchTopics();
    fetchLevels();
  }, []);

  const fetchQuestions = async () => {
    try {
      const response = await axios.get(API_BASE);
      const data = response.data;
      setQuestions(
        Array.isArray(data) ? data :
        Array.isArray(data.data) ? data.data :
        Array.isArray(data.soal) ? data.soal : []
      );
    } catch (error) {
      toast({ title: 'Error fetching questions', status: 'error' });
      setQuestions([]);
    }
  };

  const fetchTopics = async () => {
    try {
      const response = await axios.get(TOPICS_API);
      const data = response.data;
      setTopics(
        Array.isArray(data) ? data :
        Array.isArray(data.data) ? data.data :
        Array.isArray(data.materi) ? data.materi : []
      );
    } catch (error) {
      toast({ title: 'Error fetching topics', status: 'error' });
      setTopics([]);
    }
  };

  const fetchLevels = async () => {
    try {
      const response = await axios.get(LEVELS_API);
      const data = response.data;
      setLevels(
        Array.isArray(data) ? data :
        Array.isArray(data.data) ? data.data :
        Array.isArray(data.tingkat) ? data.tingkat : []
      );
    } catch (error) {
      toast({ title: 'Error fetching levels', status: 'error' });
      setLevels([]);
    }
  };

  const handleCreate = () => {
    setEditingQuestion(null);
    setFormData({
      idMateri: '',
      idTingkat: '',
      pertanyaan: '',
      opsiA: '',
      opsiB: '',
      opsiC: '',
      opsiD: '',
      jawabanBenar: '',
    });
    onOpen();
  };

  const handleEdit = (question: Question) => {
    setEditingQuestion(question);
    setFormData({
      idMateri: question.materi.id.toString(),
      idTingkat: question.materi.tingkat.id.toString(),
      pertanyaan: question.pertanyaan,
      opsiA: question.opsiA,
      opsiB: question.opsiB,
      opsiC: question.opsiC,
      opsiD: question.opsiD,
      jawabanBenar: question.jawabanBenar,
    });
    onOpen();
  };

  const handleDelete = async (id: number) => {
    try {
      await axios.delete(`${API_BASE}/${id}`);
      fetchQuestions();
      toast({ title: 'Question deleted', status: 'success' });
    } catch (error) {
      toast({ title: 'Error deleting question', status: 'error' });
    }
  };

  const handleSubmit = async () => {
    const data = {
      idMateri: parseInt(formData.idMateri),
      idTingkat: parseInt(formData.idTingkat),
      pertanyaan: formData.pertanyaan,
      opsiA: formData.opsiA,
      opsiB: formData.opsiB,
      opsiC: formData.opsiC,
      opsiD: formData.opsiD,
      jawabanBenar: formData.jawabanBenar,
    };
    try {
      if (editingQuestion) {
        await axios.put(`${API_BASE}/${editingQuestion.id}`, data);
        toast({ title: 'Question updated', status: 'success' });
      } else {
        await axios.post(API_BASE, data);
        toast({ title: 'Question created', status: 'success' });
      }
      fetchQuestions();
      onClose();
    } catch (error) {
      toast({ title: 'Error saving question', status: 'error' });
    }
  };

  const getTopicName = (id: number) => topics.find(t => t.id === id)?.nama || 'Unknown';
  const getLevelName = (id: number) => levels.find(l => l.id === id)?.nama || 'Unknown';

  const groupedQuestions = questions.reduce((acc, q) => {
    const key = `${q.materi.id}-${q.materi.tingkat.id}`;
    if (!acc[key]) {
      acc[key] = {
        topic: q.materi,
        level: q.materi.tingkat,
        questions: []
      };
    }
    acc[key].questions.push(q);
    return acc;
  }, {} as Record<string, { topic: any; level: any; questions: Question[] }>);

  return (
    <Container maxW="container.xl" py={10}>
      <Link href="/">
        <Button mb={4} variant="outline">
          Back to Home
        </Button>
      </Link>
      <Heading as="h1" size="xl" mb={8}>
        Manage Questions
      </Heading>
      <Button colorScheme="orange" onClick={handleCreate} mb={4}>
        Add Question
      </Button>
      <VStack spacing={8} align="stretch">
        {Object.values(groupedQuestions).map((group) => (
          <Box key={`${group.topic.id}-${group.level.id}`}>
            <Heading size="md" mb={4}>
              Topic: {group.topic.nama} (Level: {group.level.nama})
            </Heading>
            <Table variant="simple">
              <Thead>
                <Tr>
                  <Th>ID</Th>
                  <Th>Question</Th>
                  <Th>Correct Answer</Th>
                  <Th>Actions</Th>
                </Tr>
              </Thead>
              <Tbody>
                {group.questions.map((question) => (
                  <Tr key={question.id}>
                    <Td>{question.id}</Td>
                    <Td>{question.pertanyaan.substring(0, 50)}...</Td>
                    <Td>{question.jawabanBenar}</Td>
                    <Td>
                      <Button size="sm" mr={2} onClick={() => handleEdit(question)}>
                        Edit
                      </Button>
                      <Button size="sm" colorScheme="red" onClick={() => handleDelete(question.id)}>
                        Delete
                      </Button>
                    </Td>
                  </Tr>
                ))}
              </Tbody>
            </Table>
          </Box>
        ))}
      </VStack>

      <Modal isOpen={isOpen} onClose={onClose} size="xl">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>{editingQuestion ? 'Edit Question' : 'Add Question'}</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4}>
              <FormControl>
                <FormLabel>Topic</FormLabel>
                <Select
                  value={formData.idMateri}
                  onChange={(e) => setFormData({ ...formData, idMateri: e.target.value })}
                >
                  <option value="">Select Topic</option>
                  {topics.map((topic) => (
                    <option key={topic.id} value={topic.id.toString()}>
                      {topic.nama}
                    </option>
                  ))}
                </Select>
              </FormControl>
              <FormControl>
                <FormLabel>Level</FormLabel>
                <Select
                  value={formData.idTingkat}
                  onChange={(e) => setFormData({ ...formData, idTingkat: e.target.value })}
                >
                  <option value="">Select Level</option>
                  {levels.map((level) => (
                    <option key={level.id} value={level.id.toString()}>
                      {level.nama}
                    </option>
                  ))}
                </Select>
              </FormControl>
              <FormControl>
                <FormLabel>Question</FormLabel>
                <Textarea
                  value={formData.pertanyaan}
                  onChange={(e) => setFormData({ ...formData, pertanyaan: e.target.value })}
                />
              </FormControl>
              <FormControl>
                <FormLabel>Option A</FormLabel>
                <Input
                  value={formData.opsiA}
                  onChange={(e) => setFormData({ ...formData, opsiA: e.target.value })}
                />
              </FormControl>
              <FormControl>
                <FormLabel>Option B</FormLabel>
                <Input
                  value={formData.opsiB}
                  onChange={(e) => setFormData({ ...formData, opsiB: e.target.value })}
                />
              </FormControl>
              <FormControl>
                <FormLabel>Option C</FormLabel>
                <Input
                  value={formData.opsiC}
                  onChange={(e) => setFormData({ ...formData, opsiC: e.target.value })}
                />
              </FormControl>
              <FormControl>
                <FormLabel>Option D</FormLabel>
                <Input
                  value={formData.opsiD}
                  onChange={(e) => setFormData({ ...formData, opsiD: e.target.value })}
                />
              </FormControl>
              <FormControl>
                <FormLabel>Correct Answer</FormLabel>
                <Select
                  value={formData.jawabanBenar}
                  onChange={(e) => setFormData({ ...formData, jawabanBenar: e.target.value })}
                >
                  <option value="">Select Correct Answer</option>
                  <option value="A">A</option>
                  <option value="B">B</option>
                  <option value="C">C</option>
                  <option value="D">D</option>
                </Select>
              </FormControl>
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button colorScheme="orange" mr={3} onClick={handleSubmit}>
              Save
            </Button>
            <Button variant="ghost" onClick={onClose}>
              Cancel
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </Container>
  );
}