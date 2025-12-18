'use client';

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

export default function TopicsTab() {
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
      toast({ title: 'Error mengambil materi', status: 'error' });
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
      toast({ title: 'Error mengambil tingkat', status: 'error' });
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
      toast({ title: 'Error mengambil mata pelajaran', status: 'error' });
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
      toast({ title: 'Materi dihapus', status: 'success' });
    } catch (error) {
      toast({ title: 'Error menghapus materi', status: 'error' });
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
        toast({ title: 'Materi diperbarui', status: 'success' });
      } else {
        await axios.post(API_BASE, data);
        toast({ title: 'Materi dibuat', status: 'success' });
      }
      fetchTopics();
      onClose();
    } catch (error) {
      toast({ title: 'Error menyimpan materi', status: 'error' });
    }
  };

  return (
    <Box>
      <Button colorScheme="purple" onClick={handleCreate} mb={4}>
        Tambah Materi
      </Button>
      <Table variant="simple">
        <Thead>
          <Tr>
            <Th>ID</Th>
            <Th>Mata Pelajaran</Th>
            <Th>Tingkat</Th>
            <Th>Nama</Th>
            <Th>Aksi</Th>
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
                  Hapus
                </Button>
              </Td>
            </Tr>
          ))}
        </Tbody>
      </Table>

      <Modal isOpen={isOpen} onClose={onClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>{editingTopic ? 'Edit Materi' : 'Tambah Materi'}</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4}>
              <FormControl>
                <FormLabel>Mata Pelajaran</FormLabel>
                <Select
                  value={formData.idMataPelajaran}
                  onChange={(e) => setFormData({ ...formData, idMataPelajaran: e.target.value })}
                >
                  <option value="">Pilih Mata Pelajaran</option>
                  {subjects.map((subject) => (
                    <option key={subject.id} value={subject.id.toString()}>
                      {subject.nama}
                    </option>
                  ))}
                </Select>
              </FormControl>
              <FormControl>
                <FormLabel>Tingkat</FormLabel>
                <Select
                  value={formData.idTingkat}
                  onChange={(e) => setFormData({ ...formData, idTingkat: e.target.value })}
                >
                  <option value="">Pilih Tingkat</option>
                  {levels.map((level) => (
                    <option key={level.id} value={level.id.toString()}>
                      {level.nama}
                    </option>
                  ))}
                </Select>
              </FormControl>
              <FormControl>
                <FormLabel>Nama</FormLabel>
                <Input
                  value={formData.nama}
                  onChange={(e) => setFormData({ ...formData, nama: e.target.value })}
                />
              </FormControl>
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button colorScheme="purple" mr={3} onClick={handleSubmit}>
              Simpan
            </Button>
            <Button variant="ghost" onClick={onClose}>
              Batal
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </Box>
  );
}