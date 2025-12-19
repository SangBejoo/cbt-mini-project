'use client';

import { useState, useEffect, useMemo } from 'react';
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
  Text,
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

const API_BASE = process.env.NEXT_PUBLIC_API_BASE + '/v1/topics';
const LEVELS_API = process.env.NEXT_PUBLIC_API_BASE + '/v1/levels';
const SUBJECTS_API = process.env.NEXT_PUBLIC_API_BASE + '/v1/subjects';

export default function TopicsTab() {
  const [topics, setTopics] = useState<Topic[]>([]);
  const [levels, setLevels] = useState<Level[]>([]);
  const [subjects, setSubjects] = useState<Subject[]>([]);
  const [editingTopic, setEditingTopic] = useState<Topic | null>(null);
  const [formData, setFormData] = useState({ idMataPelajaran: '', idTingkat: '', nama: '' });
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedSearchQuery, setDebouncedSearchQuery] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 10;
  const { isOpen, onOpen, onClose } = useDisclosure();
  const toast = useToast();

  useEffect(() => {
    fetchTopics();
    fetchLevels();
    fetchSubjects();
  }, []);

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearchQuery(searchQuery), 500);
    return () => clearTimeout(timer);
  }, [searchQuery]);

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

  const filteredTopics = useMemo(() => {
    return topics.filter(topic =>
      topic.nama.toLowerCase().includes(debouncedSearchQuery.toLowerCase()) ||
      topic.mataPelajaran.nama.toLowerCase().includes(debouncedSearchQuery.toLowerCase()) ||
      topic.tingkat.nama.toLowerCase().includes(debouncedSearchQuery.toLowerCase())
    );
  }, [topics, debouncedSearchQuery]);

  const totalPages = Math.ceil(filteredTopics.length / itemsPerPage);
  const paginatedTopics = filteredTopics.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);

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

  const handleSearchKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      setDebouncedSearchQuery(searchQuery);
    }
  };

  return (
    <Box>
      <Button colorScheme="purple" onClick={handleCreate} mb={4}>
        Tambah Materi
      </Button>
      <Input
        placeholder="Cari materi, mata pelajaran, atau tingkat..."
        value={searchQuery}
        onChange={(e) => setSearchQuery(e.target.value)}
        onKeyDown={handleSearchKeyDown}
        mb={4}
      />
      <Table variant="simple">
        <Thead>
          <Tr>
            <Th>Mata Pelajaran</Th>
            <Th>Tingkat</Th>
            <Th>Nama Materi</Th>
            <Th>Aksi</Th>
          </Tr>
        </Thead>
        <Tbody>
          {paginatedTopics.map((topic) => (
            <Tr key={topic.id}>
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
      <Box mt={4} display="flex" justifyContent="space-between" alignItems="center">
        <Button isDisabled={currentPage === 1} onClick={() => setCurrentPage(currentPage - 1)}>
          Prev
        </Button>
        <Text>Halaman {currentPage} dari {totalPages}</Text>
        <Button isDisabled={currentPage === totalPages} onClick={() => setCurrentPage(currentPage + 1)}>
          Next
        </Button>
      </Box>

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