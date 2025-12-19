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

const API_BASE = process.env.NEXT_PUBLIC_API_BASE + '/v1/levels';

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
      toast({ title: 'Error mengambil tingkat', status: 'error' });
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
      toast({ title: 'Tingkat dihapus', status: 'success' });
    } catch (error) {
      toast({ title: 'Error menghapus tingkat', status: 'error' });
    }
  };

  const handleSubmit = async () => {
    try {
      if (editingLevel) {
        await axios.put(`${API_BASE}/${editingLevel.id}`, formData);
        toast({ title: 'Tingkat diperbarui', status: 'success' });
      } else {
        await axios.post(API_BASE, formData);
        toast({ title: 'Tingkat dibuat', status: 'success' });
      }
      fetchLevels();
      onClose();
    } catch (error) {
      toast({ title: 'Error menyimpan tingkat', status: 'error' });
    }
  };

  return (
    <Box>
      <Button colorScheme="blue" onClick={handleCreate} mb={4}>
        Tambah Tingkat
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
          {levels.map((level) => (
            <Tr key={level.id}>
              <Td>{level.id}</Td>
              <Td>{level.nama}</Td>
              <Td>
                <Button size="sm" mr={2} onClick={() => handleEdit(level)}>
                  Edit
                </Button>
                <Button size="sm" colorScheme="red" onClick={() => handleDelete(level.id)}>
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
          <ModalHeader>{editingLevel ? 'Edit Tingkat' : 'Tambah Tingkat'}</ModalHeader>
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
            <Button colorScheme="blue" mr={3} onClick={handleSubmit}>
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