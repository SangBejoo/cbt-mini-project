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
  useDisclosure,
  useToast,
  Text,
} from '@chakra-ui/react';
import axios from 'axios';

interface Subject {
  id: number;
  nama: string;
}

const API_BASE = process.env.NEXT_PUBLIC_API_BASE + '/v1/subjects';

export default function SubjectsTab() {
  const [subjects, setSubjects] = useState<Subject[]>([]);
  const [editingSubject, setEditingSubject] = useState<Subject | null>(null);
  const [formData, setFormData] = useState({ nama: '' });
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedSearchQuery, setDebouncedSearchQuery] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 10;
  const { isOpen, onOpen, onClose } = useDisclosure();
  const toast = useToast();

  useEffect(() => {
    fetchSubjects();
  }, []);

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearchQuery(searchQuery), 500);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  const fetchSubjects = async () => {
    try {
      const response = await axios.get(API_BASE);
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

  const filteredSubjects = useMemo(() => {
    return subjects.filter(subject => subject.nama.toLowerCase().includes(debouncedSearchQuery.toLowerCase()));
  }, [subjects, debouncedSearchQuery]);

  const totalPages = Math.ceil(filteredSubjects.length / itemsPerPage);
  const paginatedSubjects = filteredSubjects.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);

  const handleCreate = () => {
    setEditingSubject(null);
    setFormData({ nama: '' });
    onOpen();
  };

  const handleEdit = (subject: Subject) => {
    setEditingSubject(subject);
    setFormData({ nama: subject.nama });
    onOpen();
  };

  const handleDelete = async (id: number) => {
    try {
      await axios.delete(`${API_BASE}/${id}`);
      fetchSubjects();
      toast({ title: 'Mata pelajaran dihapus', status: 'success' });
    } catch (error) {
      toast({ title: 'Error menghapus mata pelajaran', status: 'error' });
    }
  };

  const handleSubmit = async () => {
    try {
      if (editingSubject) {
        await axios.put(`${API_BASE}/${editingSubject.id}`, formData);
        toast({ title: 'Mata pelajaran diperbarui', status: 'success' });
      } else {
        await axios.post(API_BASE, formData);
        toast({ title: 'Mata pelajaran dibuat', status: 'success' });
      }
      fetchSubjects();
      onClose();
    } catch (error) {
      toast({ title: 'Error menyimpan mata pelajaran', status: 'error' });
    }
  };

  const handleSearchKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      setDebouncedSearchQuery(searchQuery);
    }
  };

  return (
    <Box>
      <Button colorScheme="green" onClick={handleCreate} mb={4}>
        Tambah Mata Pelajaran
      </Button>
      <Input
        placeholder="Cari mata pelajaran..."
        value={searchQuery}
        onChange={(e) => setSearchQuery(e.target.value)}
        onKeyDown={handleSearchKeyDown}
        mb={4}
      />
      <Table variant="simple">
        <Thead>
          <Tr>
            <Th>Nama Mata Pelajaran</Th>
            <Th>Aksi</Th>
          </Tr>
        </Thead>
        <Tbody>
          {paginatedSubjects.map((subject) => (
            <Tr key={subject.id}>
              <Td>{subject.nama}</Td>
              <Td>
                <Button size="sm" mr={2} onClick={() => handleEdit(subject)}>
                  Edit
                </Button>
                <Button size="sm" colorScheme="red" onClick={() => handleDelete(subject.id)}>
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
          <ModalHeader>{editingSubject ? 'Edit Mata Pelajaran' : 'Tambah Mata Pelajaran'}</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <FormControl>
              <FormLabel>Nama</FormLabel>
              <Input
                value={formData.nama}
                onChange={(e) => setFormData({ nama: e.target.value })}
              />
            </FormControl>
          </ModalBody>
          <ModalFooter>
            <Button colorScheme="green" mr={3} onClick={handleSubmit}>
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