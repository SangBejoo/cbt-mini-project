'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { Box, Button, VStack, Heading, Container, Text, HStack } from '@chakra-ui/react';
import { useAuth } from '../auth-context';

export default function StudentHome() {
  const { user, logout, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && (!user || user.role !== 'SISWA')) {
      router.push('/login');
    }
  }, [user, isLoading, router]);

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  if (isLoading) {
    return (
      <Container maxW="container.lg" py={16}>
        <Box textAlign="center">Loading...</Box>
      </Container>
    );
  }

  if (!user || user.role !== 'SISWA') {
    return null; // Will redirect
  }

  return (
    <Container maxW="container.lg" py={16}>
      <HStack justify="space-between" mb={6}>
        <Box />
        <Button colorScheme="red" onClick={handleLogout}>
          Logout
        </Button>
      </HStack>

      <VStack spacing={12}>
        <Box textAlign="center">
          <Heading as="h1" size="3xl" color="orange.600" mb={4}>
            Portal Siswa CBT
          </Heading>
          <Text fontSize="xl" color="gray.600" maxW="2xl" mx="auto" mb={2}>
            Selamat datang, {user.nama}!
          </Text>
          <Text fontSize="lg" color="gray.500">
            Sistem Computer-Based Test - Pilih menu di bawah untuk memulai pembelajaran Anda.
          </Text>
        </Box>

        <Box
          bg="white"
          p={8}
          borderRadius="2xl"
          shadow="xl"
          border="1px solid"
          borderColor="gray.200"
          w="full"
          maxW="md"
        >
          <VStack spacing={6}>
            <Link href="/student/sessions" style={{ width: '100%' }}>
              <Button
                colorScheme="orange"
                size="lg"
                width="full"
                height="16"
                fontSize="lg"
                borderRadius="xl"
                shadow="md"
                _hover={{ shadow: 'lg', transform: 'translateY(-2px)' }}
                transition="all 0.2s"
              >
                Ikuti Tes
              </Button>
            </Link>

            <Link href="/student/history" style={{ width: '100%' }}>
              <Button
                colorScheme="orange"
                variant="outline"
                size="lg"
                width="full"
                height="16"
                fontSize="lg"
                borderRadius="xl"
                shadow="md"
                _hover={{ shadow: 'lg', transform: 'translateY(-2px)' }}
                transition="all 0.2s"
              >
                Lihat Riwayat
              </Button>
            </Link>
          </VStack>
        </Box>
      </VStack>
    </Container>
  );
}