'use client';

import Link from 'next/link';
import { Box, Button, VStack, Heading, Container, Text } from '@chakra-ui/react';

export default function StudentHome() {
  return (
    <Container maxW="container.md" py={10}>
      <Heading as="h1" size="xl" textAlign="center" mb={8}>
        CBT Student Portal
      </Heading>
      <Text textAlign="center" mb={8} fontSize="lg">
        Welcome to the Computer-Based Test system. Choose an option below to get started.
      </Text>
      <VStack spacing={4}>
        <Link href="/student/sessions">
          <Button colorScheme="blue" size="lg" width="full">
            Take a Test
          </Button>
        </Link>
        <Link href="/student/history">
          <Button colorScheme="green" size="lg" width="full">
            View My History
          </Button>
        </Link>
        <Link href="/">
          <Button variant="outline" size="lg" width="full">
            Back to Home
          </Button>
        </Link>
      </VStack>
    </Container>
  );
}