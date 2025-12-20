"use client";

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { Box, Button, VStack, Heading, Container, Tabs, TabList, Tab, TabPanels, TabPanel, HStack, Text } from '@chakra-ui/react';
import { useAuth } from '../auth-context';
import LevelsTab from './components/LevelsTab';
const SubjectsTab = dynamic(() => import('./components/SubjectsTab'), { ssr: false });
import TopicsTab from './components/TopicsTab';
import dynamic from 'next/dynamic';

const QuestionsTab = dynamic(() => import('./components/QuestionsTab'), { ssr: false });
import UsersTab from './components/UsersTab';
import HistoryTab from './components/HistoryTab';

export default function AdminHome() {
  const { user, logout, isLoading } = useAuth();
  const router = useRouter();
  const [activeTab, setActiveTab] = useState(0);

  useEffect(() => {
    if (!isLoading && (!user || user.role !== 'ADMIN')) {
      router.push('/login');
    }
  }, [user, isLoading, router]);

  useEffect(() => {
    // Load active tab from localStorage
    const savedTab = localStorage.getItem('adminActiveTab');
    if (savedTab) {
      setActiveTab(parseInt(savedTab, 10));
    }
  }, []);

  const handleTabChange = (index: number) => {
    setActiveTab(index);
    localStorage.setItem('adminActiveTab', index.toString());
  };

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

      <Tabs variant="enclosed" colorScheme="blue" isLazy index={activeTab} onChange={handleTabChange}>
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