'use client';

import { useState, useEffect, useMemo, useCallback } from 'react';
import React from 'react';
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
  Text,
} from '@chakra-ui/react';
import { useCRUD, useForm, usePagination } from '../hooks';
import { Level } from '../types';

export default React.memo(function LevelsTab() {
  const { data: levels, create, update, remove } = useCRUD<Level>('levels');
  const [editingLevel, setEditingLevel] = useState<Level | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedSearchQuery, setDebouncedSearchQuery] = useState('');
  const { isOpen, onOpen, onClose } = useDisclosure();

  const form = useForm({
    initialValues: { nama: '' },
    onSubmit: async (values) => {
      if (editingLevel) {
        await update(editingLevel.id, values);
      } else {
        await create(values);
      }
      onClose();
      form.reset();
      setEditingLevel(null);
    },
  });

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearchQuery(searchQuery), 300);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  const filteredLevels = useMemo(() => {
    return levels.filter((level) =>
      level.nama.toLowerCase().includes(debouncedSearchQuery.toLowerCase())
    );
  }, [levels, debouncedSearchQuery]);

  const { paginatedItems, currentPage, totalPages, goToPage, nextPage, prevPage } =
    usePagination(filteredLevels, { itemsPerPage: 10 });

  const handleCreate = useCallback(() => {
    setEditingLevel(null);
    form.reset();
    onOpen();
  }, [form, onOpen]);

  const handleEdit = useCallback(
    (level: Level) => {
      setEditingLevel(level);
      form.setFieldValue('nama', level.nama);
      onOpen();
    },
    [form, onOpen]
  );

  const handleDelete = useCallback(
    async (id: number) => {
      await remove(id);
    },
    [remove]
  );

  return (
    <Box>
      <Button colorScheme="blue" onClick={handleCreate} mb={4}>
        Tambah Tingkat
      </Button>
      <Input
        placeholder="Cari tingkat..."
        value={searchQuery}
        onChange={(e) => setSearchQuery(e.target.value)}
        mb={4}
      />
      <Table variant="simple">
        <Thead>
          <Tr>
            <Th>Tingkat</Th>
            <Th>Aksi</Th>
          </Tr>
        </Thead>
        <Tbody>
          {paginatedItems.map((level) => (
            <Tr key={level.id}>
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
      <Box mt={4} display="flex" justifyContent="space-between" alignItems="center">
        <Button isDisabled={currentPage === 1} onClick={prevPage}>
          Prev
        </Button>
        <Text>
          Halaman {currentPage} dari {totalPages}
        </Text>
        <Button isDisabled={currentPage === totalPages} onClick={nextPage}>
          Next
        </Button>
      </Box>

      <Modal isOpen={isOpen} onClose={onClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>{editingLevel ? 'Edit Tingkat' : 'Tambah Tingkat'}</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <FormControl>
              <FormLabel>Nama Tingkatan</FormLabel>
              <Input
                name="nama"
                value={form.values.nama}
                onChange={form.handleChange}
                placeholder="Masukkan nama tingkatan"
              />
            </FormControl>
          </ModalBody>
          <ModalFooter>
            <Button
              colorScheme="blue"
              mr={3}
              onClick={() => form.handleSubmit()}
              isLoading={form.isSubmitting}
            >
              Simpan
            </Button>
            <Button
              variant="ghost"
              onClick={() => {
                onClose();
                setEditingLevel(null);
                form.reset();
              }}
            >
              Batal
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </Box>
  );
})