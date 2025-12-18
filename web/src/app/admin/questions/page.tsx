'use client';

import Link from 'next/link';
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
  Select,
  Textarea,
  useDisclosure,
  VStack,
  Heading,
  Container,
  useToast,
  HStack,
  Image,
  IconButton,
  Text,
  Badge,
} from '@chakra-ui/react';
import { CloseIcon } from '@chakra-ui/icons';
import axios from 'axios';

interface Question {
  id: number;
  materi: {
    id: number;
    mataPelajaran: { id: number; nama: string };
    tingkat: { id: number; nama: string };
    nama: string;
  };
  pertanyaan: string;
  opsiA: string;
  opsiB: string;
  opsiC: string;
  opsiD: string;
  jawabanBenar: string;
  gambar?: Array<{
    id: number;
    nama_file: string;
    file_path: string;
    file_size: number;
    mime_type: string;
    urutan: number;
    keterangan?: string;
    created_at: string;
  }>;
}

interface Topic {
  id: number;
  nama: string;
}

interface Level {
  id: number;
  nama: string;
}

const API_BASE = 'http://localhost:8080/v1/questions';
const TOPICS_API = 'http://localhost:8080/v1/topics';
const LEVELS_API = 'http://localhost:8080/v1/levels';

export default function QuestionsPage() {
  const [questions, setQuestions] = useState<Question[]>([]);
  const [topics, setTopics] = useState<Topic[]>([]);
  const [levels, setLevels] = useState<Level[]>([]);
  const [editingQuestion, setEditingQuestion] = useState<Question | null>(null);
  const [formData, setFormData] = useState({
    idMateri: '',
    idTingkat: '',
    pertanyaan: '',
    opsiA: '',
    opsiB: '',
    opsiC: '',
    opsiD: '',
    jawabanBenar: '',
    images: [] as File[],
    existingImages: [] as Array<{
      id: number;
      nama_file: string;
      file_path: string;
      file_size: number;
      mime_type: string;
      urutan: number;
      keterangan?: string;
      created_at: string;
    }>,
  });
  const { isOpen, onOpen, onClose } = useDisclosure();
  const toast = useToast();

  useEffect(() => {
    fetchQuestions();
    fetchTopics();
    fetchLevels();
  }, []);

  const fetchQuestions = async () => {
    try {
      const response = await axios.get(API_BASE);
      const data = response.data;
      setQuestions(
        Array.isArray(data) ? data :
        Array.isArray(data.data) ? data.data :
        Array.isArray(data.soal) ? data.soal : []
      );
    } catch (error) {
      toast({ title: 'Error fetching questions', status: 'error' });
      setQuestions([]);
    }
  };

  const fetchTopics = async () => {
    try {
      const response = await axios.get(TOPICS_API);
      const data = response.data;
      setTopics(
        Array.isArray(data) ? data :
        Array.isArray(data.data) ? data.data :
        Array.isArray(data.materi) ? data.materi : []
      );
    } catch (error) {
      toast({ title: 'Error fetching topics', status: 'error' });
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
      toast({ title: 'Error fetching levels', status: 'error' });
      setLevels([]);
    }
  };

  const uploadImageToQuestion = async (questionId: number, file: File, urutan: number, keterangan?: string) => {
    const formDataUpload = new FormData();
    formDataUpload.append('image_bytes', file);
    formDataUpload.append('nama_file', file.name);
    formDataUpload.append('urutan', urutan.toString());
    if (keterangan) formDataUpload.append('keterangan', keterangan);

    try {
      const response = await axios.post(`http://localhost:8080/v1/questions/${questionId}/images`, formDataUpload, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      toast({ title: 'Image uploaded', status: 'success' });
      return response.data.gambar;
    } catch (error) {
      toast({ title: 'Error uploading image', status: 'error' });
      throw error;
    }
  };

  const deleteImageFromQuestion = async (imageId: number) => {
    try {
      await axios.delete(`http://localhost:8080/v1/questions/images/${imageId}`);
      toast({ title: 'Image deleted', status: 'success' });
    } catch (error) {
      toast({ title: 'Error deleting image', status: 'error' });
      throw error;
    }
  };

  const arrayBufferToBase64 = (buffer: ArrayBuffer) => {
    let binary = '';
    const bytes = new Uint8Array(buffer);
    const len = bytes.byteLength;
    for (let i = 0; i < len; i++) {
      binary += String.fromCharCode(bytes[i]);
    }
    return btoa(binary);
  };

  const handleCreate = () => {
    setEditingQuestion(null);
    setFormData({
      idMateri: '',
      idTingkat: '',
      pertanyaan: '',
      opsiA: '',
      opsiB: '',
      opsiC: '',
      opsiD: '',
      jawabanBenar: '',
      images: [],
      existingImages: [],
    });
    onOpen();
  };

  const handleEdit = (question: Question) => {
    setEditingQuestion(question);
    setFormData({
      idMateri: question.materi.id.toString(),
      idTingkat: question.materi.tingkat.id.toString(),
      pertanyaan: question.pertanyaan,
      opsiA: question.opsiA,
      opsiB: question.opsiB,
      opsiC: question.opsiC,
      opsiD: question.opsiD,
      jawabanBenar: question.jawabanBenar,
      images: [],
      existingImages: question.gambar || [],
    });
    onOpen();
  };

  const handleDelete = async (id: number) => {
    try {
      await axios.delete(`${API_BASE}/${id}`);
      fetchQuestions();
      toast({ title: 'Question deleted', status: 'success' });
    } catch (error) {
      toast({ title: 'Error deleting question', status: 'error' });
    }
  };

  const handleSubmit = async () => {
    const data: any = {
      id_materi: parseInt(formData.idMateri),
      id_tingkat: parseInt(formData.idTingkat),
      pertanyaan: formData.pertanyaan,
      opsi_a: formData.opsiA,
      opsi_b: formData.opsiB,
      opsi_c: formData.opsiC,
      opsi_d: formData.opsiD,
      jawaban_benar: formData.jawabanBenar,
    };

    if (formData.images.length > 0) {
      const imageBytesArray = await Promise.all(
        formData.images.map(async (file) => {
          const buffer = await file.arrayBuffer();
          return arrayBufferToBase64(buffer);
        })
      );
      data.image_bytes = imageBytesArray;
    }

    try {
      if (editingQuestion) {
        await axios.put(`${API_BASE}/${editingQuestion.id}`, data);
        toast({ title: 'Question updated', status: 'success' });
      } else {
        await axios.post(API_BASE, data);
        toast({ title: 'Question created', status: 'success' });
      }
      fetchQuestions();
      onClose();
    } catch (error) {
      toast({ title: 'Error saving question', status: 'error' });
    }
  };

  const getTopicName = (id: number) => topics.find(t => t.id === id)?.nama || 'Unknown';
  const getLevelName = (id: number) => levels.find(l => l.id === id)?.nama || 'Unknown';

  const handleAddImages = (files: FileList | null) => {
    if (files) {
      const newImages = Array.from(files);
      setFormData(prev => ({
        ...prev,
        images: [...prev.images, ...newImages]
      }));
    }
  };

  const handleRemoveNewImage = (index: number) => {
    setFormData(prev => ({
      ...prev,
      images: prev.images.filter((_, i) => i !== index)
    }));
  };

  const handleRemoveExistingImage = async (imageId: number) => {
    try {
      await deleteImageFromQuestion(imageId);
      setFormData(prev => ({
        ...prev,
        existingImages: prev.existingImages.filter(img => img.id !== imageId)
      }));
      // Also update the questions list
      setQuestions(prev => prev.map(q => 
        q.id === editingQuestion?.id 
          ? { ...q, gambar: q.gambar?.filter(img => img.id !== imageId) }
          : q
      ));
    } catch (error) {
      // Error already shown in toast
    }
  };

  const handleUploadImage = async (file: File | null) => {
    if (!file || !editingQuestion) return;
    try {
      const nextUrutan = (formData.existingImages.length > 0 ? Math.max(...formData.existingImages.map(img => img.urutan)) : 0) + 1;
      const uploadedImage = await uploadImageToQuestion(editingQuestion.id, file, nextUrutan);
      setFormData(prev => ({
        ...prev,
        existingImages: [...prev.existingImages, uploadedImage]
      }));
      // Update questions list
      setQuestions(prev => prev.map(q => 
        q.id === editingQuestion.id 
          ? { ...q, gambar: [...(q.gambar || []), uploadedImage] }
          : q
      ));
    } catch (error) {
      // Error already shown
    }
  };

  // Image viewer state and helpers
  const [viewerOpen, setViewerOpen] = useState(false);
  const [viewerIndex, setViewerIndex] = useState(0);
  const [viewerImages, setViewerImages] = useState<string[]>([]);

  const openImageViewer = (images: any[], index: number) => {
    const urls = images.map((img) => {
      const path = img.file_path || img.filePath || '';
      return `http://localhost:8080/${path.replace(/\\/g, '/')}`;
    });
    setViewerImages(urls);
    setViewerIndex(index);
    setViewerOpen(true);
  };

  const closeImageViewer = () => setViewerOpen(false);

  const prevImage = () => setViewerIndex((i) => (i - 1 + viewerImages.length) % viewerImages.length);
  const nextImage = () => setViewerIndex((i) => (i + 1) % viewerImages.length);

  const groupedQuestions = questions.reduce((acc, q) => {
    const key = `${q.materi.id}-${q.materi.tingkat.id}`;
    if (!acc[key]) {
      acc[key] = {
        topic: q.materi,
        level: q.materi.tingkat,
        questions: []
      };
    }
    acc[key].questions.push(q);
    return acc;
  }, {} as Record<string, { topic: any; level: any; questions: Question[] }>);

  return (
    <Container maxW="container.xl" py={10}>
      <Link href="/">
        <Button mb={4} variant="outline">
          Back to Home
        </Button>
      </Link>
      <Heading as="h1" size="xl" mb={8}>
        Manage Questions
      </Heading>
      <Button colorScheme="orange" onClick={handleCreate} mb={4}>
        Add Question
      </Button>
      <VStack spacing={8} align="stretch">
        {Object.values(groupedQuestions).map((group) => (
          <Box key={`${group.topic.id}-${group.level.id}`}>
            <Heading size="md" mb={4}>
              Topic: {group.topic.nama} (Level: {group.level.nama})
            </Heading>
            <Table variant="simple">
              <Thead>
                <Tr>
                  <Th>ID</Th>
                  <Th>Question</Th>
                  <Th>Correct Answer</Th>
                  <Th>Image</Th>
                  <Th>Actions</Th>
                </Tr>
              </Thead>
              <Tbody>
                {group.questions.map((question) => (
                  <Tr key={question.id}>
                    <Td>{question.id}</Td>
                    <Td>{question.pertanyaan.substring(0, 50)}...</Td>
                    <Td>{question.jawabanBenar}</Td>
                    <Td>
                      {question.gambar && question.gambar.length > 0 ? (
                        <HStack spacing={2}>
                          {question.gambar.map((img, index) => {
                            const path = (img as any).file_path || (img as any).filePath || '';
                            const src = `http://localhost:8080/${path.replace(/\\/g, '/')}`;
                            return (
                              <Box key={img.id} position="relative" cursor="pointer" onClick={() => openImageViewer(question.gambar, index)}>
                                <Image
                                  src={src}
                                  alt={`Question ${index + 1}`}
                                  boxSize="50px"
                                  objectFit="cover"
                                  borderRadius="md"
                                />
                                <Badge
                                  position="absolute"
                                  top="-2"
                                  right="-2"
                                  colorScheme="blue"
                                  fontSize="xs"
                                >
                                  {img.urutan}
                                </Badge>
                              </Box>
                            )
                          })}
                        </HStack>
                      ) : (
                        <Text color="gray.500">No images</Text>
                      )}
                    </Td>
                    <Td>
                      <Button size="sm" mr={2} onClick={() => handleEdit(question)}>
                        Edit
                      </Button>
                      <Button size="sm" colorScheme="red" onClick={() => handleDelete(question.id)}>
                        Delete
                      </Button>
                    </Td>
                  </Tr>
                ))}
              </Tbody>
            </Table>
          </Box>
        ))}
      </VStack>

      <Modal isOpen={isOpen} onClose={onClose} size="xl">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>{editingQuestion ? 'Edit Question' : 'Add Question'}</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4}>
              <FormControl>
                <FormLabel>Topic</FormLabel>
                <Select
                  value={formData.idMateri}
                  onChange={(e) => setFormData({ ...formData, idMateri: e.target.value })}
                >
                  <option value="">Select Topic</option>
                  {topics.map((topic) => (
                    <option key={topic.id} value={topic.id.toString()}>
                      {topic.nama}
                    </option>
                  ))}
                </Select>
              </FormControl>
              <FormControl>
                <FormLabel>Level</FormLabel>
                <Select
                  value={formData.idTingkat}
                  onChange={(e) => setFormData({ ...formData, idTingkat: e.target.value })}
                >
                  <option value="">Select Level</option>
                  {levels.map((level) => (
                    <option key={level.id} value={level.id.toString()}>
                      {level.nama}
                    </option>
                  ))}
                </Select>
              </FormControl>
              <FormControl>
                <FormLabel>Question</FormLabel>
                <Textarea
                  value={formData.pertanyaan}
                  onChange={(e) => setFormData({ ...formData, pertanyaan: e.target.value })}
                />
              </FormControl>
              <FormControl>
                <FormLabel>Option A</FormLabel>
                <Input
                  value={formData.opsiA}
                  onChange={(e) => setFormData({ ...formData, opsiA: e.target.value })}
                />
              </FormControl>
              <FormControl>
                <FormLabel>Option B</FormLabel>
                <Input
                  value={formData.opsiB}
                  onChange={(e) => setFormData({ ...formData, opsiB: e.target.value })}
                />
              </FormControl>
              <FormControl>
                <FormLabel>Option C</FormLabel>
                <Input
                  value={formData.opsiC}
                  onChange={(e) => setFormData({ ...formData, opsiC: e.target.value })}
                />
              </FormControl>
              <FormControl>
                <FormLabel>Option D</FormLabel>
                <Input
                  value={formData.opsiD}
                  onChange={(e) => setFormData({ ...formData, opsiD: e.target.value })}
                />
              </FormControl>
              <FormControl>
                <FormLabel>Correct Answer</FormLabel>
                <Select
                  value={formData.jawabanBenar}
                  onChange={(e) => setFormData({ ...formData, jawabanBenar: e.target.value })}
                >
                  <option value="">Select Correct Answer</option>
                  <option value="A">A</option>
                  <option value="B">B</option>
                  <option value="C">C</option>
                  <option value="D">D</option>
                </Select>
              </FormControl>
              <FormControl>
                <FormLabel>Images (Optional)</FormLabel>
                <Input
                  type="file"
                  accept="image/*"
                  multiple
                  onChange={(e) => handleAddImages(e.target.files)}
                />
                <Text fontSize="sm" color="gray.600" mt={1}>
                  Select multiple images to add to this question
                </Text>
              </FormControl>

              {/* Display existing images */}
              {formData.existingImages.length > 0 && (
                <FormControl>
                  <FormLabel>Current Images</FormLabel>
                  <HStack spacing={3} wrap="wrap">
                    {formData.existingImages.map((img) => (
                      <Box key={img.id} position="relative" border="1px solid" borderColor="gray.200" borderRadius="md" p={2}>
                        <Image
                          src={`http://localhost:8080/${img.file_path}`}
                          alt="Existing"
                          boxSize="80px"
                          objectFit="cover"
                          borderRadius="md"
                        />
                        <Badge position="absolute" top="1" right="1" colorScheme="blue" fontSize="xs">
                          {img.urutan}
                        </Badge>
                        <IconButton
                          aria-label="Remove image"
                          icon={<CloseIcon />}
                          size="xs"
                          colorScheme="red"
                          position="absolute"
                          top="1"
                          left="1"
                          onClick={() => handleRemoveExistingImage(img.id)}
                        />
                      </Box>
                    ))}
                  </HStack>
                </FormControl>
              )}

              {editingQuestion && (
                <FormControl>
                  <FormLabel>Upload Additional Image</FormLabel>
                  <Input
                    type="file"
                    accept="image/*"
                    onChange={(e) => handleUploadImage(e.target.files?.[0] || null)}
                  />
                  <Text fontSize="sm" color="gray.600" mt={1}>
                    Upload a new image to this question
                  </Text>
                </FormControl>
              )}

              {/* Display new images to be added */}
              {formData.images.length > 0 && (
                <FormControl>
                  <FormLabel>New Images to Add</FormLabel>
                  <HStack spacing={3} wrap="wrap">
                    {formData.images.map((file, index) => (
                      <Box key={index} position="relative" border="1px solid" borderColor="green.200" borderRadius="md" p={2}>
                        <Image
                          src={URL.createObjectURL(file)}
                          alt={`New ${index + 1}`}
                          boxSize="80px"
                          objectFit="cover"
                          borderRadius="md"
                          onClick={() => openImageViewer(formData.images.map(f => ({ filePath: URL.createObjectURL(f) })), index)}
                        />
                        <Badge position="absolute" top="1" right="1" colorScheme="green" fontSize="xs">
                          New
                        </Badge>
                        <IconButton
                          aria-label="Remove new image"
                          icon={<CloseIcon />}
                          size="xs"
                          colorScheme="red"
                          position="absolute"
                          top="1"
                          left="1"
                          onClick={() => handleRemoveNewImage(index)}
                        />
                      </Box>
                    ))}
                  </HStack>
                </FormControl>
              )}
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button colorScheme="orange" mr={3} onClick={handleSubmit}>
              Save
            </Button>
            <Button variant="ghost" onClick={onClose}>
              Cancel
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

      {/* Image Viewer Modal */}
      <Modal isOpen={viewerOpen} onClose={closeImageViewer} size="xl" isCentered>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Image Viewer</ModalHeader>
          <ModalCloseButton />
          <ModalBody display="flex" alignItems="center" justifyContent="center">
            {viewerImages.length > 0 && (
              <Image src={viewerImages[viewerIndex]} alt={`Image ${viewerIndex + 1}`} maxH="70vh" />
            )}
          </ModalBody>
          <ModalFooter>
            <Button mr={3} onClick={prevImage} disabled={viewerImages.length <= 1}>Prev</Button>
            <Button mr={3} onClick={nextImage} disabled={viewerImages.length <= 1}>Next</Button>
            <Button variant="ghost" onClick={closeImageViewer}>Close</Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </Container>
  );
}