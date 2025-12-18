'use client';

import Link from 'next/link';
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import {
  Box,
  Button,
  FormControl,
  FormLabel,
  Input,
  VStack,
  Heading,
  Container,
  useToast,
  Card,
  CardBody,
  Text,
  SimpleGrid,
  Icon,
  HStack,
  useDisclosure,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  ModalCloseButton,
} from '@chakra-ui/react';
import { BookOpen } from 'lucide-react';
import axios from 'axios';

interface Topic {
  id: number;
  mataPelajaran: { id: number; nama: string };
  tingkat: { id: number; nama: string };
  nama: string;
}

const TOPICS_API = 'http://localhost:8080/v1/topics';
const CREATE_SESSION_API = 'http://localhost:8080/v1/sessions';

export default function SessionsPage() {
  const [topics, setTopics] = useState<Topic[]>([]);
  const [selectedTopic, setSelectedTopic] = useState<Topic | null>(null);
  const [namaPeserta, setNamaPeserta] = useState('');
  const [loading, setLoading] = useState(false);
  const [mounted, setMounted] = useState(false);
  const router = useRouter();
  const toast = useToast();
  const { isOpen, onOpen, onClose } = useDisclosure();

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (mounted) {
      fetchTopics();
    }
  }, [mounted]);

  const fetchTopics = async () => {
    try {
      const response = await axios.get(TOPICS_API);
      const data = response.data;
      setTopics(Array.isArray(data) ? data : Array.isArray(data.materi) ? data.materi : []);
    } catch (error) {
      console.error('Error fetching topics:', error);
    }
  };

  const handleTopicClick = (topic: Topic) => {
    setSelectedTopic(topic);
    onOpen();
  };

  const handleStartTest = async () => {
    if (!namaPeserta || !selectedTopic) {
      toast({ title: 'Masukkan nama Anda', status: 'error' });
      return;
    }

    setLoading(true);
    try {
      const payload = {
        nama_peserta: namaPeserta,
        id_tingkat: selectedTopic.tingkat.id,
        id_mata_pelajaran: selectedTopic.mataPelajaran.id,
        durasi_menit: 60,
        jumlah_soal: 20,
      };

      const response = await axios.post(CREATE_SESSION_API, payload);
      const sessionToken =
        response.data?.testSession?.sessionToken ||
        response.data?.test_session?.session_token ||
        response.data?.session_token ||
        response.data?.token;

      if (!sessionToken) {
        toast({ title: 'Token sesi tidak ditemukan', status: 'error' });
        return;
      }

      toast({ title: 'Sesi tes berhasil dibuat!', status: 'success' });
      router.push(`/student/test/${sessionToken}`);
    } catch (error: any) {
      console.error('Error creating session:', error);
      const errorMessage = error.response?.data?.message || error.message || 'Error membuat sesi tes';
      toast({ title: errorMessage, status: 'error', duration: 5000 });
    } finally {
      setLoading(false);
      onClose();
    }
  };

  if (!mounted) return null;

  return (
    <Container maxW="container.xl" py={10}>
      <VStack spacing={6}>
        <Heading as="h1" size="xl" textAlign="center" mb={4}>
          Pilih Materi Tes
        </Heading>

        <SimpleGrid columns={{ base: 1, md: 2, lg: 3 }} spacing={6} width="full">
          {topics.map((topic) => (
            <Card
              key={topic.id}
              cursor="pointer"
              onClick={() => handleTopicClick(topic)}
              _hover={{ transform: 'translateY(-4px)', shadow: 'xl' }}
              transition="all 0.3s"
              bg="blue.50"
              borderWidth="2px"
              borderColor="blue.200"
            >
              <CardBody>
                <VStack spacing={4} align="center">
                  <Box
                    bg="orange.400"
                    p={4}
                    borderRadius="full"
                  >
                    <Text fontSize="3xl">📚</Text>
                  </Box>
                  <VStack spacing={1}>
                    <Text fontWeight="bold" fontSize="lg" textAlign="center">
                      {topic.mataPelajaran.nama.toUpperCase()} {topic.tingkat.nama} SD KELAS {topic.tingkat.nama === '1' ? 'I' : topic.tingkat.nama === '2' ? 'II' : topic.tingkat.nama === '3' ? 'III' : 'IV'}
                    </Text>
                    <Text fontSize="sm" color="gray.600" textAlign="center">
                      {topic.nama}
                    </Text>
                  </VStack>
                  <HStack spacing={3} width="full" justify="center">
                    <Button
                      size="sm"
                      variant="outline"
                      colorScheme="orange"
                      onClick={(e) => {
                        e.stopPropagation();
                        router.push('/student/history');
                      }}
                    >
                      Riwayat Nilai Tes 📋
                    </Button>
                    <Button
                      size="sm"
                      colorScheme="orange"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleTopicClick(topic);
                      }}
                    >
                      Mulai CBT
                    </Button>
                  </HStack>
                </VStack>
              </CardBody>
            </Card>
          ))}
        </SimpleGrid>

        <Link href="/student">
          <Button variant="outline" mt={4}>
            Kembali
          </Button>
        </Link>
      </VStack>

      {/* Name Input Modal */}
      <Modal isOpen={isOpen} onClose={onClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Mulai Tes</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4}>
              <Text fontSize="lg" fontWeight="bold">
                {selectedTopic?.mataPelajaran.nama} - {selectedTopic?.nama}
              </Text>
              <FormControl isRequired>
                <FormLabel>Nama Anda</FormLabel>
                <Input
                  value={namaPeserta}
                  onChange={(e) => setNamaPeserta(e.target.value)}
                  placeholder="Masukkan nama Anda"
                />
              </FormControl>
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button colorScheme="orange" mr={3} onClick={handleStartTest} isLoading={loading}>
              Mulai Tes
            </Button>
            <Button variant="ghost" onClick={onClose}>
              Batal
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </Container>
  );
}