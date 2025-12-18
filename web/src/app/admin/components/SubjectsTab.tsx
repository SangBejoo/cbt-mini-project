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
  useDisclosure,
  useToast,
} from '@chakra-ui/react';
import axios from 'axios';

interface Subject {
  id: number;
  nama: string;
}

const API_BASE = 'http://localhost:8080/v1/subjects';

export default function SubjectsTab() {
  const [subjects, setSubjects] = useState<Subject[]>([]);
  const [editingSubject, setEditingSubject] = useState<Subject | null>(null);
  const [formData, setFormData] = useState({ nama: '' });
  const { isOpen, onOpen, onClose } = useDisclosure();
  const toast = useToast();

  useEffect(() => {
    fetchSubjects();
  }, []);

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

  return (
    <Box>
      <Button colorScheme="green" onClick={handleCreate} mb={4}>
        Tambah Mata Pelajaran
      </Button>
      <Table variant="simple">
        <Thead>
          <Tr>
            <Th>ID</Th>
            <Th>Nama</Th>
            <Th>Aksi</Th>
          </Tr>
        </Thead>
        <Tbody>
          {subjects.map((subject) => (
            <Tr key={subject.id}>
              <Td>{subject.id}</Td>
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