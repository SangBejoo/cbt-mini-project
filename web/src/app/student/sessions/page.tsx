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
  Select,
  VStack,
  Heading,
  Container,
  useToast,
  Card,
  CardBody,
  Text,
} from '@chakra-ui/react';
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
  const [formData, setFormData] = useState({
    namaPeserta: '',
    idMateri: '',
    durasiMenit: '60', // Fixed to 60 minutes
  });
  const [loading, setLoading] = useState(false);
  const router = useRouter();
  const toast = useToast();

  useEffect(() => {
    fetchTopics();
  }, []);

  const fetchTopics = async () => {
    try {
      const response = await axios.get(TOPICS_API);
      const data = response.data;
      setTopics(Array.isArray(data) ? data : Array.isArray(data.materi) ? data.materi : []);
    } catch (error) {
      console.error('Error fetching topics:', error);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.namaPeserta || !formData.idMateri) {
      toast({ title: 'Please fill all fields', status: 'error' });
      return;
    }

    setLoading(true);
    try {
      const selectedTopic = topics.find(t => t.id.toString() === formData.idMateri);
      if (!selectedTopic) {
        toast({ title: 'Invalid topic selected', status: 'error' });
        return;
      }

      const payload = {
        nama_peserta: formData.namaPeserta,
        id_tingkat: selectedTopic.tingkat.id,
        id_mata_pelajaran: selectedTopic.mataPelajaran.id,
        durasi_menit: parseInt(formData.durasiMenit),
        jumlah_soal: 15, // Reduced for testing - only 1 question
      };

      const response = await axios.post(CREATE_SESSION_API, payload);
      console.log('Session creation response:', response.data);
      // grpc-gateway maps proto field names to lowerCamelCase JSON keys by default
      const sessionToken =
        response.data?.testSession?.sessionToken ||
        response.data?.test_session?.session_token ||
        response.data?.session_token ||
        response.data?.token;
      console.log('Extracted session token:', sessionToken);

      if (!sessionToken) {
        toast({ title: 'Session token missing from response', status: 'error' });
        return;
      }

      toast({ title: 'Test session created!', status: 'success' });
      router.push(`/student/test/${sessionToken}`);
    } catch (error) {
      console.error('Error creating session:', error);
      toast({ title: 'Error creating session', status: 'error' });
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container maxW="container.md" py={10}>
      <Heading as="h1" size="xl" mb={8}>
        Start a New Test
      </Heading>

      <Card>
        <CardBody>
          <form onSubmit={handleSubmit}>
            <VStack spacing={4}>
              <FormControl isRequired>
                <FormLabel>Participant Name</FormLabel>
                <Input
                  value={formData.namaPeserta}
                  onChange={(e) => setFormData({ ...formData, namaPeserta: e.target.value })}
                  placeholder="Enter your name"
                />
              </FormControl>

              <FormControl isRequired>
                <FormLabel>Topic</FormLabel>
                <Select
                  value={formData.idMateri}
                  onChange={(e) => setFormData({ ...formData, idMateri: e.target.value })}
                  placeholder="Select topic"
                >
                  {topics.map((topic) => (
                    <option key={topic.id} value={topic.id.toString()}>
                      {topic.tingkat.nama} - {topic.mataPelajaran.nama} - {topic.nama}
                    </option>
                  ))}
                </Select>
              </FormControl>

              <Text fontSize="sm" color="gray.600">
                Duration: 60 minutes (fixed for elementary school students)
              </Text>

              <Button
                type="submit"
                colorScheme="blue"
                size="lg"
                width="full"
                isLoading={loading}
                loadingText="Creating session..."
              >
                Start Test
              </Button>
            </VStack>
          </form>
        </CardBody>
      </Card>

      <Box mt={8}>
        <Link href="/student">
          <Button variant="outline" size="sm">
            ← Back to Student Portal
          </Button>
        </Link>
      </Box>
    </Container>
  );
}