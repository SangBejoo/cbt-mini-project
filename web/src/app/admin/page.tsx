"use client";

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Box, Button, VStack, Heading, Container, Tabs, TabList, Tab, TabPanels, TabPanel, HStack, Text } from '@chakra-ui/react';
import { useAuth } from '../auth-context';
import LevelsTab from './components/LevelsTab';
import SubjectsTab from './components/SubjectsTab';
import TopicsTab from './components/TopicsTab';
import QuestionsTab from './components/QuestionsTab';
import UsersTab from './components/UsersTab';
import HistoryTab from './components/HistoryTab';

export default function AdminHome() {
  const { user, logout, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && (!user || user.role !== 'ADMIN')) {
      router.push('/login');
    }
  }, [user, isLoading, router]);

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  if (isLoading) {
    return (
      <Container maxW="container.xl" py={10}>
        <Box textAlign="center">Loading...</Box>
      </Container>
    );
  }

  if (!user || user.role !== 'ADMIN') {
    return null; // Will redirect
  }

  return (
    <Container maxW="container.xl" py={10}>
      <HStack justify="space-between" mb={6}>
        <Box>
          <Heading as="h1" size="xl">
            Panel Admin CBT
          </Heading>
          <Text color="gray.600">Welcome, {user.nama}</Text>
        </Box>
        <Button colorScheme="red" onClick={handleLogout}>
          Logout
        </Button>
      </HStack>

      <Tabs variant="enclosed" colorScheme="blue" isLazy>
        <TabList>
          <Tab>Tingkat</Tab>
          <Tab>Mata Pelajaran</Tab>
          <Tab>Materi</Tab>
          <Tab>Soal</Tab>
          <Tab>Users</Tab>
          <Tab>History</Tab>
        </TabList>
        <TabPanels>
          <TabPanel>
            <LevelsTab />
          </TabPanel>
          <TabPanel>
            <SubjectsTab />
          </TabPanel>
          <TabPanel>
            <TopicsTab />
          </TabPanel>
          <TabPanel>
            <QuestionsTab />
          </TabPanel>
          <TabPanel>
            <UsersTab />
          </TabPanel>
          <TabPanel>
            <HistoryTab />
          </TabPanel>
        </TabPanels>
      </Tabs>
    </Container>
  );
}