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

interface Level {
  id: number;
  nama: string;
}

const API_BASE = 'http://localhost:8080/v1/levels';

export default function LevelsTab() {
  const [levels, setLevels] = useState<Level[]>([]);
  const [editingLevel, setEditingLevel] = useState<Level | null>(null);
  const [formData, setFormData] = useState({ nama: '' });
  const { isOpen, onOpen, onClose } = useDisclosure();
  const toast = useToast();

  useEffect(() => {
    fetchLevels();
  }, []);

  const fetchLevels = async () => {
    try {
      const response = await axios.get(API_BASE);
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
    setEditingLevel(null);
    setFormData({ nama: '' });
    onOpen();
  };

  const handleEdit = (level: Level) => {
    setEditingLevel(level);
    setFormData({ nama: level.nama });
    onOpen();
  };

  const handleDelete = async (id: number) => {
    try {
      await axios.delete(`${API_BASE}/${id}`);
      fetchLevels();
      toast({ title: 'Level deleted', status: 'success' });
    } catch (error) {
      toast({ title: 'Error deleting level', status: 'error' });
    }
  };

  const handleSubmit = async () => {
    try {
      if (editingLevel) {
        await axios.put(`${API_BASE}/${editingLevel.id}`, formData);
        toast({ title: 'Level updated', status: 'success' });
      } else {
        await axios.post(API_BASE, formData);
        toast({ title: 'Level created', status: 'success' });
      }
      fetchLevels();
      onClose();
    } catch (error) {
      toast({ title: 'Error saving level', status: 'error' });
    }
  };

  return (
    <Box>
      <Button colorScheme="blue" onClick={handleCreate} mb={4}>
        Add Level
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
          {levels.map((level) => (
            <Tr key={level.id}>
              <Td>{level.id}</Td>
              <Td>{level.nama}</Td>
              <Td>
                <Button size="sm" mr={2} onClick={() => handleEdit(level)}>
                  Edit
                </Button>
                <Button size="sm" colorScheme="red" onClick={() => handleDelete(level.id)}>
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
          <ModalHeader>{editingLevel ? 'Edit Level' : 'Add Level'}</ModalHeader>
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
            <Button colorScheme="blue" mr={3} onClick={handleSubmit}>
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