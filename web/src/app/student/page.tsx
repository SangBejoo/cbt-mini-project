'use client';

import Link from 'next/link';
import { Box, Button, VStack, Heading, Container, Text } from '@chakra-ui/react';

export default function StudentHome() {
  return (
    <Container maxW="container.lg" py={16}>
      <VStack spacing={12}>
        <Box textAlign="center">
          <Heading as="h1" size="3xl" color="orange.600" mb={4}>
            Portal Siswa CBT
          </Heading>
          <Text fontSize="xl" color="gray.600" maxW="2xl" mx="auto">
            Selamat datang di sistem Computer-Based Test. Pilih menu di bawah untuk memulai pembelajaran Anda.
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
                Lihat Riwayat Siswa/i
              </Button>
            </Link>

            <Link href="/" style={{ width: '100%' }}>
              <Button
                variant="outline"
                size="lg"
                width="full"
                height="14"
                fontSize="lg"
                borderRadius="xl"
                borderWidth="2px"
                _hover={{ bg: 'gray.50', borderColor: 'gray.300' }}
                transition="all 0.2s"
              >
                Kembali ke Beranda
              </Button>
            </Link>
          </VStack>
        </Box>
      </VStack>
    </Container>
  );
}