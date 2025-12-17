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
  useDisclosure,
  VStack,
  Heading,
  Container,
  useToast,
} from '@chakra-ui/react';
import axios from 'axios';

interface Topic {
  id: number;
  mataPelajaran: { id: number; nama: string };
  tingkat: { id: number; nama: string };
  nama: string;
}

interface Level {
  id: number;
  nama: string;
}

interface Subject {
  id: number;
  nama: string;
}

const API_BASE = 'http://localhost:8080/v1/topics';
const LEVELS_API = 'http://localhost:8080/v1/levels';
const SUBJECTS_API = 'http://localhost:8080/v1/subjects';

export default function TopicsPage() {
  const [topics, setTopics] = useState<Topic[]>([]);
  const [levels, setLevels] = useState<Level[]>([]);
  const [subjects, setSubjects] = useState<Subject[]>([]);
  const [editingTopic, setEditingTopic] = useState<Topic | null>(null);
  const [formData, setFormData] = useState({ idMataPelajaran: '', idTingkat: '', nama: '' });
  const { isOpen, onOpen, onClose } = useDisclosure();
  const toast = useToast();

  useEffect(() => {
    fetchTopics();
    fetchLevels();
    fetchSubjects();
  }, []);

  const fetchTopics = async () => {
    try {
      const response = await axios.get(API_BASE);
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

  const fetchSubjects = async () => {
    try {
      const response = await axios.get(SUBJECTS_API);
      const data = response.data;
      setSubjects(
        Array.isArray(data) ? data :
        Array.isArray(data.data) ? data.data :
        Array.isArray(data.mataPelajaran) ? data.mataPelajaran : []
      );
    } catch (error) {
      toast({ title: 'Error fetching subjects', status: 'error' });
      setSubjects([]);
    }
  };

  const handleCreate = () => {
    setEditingTopic(null);
    setFormData({ idMataPelajaran: '', idTingkat: '', nama: '' });
    onOpen();
  };

  const handleEdit = (topic: Topic) => {
    setEditingTopic(topic);
    setFormData({
      idMataPelajaran: topic.mataPelajaran.id.toString(),
      idTingkat: topic.tingkat.id.toString(),
      nama: topic.nama,
    });
    onOpen();
  };

  const handleDelete = async (id: number) => {
    try {
      await axios.delete(`${API_BASE}/${id}`);
      fetchTopics();
      toast({ title: 'Topic deleted', status: 'success' });
    } catch (error) {
      toast({ title: 'Error deleting topic', status: 'error' });
    }
  };

  const handleSubmit = async () => {
    const data = {
      idMataPelajaran: parseInt(formData.idMataPelajaran),
      idTingkat: parseInt(formData.idTingkat),
      nama: formData.nama,
    };
    try {
      if (editingTopic) {
        await axios.put(`${API_BASE}/${editingTopic.id}`, data);
        toast({ title: 'Topic updated', status: 'success' });
      } else {
        await axios.post(API_BASE, data);
        toast({ title: 'Topic created', status: 'success' });
      }
      fetchTopics();
      onClose();
    } catch (error) {
      toast({ title: 'Error saving topic', status: 'error' });
    }
  };

  const getSubjectName = (id: number) => subjects.find(s => s.id === id)?.nama || 'Unknown';
  const getLevelName = (id: number) => levels.find(l => l.id === id)?.nama || 'Unknown';

  return (
    <Container maxW="container.lg" py={10}>
      <Link href="/">
        <Button mb={4} variant="outline">
          Back to Home
        </Button>
      </Link>
      <Heading as="h1" size="xl" mb={8}>
        Manage Topics
      </Heading>
      <Button colorScheme="purple" onClick={handleCreate} mb={4}>
        Add Topic
      </Button>
      <Table variant="simple">
        <Thead>
          <Tr>
            <Th>ID</Th>
            <Th>Subject</Th>
            <Th>Level</Th>
            <Th>Name</Th>
            <Th>Actions</Th>
          </Tr>
        </Thead>
        <Tbody>
          {topics.map((topic) => (
            <Tr key={topic.id}>
              <Td>{topic.id}</Td>
              <Td>{topic.mataPelajaran.nama}</Td>
              <Td>{topic.tingkat.nama}</Td>
              <Td>{topic.nama}</Td>
              <Td>
                <Button size="sm" mr={2} onClick={() => handleEdit(topic)}>
                  Edit
                </Button>
                <Button size="sm" colorScheme="red" onClick={() => handleDelete(topic.id)}>
                  Delete
                </Button>
              </Td>
            </Tr>
          ))}
        </Tbody>
      </Table>

      <Modal isOpen={isOpen} onClose={onClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>{editingTopic ? 'Edit Topic' : 'Add Topic'}</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4}>
              <FormControl>
                <FormLabel>Subject</FormLabel>
                <Select
                  value={formData.idMataPelajaran}
                  onChange={(e) => setFormData({ ...formData, idMataPelajaran: e.target.value })}
                >
                  <option value="">Select Subject</option>
                  {subjects.map((subject) => (
                    <option key={subject.id} value={subject.id.toString()}>
                      {subject.nama}
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
                <FormLabel>Name</FormLabel>
                <Input
                  value={formData.nama}
                  onChange={(e) => setFormData({ ...formData, nama: e.target.value })}
                />
              </FormControl>
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button colorScheme="purple" mr={3} onClick={handleSubmit}>
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