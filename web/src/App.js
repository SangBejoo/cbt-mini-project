import React, { useState, useEffect } from 'react';
import {
  Box,
  Button,
  Card,
  CardBody,
  CardHeader,
  Container,
  Flex,
  FormControl,
  FormLabel,
  Grid,
  GridItem,
  Heading,
  Input,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  Text,
  Textarea,
  useDisclosure,
  useToast,
  VStack,
  HStack,
  IconButton,
  AlertDialog,
  AlertDialogBody,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogContent,
  AlertDialogOverlay,
  Spinner,
  Center
} from '@chakra-ui/react';
import { AddIcon, EditIcon, DeleteIcon, RepeatIcon } from '@chakra-ui/icons';

const API_BASE_URL = 'http://localhost:8000/v1/notes';

export default function NotesApp() {
  const [notes, setNotes] = useState([]);
  const [loading, setLoading] = useState(false);
  const [selectedNote, setSelectedNote] = useState(null);
  const [formData, setFormData] = useState({ title: '', content: '' });
  const [deleteId, setDeleteId] = useState(null);
  
  const { isOpen: isCreateOpen, onOpen: onCreateOpen, onClose: onCreateClose } = useDisclosure();
  const { isOpen: isEditOpen, onOpen: onEditOpen, onClose: onEditClose } = useDisclosure();
  const { isOpen: isDeleteOpen, onOpen: onDeleteOpen, onClose: onDeleteClose } = useDisclosure();
  
  const toast = useToast();
  const cancelRef = React.useRef();

  // Fetch all notes
  const fetchNotes = async () => {
    setLoading(true);
    try {
      const response = await fetch(API_BASE_URL);
      if (response.ok) {
        const data = await response.json();
        setNotes(data.notes || []);
        toast({
          title: 'Notes refreshed successfully',
          status: 'success',
          duration: 2000,
          isClosable: true,
        });
      } else {
        throw new Error('Failed to fetch notes');
      }
    } catch (error) {
      toast({
        title: 'Error fetching notes',
        description: error.message,
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  // Create new note
  const createNote = async () => {
    if (!formData.title.trim() || !formData.content.trim()) {
      toast({
        title: 'Please fill in all fields',
        status: 'warning',
        duration: 2000,
        isClosable: true,
      });
      return;
    }

    try {
      const response = await fetch(API_BASE_URL, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          note: {
            title: formData.title,
            content: formData.content,
          }
        }),
      });

      if (response.ok) {
        toast({
          title: 'Note created successfully',
          status: 'success',
          duration: 2000,
          isClosable: true,
        });
        setFormData({ title: '', content: '' });
        onCreateClose();
        fetchNotes();
      } else {
        throw new Error('Failed to create note');
      }
    } catch (error) {
      toast({
        title: 'Error creating note',
        description: error.message,
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
    }
  };

  // Update note
  const updateNote = async () => {
    if (!formData.title.trim() || !formData.content.trim()) {
      toast({
        title: 'Please fill in all fields',
        status: 'warning',
        duration: 2000,
        isClosable: true,
      });
      return;
    }

    try {
      console.log('Updating note with data:', formData); // Debug log
      const requestBody = {
        id: selectedNote.id,
        title: formData.title,
        content: formData.content,
      };
      console.log('Request body:', JSON.stringify(requestBody)); // Debug log

      const response = await fetch(`${API_BASE_URL}/${selectedNote.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestBody),
      });

      if (response.ok) {
        toast({
          title: 'Note updated successfully',
          status: 'success',
          duration: 2000,
          isClosable: true,
        });
        setFormData({ title: '', content: '' });
        setSelectedNote(null);
        onEditClose();
        fetchNotes();
      } else {
        const errorText = await response.text();
        console.error('Update error response:', errorText);
        throw new Error(`Failed to update note: ${response.status}`);
      }
    } catch (error) {
      console.error('Update error:', error);
      toast({
        title: 'Error updating note',
        description: error.message,
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
    }
  };

  // Delete note
  const deleteNote = async () => {
    try {
      console.log('Deleting note with ID:', deleteId); // Debug log
      
      const response = await fetch(`${API_BASE_URL}/${deleteId}`, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        toast({
          title: 'Note deleted successfully',
          status: 'success',
          duration: 2000,
          isClosable: true,
        });
        setDeleteId(null);
        onDeleteClose();
        fetchNotes();
      } else {
        const errorText = await response.text();
        console.error('Delete error response:', errorText);
        throw new Error(`Failed to delete note: ${response.status}`);
      }
    } catch (error) {
      console.error('Delete error:', error);
      toast({
        title: 'Error deleting note',
        description: error.message,
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
    }
  };

  // Handle edit note
  const handleEdit = (note) => {
    console.log('Editing note:', note); // Debug log
    setSelectedNote(note);
    setFormData({ title: note.title || '', content: note.content || '' });
    onEditOpen();
  };

  // Handle delete note
  const handleDelete = (id) => {
    setDeleteId(id);
    onDeleteOpen();
  };

  // Handle create note
  const handleCreate = () => {
    setFormData({ title: '', content: '' });
    onCreateOpen();
  };

  // Load notes on component mount
  useEffect(() => {
    fetchNotes();
  }, []);

  return (
    <Container maxW="container.xl" py={8}>
      <VStack spacing={6} align="stretch">
        {/* Header */}
        <Flex justify="space-between" align="center">
          <Heading size="xl" color="blue.600">
            My Notes
          </Heading>
          <HStack spacing={4}>
            <Button
              leftIcon={<RepeatIcon />}
              colorScheme="blue"
              variant="outline"
              onClick={fetchNotes}
              isLoading={loading}
              loadingText="Refreshing..."
            >
              Refresh
            </Button>
            <Button
              leftIcon={<AddIcon />}
              colorScheme="blue"
              onClick={handleCreate}
            >
              Add Note
            </Button>
          </HStack>
        </Flex>

        {/* Notes Grid */}
        {loading ? (
          <Center py={10}>
            <VStack spacing={4}>
              <Spinner size="xl" color="blue.500" />
              <Text>Loading notes...</Text>
            </VStack>
          </Center>
        ) : notes.length === 0 ? (
          <Center py={10}>
            <VStack spacing={4}>
              <Text fontSize="lg" color="gray.500">
                No notes found
              </Text>
              <Button
                leftIcon={<AddIcon />}
                colorScheme="blue"
                onClick={handleCreate}
              >
                Create your first note
              </Button>
            </VStack>
          </Center>
        ) : (
          <Grid templateColumns="repeat(auto-fill, minmax(300px, 1fr))" gap={6}>
            {notes.map((note) => (
              <GridItem key={note.id}>
                <Card h="full" shadow="md" _hover={{ shadow: 'lg' }}>
                  <CardHeader pb={2}>
                    <Flex justify="space-between" align="start">
                      <Heading size="md" noOfLines={2} flex={1} mr={2}>
                        {note.title}
                      </Heading>
                      <HStack spacing={1}>
                        <IconButton
                          icon={<EditIcon />}
                          size="sm"
                          colorScheme="green"
                          variant="ghost"
                          onClick={() => handleEdit(note)}
                          aria-label="Edit note"
                        />
                        <IconButton
                          icon={<DeleteIcon />}
                          size="sm"
                          colorScheme="red"
                          variant="ghost"
                          onClick={() => handleDelete(note.id)}
                          aria-label="Delete note"
                        />
                      </HStack>
                    </Flex>
                  </CardHeader>
                  <CardBody pt={0}>
                    <Text color="gray.600" noOfLines={4}>
                      {note.content}
                    </Text>
                    <Text fontSize="sm" color="gray.400" mt={4}>
                      ID: {note.id}
                    </Text>
                  </CardBody>
                </Card>
              </GridItem>
            ))}
          </Grid>
        )}

        {/* Create Note Modal */}
        <Modal isOpen={isCreateOpen} onClose={onCreateClose} size="lg">
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>Create New Note</ModalHeader>
            <ModalCloseButton />
            <ModalBody>
              <VStack spacing={4}>
                <FormControl isRequired>
                  <FormLabel>Title</FormLabel>
                  <Input
                    value={formData.title}
                    onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                    placeholder="Enter note title"
                  />
                </FormControl>
                <FormControl isRequired>
                  <FormLabel>Content</FormLabel>
                  <Textarea
                    value={formData.content}
                    onChange={(e) => setFormData({ ...formData, content: e.target.value })}
                    placeholder="Enter note content"
                    rows={6}
                  />
                </FormControl>
              </VStack>
            </ModalBody>
            <ModalFooter>
              <Button variant="ghost" mr={3} onClick={onCreateClose}>
                Cancel
              </Button>
              <Button colorScheme="blue" onClick={createNote}>
                Create Note
              </Button>
            </ModalFooter>
          </ModalContent>
        </Modal>

        {/* Edit Note Modal */}
        <Modal isOpen={isEditOpen} onClose={onEditClose} size="lg">
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>Edit Note</ModalHeader>
            <ModalCloseButton />
            <ModalBody>
              <VStack spacing={4}>
                <FormControl isRequired>
                  <FormLabel>Title</FormLabel>
                  <Input
                    value={formData.title}
                    onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                    placeholder="Enter note title"
                  />
                </FormControl>
                <FormControl isRequired>
                  <FormLabel>Content</FormLabel>
                  <Textarea
                    value={formData.content}
                    onChange={(e) => setFormData({ ...formData, content: e.target.value })}
                    placeholder="Enter note content"
                    rows={6}
                  />
                </FormControl>
              </VStack>
            </ModalBody>
            <ModalFooter>
              <Button variant="ghost" mr={3} onClick={onEditClose}>
                Cancel
              </Button>
              <Button colorScheme="green" onClick={updateNote}>
                Update Note
              </Button>
            </ModalFooter>
          </ModalContent>
        </Modal>

        {/* Delete Confirmation Dialog */}
        <AlertDialog
          isOpen={isDeleteOpen}
          leastDestructiveRef={cancelRef}
          onClose={onDeleteClose}
        >
          <AlertDialogOverlay>
            <AlertDialogContent>
              <AlertDialogHeader fontSize="lg" fontWeight="bold">
                Delete Note
              </AlertDialogHeader>
              <AlertDialogBody>
                Are you sure you want to delete this note? This action cannot be undone.
              </AlertDialogBody>
              <AlertDialogFooter>
                <Button ref={cancelRef} onClick={onDeleteClose}>
                  Cancel
                </Button>
                <Button colorScheme="red" onClick={deleteNote} ml={3}>
                  Delete
                </Button>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialogOverlay>
        </AlertDialog>
      </VStack>
    </Container>
  );
}