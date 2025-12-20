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
  Select,
  useDisclosure,
  VStack,
  Text,
} from '@chakra-ui/react';
import { useCRUD, useForm, usePagination } from '../hooks';
import { Topic, Level, Subject } from '../types';

export default React.memo(function TopicsTab() {
  const { data: topics, create, update, remove } = useCRUD<Topic>('topics');
  const { data: levels } = useCRUD<Level>('levels');
  const { data: subjects } = useCRUD<Subject>('subjects');
  const [editingTopic, setEditingTopic] = useState<Topic | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedSearchQuery, setDebouncedSearchQuery] = useState('');
  const { isOpen, onOpen, onClose } = useDisclosure();

  const form = useForm({
    initialValues: { idMataPelajaran: '', idTingkat: '', nama: '' },
    onSubmit: async (values) => {
      const data = {
        id_mata_pelajaran: parseInt(values.idMataPelajaran),
        id_tingkat: parseInt(values.idTingkat),
        nama: values.nama,
      } as any;
      if (editingTopic) {
        await update(editingTopic.id, data);
      } else {
        await create(data);
      }
      onClose();
      form.reset();
      setEditingTopic(null);
    },
  });

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearchQuery(searchQuery), 300);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  const filteredTopics = useMemo(() => {
    return topics.filter((topic) =>
      topic.nama.toLowerCase().includes(debouncedSearchQuery.toLowerCase()) ||
      topic.mataPelajaran.nama
        .toLowerCase()
        .includes(debouncedSearchQuery.toLowerCase()) ||
      topic.tingkat.nama.toLowerCase().includes(debouncedSearchQuery.toLowerCase())
    );
  }, [topics, debouncedSearchQuery]);

  const { paginatedItems, currentPage, totalPages, nextPage, prevPage } =
    usePagination(filteredTopics, { itemsPerPage: 10 });

  const handleCreate = useCallback(() => {
    setEditingTopic(null);
    form.reset();
    onOpen();
  }, [form, onOpen]);

  const handleEdit = useCallback(
    (topic: Topic) => {
      setEditingTopic(topic);
      form.setFieldValue('idMataPelajaran', topic.mataPelajaran.id.toString());
      form.setFieldValue('idTingkat', topic.tingkat.id.toString());
      form.setFieldValue('nama', topic.nama);
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
      <Button colorScheme="purple" onClick={handleCreate} mb={4}>
        Tambah Materi
      </Button>
      <Input
        placeholder="Cari materi, mata pelajaran, atau tingkat..."
        value={searchQuery}
        onChange={(e) => setSearchQuery(e.target.value)}
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
          {paginatedItems.map((topic) => (
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
          <ModalHeader>{editingTopic ? 'Edit Materi' : 'Tambah Materi'}</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4}>
              <FormControl>
                <FormLabel>Mata Pelajaran</FormLabel>
                <Select
                  name="idMataPelajaran"
                  value={form.values.idMataPelajaran}
                  onChange={form.handleChange}
                  placeholder="Pilih mata pelajaran"
                >
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
                  name="idTingkat"
                  value={form.values.idTingkat}
                  onChange={form.handleChange}
                  placeholder="Pilih tingkat"
                >
                  {levels.map((level) => (
                    <option key={level.id} value={level.id.toString()}>
                      {level.nama}
                    </option>
                  ))}
                </Select>
              </FormControl>
              <FormControl>
                <FormLabel>Nama Materi</FormLabel>
                <Input
                  name="nama"
                  value={form.values.nama}
                  onChange={form.handleChange}
                  placeholder="Masukkan nama materi"
                />
              </FormControl>
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button
              colorScheme="purple"
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
                setEditingTopic(null);
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