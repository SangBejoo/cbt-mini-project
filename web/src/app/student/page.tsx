'use client';

import Link from 'next/link';
import { Box, Button, VStack, Heading, Container, Text } from '@chakra-ui/react';

export default function StudentHome() {
  return (
    <Container maxW="container.lg" py={16}>
      <VStack spacing={12}>
        <Box textAlign="center">
          <Heading as="h1" size="3xl" color="blue.600" mb={4}>
            🎓 CBT Student Portal
          </Heading>
          <Text fontSize="xl" color="gray.600" maxW="2xl" mx="auto">
            Welcome to the Computer-Based Test system. Choose an option below to get started with your learning journey.
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
                colorScheme="blue"
                size="lg"
                width="full"
                height="16"
                fontSize="lg"
                leftIcon={<Text fontSize="2xl">📝</Text>}
                borderRadius="xl"
                shadow="md"
                _hover={{ shadow: 'lg', transform: 'translateY(-2px)' }}
                transition="all 0.2s"
              >
                Take a Test
              </Button>
            </Link>

            <Link href="/student/history" style={{ width: '100%' }}>
              <Button
                colorScheme="green"
                size="lg"
                width="full"
                height="16"
                fontSize="lg"
                leftIcon={<Text fontSize="2xl">📊</Text>}
                borderRadius="xl"
                shadow="md"
                _hover={{ shadow: 'lg', transform: 'translateY(-2px)' }}
                transition="all 0.2s"
              >
                View My History
              </Button>
            </Link>

            <Link href="/" style={{ width: '100%' }}>
              <Button
                variant="outline"
                size="lg"
                width="full"
                height="14"
                fontSize="lg"
                leftIcon={<Text fontSize="xl">🏠</Text>}
                borderRadius="xl"
                borderWidth="2px"
                _hover={{ bg: 'gray.50', borderColor: 'gray.300' }}
                transition="all 0.2s"
              >
                Back to Home
              </Button>
            </Link>
          </VStack>
        </Box>
      </VStack>
    </Container>
  );
}