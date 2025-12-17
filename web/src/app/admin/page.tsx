import Link from 'next/link';
import { Box, Button, VStack, Heading, Container } from '@chakra-ui/react';

export default function AdminHome() {
  return (
    <Container maxW="container.md" py={10}>
      <Heading as="h1" size="xl" textAlign="center" mb={8}>
        CBT Admin Panel
      </Heading>
      <VStack spacing={4}>
        <Link href="/admin/levels">
          <Button colorScheme="blue" size="lg" width="full">
            Manage Levels
          </Button>
        </Link>
        <Link href="/admin/subjects">
          <Button colorScheme="green" size="lg" width="full">
            Manage Subjects
          </Button>
        </Link>
        <Link href="/admin/topics">
          <Button colorScheme="purple" size="lg" width="full">
            Manage Topics
          </Button>
        </Link>
        <Link href="/admin/questions">
          <Button colorScheme="orange" size="lg" width="full">
            Manage Questions
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