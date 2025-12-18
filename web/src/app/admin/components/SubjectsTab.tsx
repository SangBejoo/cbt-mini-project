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
      toast({ title: 'Error fetching subjects', status: 'error' });
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
      toast({ title: 'Subject deleted', status: 'success' });
    } catch (error) {
      toast({ title: 'Error deleting subject', status: 'error' });
    }
  };

  const handleSubmit = async () => {
    try {
      if (editingSubject) {
        await axios.put(`${API_BASE}/${editingSubject.id}`, formData);
        toast({ title: 'Subject updated', status: 'success' });
      } else {
        await axios.post(API_BASE, formData);
        toast({ title: 'Subject created', status: 'success' });
      }
      fetchSubjects();
      onClose();
    } catch (error) {
      toast({ title: 'Error saving subject', status: 'error' });
    }
  };

  return (
    <Box>
      <Button colorScheme="green" onClick={handleCreate} mb={4}>
        Add Subject
      </Button>
      <Table variant="simple">
        <Thead>
          <Tr>
            <Th>ID</Th>
            <Th>Name</Th>
            <Th>Actions</Th>
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
          <ModalHeader>{editingSubject ? 'Edit Subject' : 'Add Subject'}</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <FormControl>
              <FormLabel>Name</FormLabel>
              <Input
                value={formData.nama}
                onChange={(e) => setFormData({ nama: e.target.value })}
              />
            </FormControl>
          </ModalBody>
          <ModalFooter>
            <Button colorScheme="green" mr={3} onClick={handleSubmit}>
              Save
            </Button>
            <Button variant="ghost" onClick={onClose}>
              Cancel
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </Box>
  );
}