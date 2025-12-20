'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import {
  Box,
  Button,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  IconButton,
  useDisclosure,
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
  useToast,
  HStack,
  VStack,
  Text,
  Badge,
  Heading,
  SimpleGrid,
  Divider,
  Switch,
  FormErrorMessage,
  Flex,
  Spacer,
} from '@chakra-ui/react';
import { EditIcon, DeleteIcon, AddIcon, ChevronLeftIcon, ChevronRightIcon } from '@chakra-ui/icons';
import axios from 'axios';

interface User {
  id: number;
  email: string;
  nama: string;
  role: string; // Change from number to string to match API response
  isActive: boolean;
  createdAt: string | null;
  updatedAt: string | null;
}

interface CreateUserData {
  email: string;
  password: string;
  nama: string;
  role: string; // Change to string
}

interface UpdateUserData {
  id: number;
  email: string;
  nama: string;
  role: string; // Change to string
  isActive: boolean;
}

interface PaginationInfo {
  totalCount: number;
  totalPages: number;
  currentPage: number;
  pageSize: number;
}

const API_BASE = process.env.NEXT_PUBLIC_API_BASE + '/v1/auth';

export default function UsersTab() {
  const toast = useToast();

  // --- State ---
  const [users, setUsers] = useState<User[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedRole, setSelectedRole] = useState<string>('all');
  const [selectedStatus, setSelectedStatus] = useState<string>('all');
  const [pagination, setPagination] = useState<PaginationInfo>({
    totalCount: 0,
    totalPages: 0,
    currentPage: 1,
    pageSize: 10,
  });
  const [isAfterDelete, setIsAfterDelete] = useState(false);

  // Modal states
  const { isOpen: isCreateOpen, onOpen: onCreateOpen, onClose: onCreateClose } = useDisclosure();
  const { isOpen: isEditOpen, onOpen: onEditOpen, onClose: onEditClose } = useDisclosure();
  const { isOpen: isDeleteOpen, onOpen: onDeleteOpen, onClose: onDeleteClose } = useDisclosure();

  // Form states
  const [createForm, setCreateForm] = useState<CreateUserData>({
    email: '',
    password: '',
    nama: '',
    role: 'SISWA',
  });
  const [editForm, setEditForm] = useState<UpdateUserData>({
    id: 0,
    email: '',
    nama: '',
    role: 'SISWA',
    isActive: true,
  });
  const [currentDeleteId, setCurrentDeleteId] = useState<number | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);
  const toastIdRef = useRef<string | number | null>(null);

  // Form validation
  const [createErrors, setCreateErrors] = useState<{[key: string]: string}>({});
  const [editErrors, setEditErrors] = useState<{[key: string]: string}>({});

  // --- Fetch Data ---
  const fetchUsers = async (page: number = 1) => {
    try {
      setIsLoading(true);
      const params = new URLSearchParams();
      
      // Add pagination parameters
      params.append('page', page.toString());
      params.append('page_size', pagination.pageSize.toString());
      
      // Add role filter if selected
      if (selectedRole !== 'all') {
        params.append('role', selectedRole === 'siswa' ? 'SISWA' : 'ADMIN');
      }
      
      // Add status filter if selected
      if (selectedStatus !== 'all') {
        const statusValue = selectedStatus === 'active' ? 1 : 2;
        params.append('status_filter', statusValue.toString());
      }

      const response = await axios.get(`${API_BASE}/users?${params.toString()}`);
      
      if (response.data.success) {
        const fetchedUsers = response.data.users || [];
        setUsers(fetchedUsers);
        
        // Update pagination info
        if (response.data.pagination) {
          setPagination({
            totalCount: response.data.pagination.total_count || 0,
            totalPages: response.data.pagination.total_pages || 0,
            currentPage: response.data.pagination.current_page || 1,
            pageSize: response.data.pagination.page_size || 10,
          });
        }

        // If no users returned after delete and not on first page, go to previous page
        if (isAfterDelete && fetchedUsers.length === 0 && page > 1) {
          setIsAfterDelete(false);
          fetchUsers(page - 1);
        } else {
          setIsAfterDelete(false);
        }
      }
    } catch (error) {
      console.error('Failed to fetch users', error);
      toast({ title: 'Gagal memuat data user', status: 'error' });
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchUsers(1); // Reset to page 1 when filters change
  }, [selectedRole, selectedStatus]);

  // Handle page change
  const handlePageChange = (newPage: number) => {
    if (newPage >= 1 && newPage <= pagination.totalPages) {
      fetchUsers(newPage);
    }
  };

  // --- Form Handlers ---
  const validateCreateForm = (): boolean => {
    const errors: {[key: string]: string} = {};

    if (!createForm.email) errors.email = 'Email wajib diisi';
    if (!createForm.password) errors.password = 'Password wajib diisi';
    if (!createForm.nama) errors.nama = 'Nama wajib diisi';

    setCreateErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const validateEditForm = (): boolean => {
    const errors: {[key: string]: string} = {};

    if (!editForm.email) errors.email = 'Email wajib diisi';
    if (!editForm.nama) errors.nama = 'Nama wajib diisi';

    setEditErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleCreateUser = async () => {
    if (!validateCreateForm()) return;

    try {
      const response = await axios.post(`${API_BASE}/users`, createForm);
      if (response.data.success) {
        toast({ title: 'User berhasil dibuat', status: 'success' });
        fetchUsers(1); // Reset to first page after create
        onCreateClose();
        setCreateForm({ email: '', password: '', nama: '', role: 'SISWA' });
        setCreateErrors({});
      } else {
        toast({ title: response.data.message || 'Gagal membuat user', status: 'error' });
      }
    } catch (error: any) {
      toast({ title: error.response?.data?.message || 'Error membuat user', status: 'error' });
    }
  };

  const handleEditUser = async () => {
    if (!validateEditForm()) return;

    try {
      const { id, ...updateData } = editForm;
      const response = await axios.put(`${API_BASE}/users/${id}`, updateData);
      if (response.data.success) {
        toast({ title: 'User berhasil diupdate', status: 'success' });
        fetchUsers(pagination.currentPage); // Stay on current page after update
        onEditClose();
        setEditErrors({});
      } else {
        toast({ title: response.data.message || 'Gagal update user', status: 'error' });
      }
    } catch (error: any) {
      toast({ title: error.response?.data?.message || 'Error update user', status: 'error' });
    }
  };

  const handleDeleteUser = async () => {
    if (!currentDeleteId || isDeleting) return;

    setIsDeleting(true);
    try {
      const response = await axios.delete(`${API_BASE}/users/${currentDeleteId}`);
      
      // Always close modal and reset state first
      onDeleteClose();
      setCurrentDeleteId(null);
      setIsDeleting(false);
      
      if (response.data.success || response.status === 200) {
        // Close any previous toast
        if (toastIdRef.current) {
          toast.close(toastIdRef.current);
        }
        
        // Show success toast once
        toastIdRef.current = toast({ 
          title: 'User berhasil dihapus', 
          status: 'success',
          duration: 2000,
          isClosable: true,
          position: 'top-right',
        });
        
        // Refresh data immediately
        setIsAfterDelete(true);
        fetchUsers(pagination.currentPage);
      } else {
        toast({ 
          title: response.data.message || 'Gagal hapus user', 
          status: 'error',
          position: 'top-right',
          duration: 4000,
          isClosable: true,
        });
      }
    } catch (error: any) {
      // Close modal and reset state even on error
      onDeleteClose();
      setCurrentDeleteId(null);
      setIsDeleting(false);
      
      toast({ 
        title: error.response?.data?.message || 'Error hapus user', 
        status: 'error',
        position: 'top-right',
        duration: 4000,
        isClosable: true,
      });
    }
  };

  const openEditModal = (user: User) => {
    setEditForm({
      id: user.id,
      email: user.email,
      nama: user.nama,
      role: user.role,
      isActive: user.isActive,
    });
    onEditOpen();
  };

  const openDeleteModal = (userId: number) => {
    setCurrentDeleteId(userId);
    onDeleteOpen();
  };

  // --- Helpers ---
  const getRoleLabel = (role: string) => {
    return role;
  };

  const getRoleColor = (role: string) => {
    return role === 'SISWA' ? 'blue' : role === 'ADMIN' ? 'red' : 'gray';
  };

  if (isLoading) {
    return <Box textAlign="center" py={8}>Loading users...</Box>;
  }

  return (
    <Box>
      <HStack justify="space-between" mb={6}>
        <Heading size="lg">User Management</Heading>
        <Button leftIcon={<AddIcon />} colorScheme="blue" onClick={onCreateOpen}>
          Tambah User
        </Button>
      </HStack>

      {/* Filters */}
      <HStack spacing={4} mb={6}>
        <FormControl maxW="200px">
          <FormLabel>Role</FormLabel>
          <Select value={selectedRole} onChange={(e) => setSelectedRole(e.target.value)}>
            <option value="all">Semua Role</option>
            <option value="siswa">Siswa</option>
            <option value="admin">Admin</option>
          </Select>
        </FormControl>

        <FormControl maxW="200px">
          <FormLabel>Status</FormLabel>
          <Select value={selectedStatus} onChange={(e) => setSelectedStatus(e.target.value)}>
            <option value="all">Semua Status</option>
            <option value="active">Aktif</option>
            <option value="inactive">Tidak Aktif</option>
          </Select>
        </FormControl>
      </HStack>

      {/* Users Table */}
      <Box overflowX="auto">
        <Table variant="simple">
          <Thead>
            <Tr>
              <Th>Nama</Th>
              <Th>Email</Th>
              <Th>Role</Th>
              <Th>Status</Th>
              <Th>Dibuat</Th>
              <Th>Actions</Th>
            </Tr>
          </Thead>
          <Tbody>
            {users.map((user) => (
              <Tr key={user.id}>
                <Td>{user.nama}</Td>
                <Td>{user.email}</Td>
                <Td>
                  <Badge colorScheme={getRoleColor(user.role)}>
                    {getRoleLabel(user.role)}
                  </Badge>
                </Td>
                <Td>
                  <Badge colorScheme={user.isActive ? 'green' : 'red'}>
                    {user.isActive ? 'Aktif' : 'Tidak Aktif'}
                  </Badge>
                </Td>
                <Td>{user.createdAt ? new Date(user.createdAt).toLocaleDateString('id-ID') : '-'}</Td>
                <Td>
                  <HStack spacing={2}>
                    <IconButton
                      aria-label="Edit user"
                      icon={<EditIcon />}
                      size="sm"
                      onClick={() => openEditModal(user)}
                    />
                    <IconButton
                      aria-label="Delete user"
                      icon={<DeleteIcon />}
                      size="sm"
                      colorScheme="red"
                      onClick={() => openDeleteModal(user.id)}
                    />
                  </HStack>
                </Td>
              </Tr>
            ))}
          </Tbody>
        </Table>
      </Box>

      {users.length === 0 && (
        <Box textAlign="center" py={8}>
          <Text>Tidak ada user ditemukan</Text>
        </Box>
      )}

      {/* Pagination */}
      {pagination.totalPages > 1 && (
        <Flex justify="space-between" align="center" mt={6}>
          <Text fontSize="sm" color="gray.600">
            Menampilkan {users.length} dari {pagination.totalCount} user
          </Text>
          
          <HStack spacing={2}>
            <Button
              size="sm"
              variant="outline"
              leftIcon={<ChevronLeftIcon />}
              isDisabled={pagination.currentPage <= 1}
              onClick={() => handlePageChange(pagination.currentPage - 1)}
            >
              Previous
            </Button>
            
            <HStack spacing={1}>
              {Array.from({ length: Math.min(5, pagination.totalPages) }, (_, i) => {
                let pageNum;
                if (pagination.totalPages <= 5) {
                  pageNum = i + 1;
                } else if (pagination.currentPage <= 3) {
                  pageNum = i + 1;
                } else if (pagination.currentPage >= pagination.totalPages - 2) {
                  pageNum = pagination.totalPages - 4 + i;
                } else {
                  pageNum = pagination.currentPage - 2 + i;
                }
                
                return (
                  <Button
                    key={pageNum}
                    size="sm"
                    variant={pageNum === pagination.currentPage ? "solid" : "outline"}
                    colorScheme={pageNum === pagination.currentPage ? "blue" : "gray"}
                    onClick={() => handlePageChange(pageNum)}
                  >
                    {pageNum}
                  </Button>
                );
              })}
            </HStack>
            
            <Button
              size="sm"
              variant="outline"
              rightIcon={<ChevronRightIcon />}
              isDisabled={pagination.currentPage >= pagination.totalPages}
              onClick={() => handlePageChange(pagination.currentPage + 1)}
            >
              Next
            </Button>
          </HStack>
        </Flex>
      )}

      {/* Create User Modal */}
      <Modal isOpen={isCreateOpen} onClose={onCreateClose} size="lg">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Tambah User Baru</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4}>
              <FormControl isInvalid={!!createErrors.nama}>
                <FormLabel>Nama</FormLabel>
                <Input
                  value={createForm.nama}
                  onChange={(e) => setCreateForm({ ...createForm, nama: e.target.value })}
                  placeholder="Masukkan nama lengkap"
                />
                <FormErrorMessage>{createErrors.nama}</FormErrorMessage>
              </FormControl>

              <FormControl isInvalid={!!createErrors.email}>
                <FormLabel>Email</FormLabel>
                <Input
                  type="email"
                  value={createForm.email}
                  onChange={(e) => setCreateForm({ ...createForm, email: e.target.value })}
                  placeholder="Masukkan email"
                />
                <FormErrorMessage>{createErrors.email}</FormErrorMessage>
              </FormControl>

              <FormControl isInvalid={!!createErrors.password}>
                <FormLabel>Password</FormLabel>
                <Input
                  type="password"
                  value={createForm.password}
                  onChange={(e) => setCreateForm({ ...createForm, password: e.target.value })}
                  placeholder="Masukkan password"
                />
                <FormErrorMessage>{createErrors.password}</FormErrorMessage>
              </FormControl>

              <FormControl>
                <FormLabel>Role</FormLabel>
                <Select
                  value={createForm.role}
                  onChange={(e) => setCreateForm({ ...createForm, role: e.target.value })}
                >
                  <option value="SISWA">Siswa</option>
                  <option value="ADMIN">Admin</option>
                </Select>
              </FormControl>
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={onCreateClose}>
              Batal
            </Button>
            <Button colorScheme="blue" onClick={handleCreateUser}>
              Simpan
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

      {/* Edit User Modal */}
      <Modal isOpen={isEditOpen} onClose={onEditClose} size="lg">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Edit User</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4}>
              <FormControl isInvalid={!!editErrors.nama}>
                <FormLabel>Nama</FormLabel>
                <Input
                  value={editForm.nama}
                  onChange={(e) => setEditForm({ ...editForm, nama: e.target.value })}
                  placeholder="Masukkan nama lengkap"
                />
                <FormErrorMessage>{editErrors.nama}</FormErrorMessage>
              </FormControl>

              <FormControl isInvalid={!!editErrors.email}>
                <FormLabel>Email</FormLabel>
                <Input
                  type="email"
                  value={editForm.email}
                  onChange={(e) => setEditForm({ ...editForm, email: e.target.value })}
                  placeholder="Masukkan email"
                />
                <FormErrorMessage>{editErrors.email}</FormErrorMessage>
              </FormControl>

              <FormControl>
                <FormLabel>Role</FormLabel>
                <Select
                  value={editForm.role}
                  onChange={(e) => setEditForm({ ...editForm, role: parseInt(e.target.value) })}
                >
                  <option value={1}>Siswa</option>
                  <option value={2}>Admin</option>
                </Select>
              </FormControl>

              <FormControl>
                <FormLabel>Status Aktif</FormLabel>
                <Switch
                  isChecked={editForm.isActive}
                  onChange={(e) => setEditForm({ ...editForm, isActive: e.target.checked })}
                />
              </FormControl>
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={onEditClose}>
              Batal
            </Button>
            <Button colorScheme="blue" onClick={handleEditUser}>
              Update
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal isOpen={isDeleteOpen} onClose={onDeleteClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Konfirmasi Hapus</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <Text>Apakah Anda yakin ingin menghapus user ini? Tindakan ini tidak dapat dibatalkan.</Text>
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={onDeleteClose}>
              Batal
            </Button>
            <Button colorScheme="red" onClick={handleDeleteUser} isLoading={isDeleting} disabled={isDeleting}>
              Hapus
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </Box>
  );
}