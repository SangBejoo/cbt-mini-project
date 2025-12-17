'use client';

import Link from 'next/link';
import { Box, Button, VStack, Heading, Container, Text } from '@chakra-ui/react';
import { useEffect, useState } from 'react';

export default function Home() {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) {
    return (
      <Box py={10}>
        <Heading as="h1" size="xl" textAlign="center" mb={8}>
          CBT System
        </Heading>
      </Box>
    );
  }

  return (
    <Container maxW="container.md" py={10}>
      <Heading as="h1" size="xl" textAlign="center" mb={8}>
        CBT System
      </Heading>
      <Text textAlign="center" mb={8} fontSize="lg">
        Computer-Based Test Management System
      </Text>
      <VStack spacing={6}>
        <Link href="/admin">
          <Button colorScheme="blue" size="lg" width="full">
            Admin Panel
          </Button>
        </Link>
        <Link href="/student">
          <Button colorScheme="green" size="lg" width="full">
            Student Portal
          </Button>
        </Link>
      </VStack>
    </Container>
  );
}
