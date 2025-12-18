"use client";

import Link from 'next/link';
import { Box, Button, VStack, Heading, Container, Tabs, TabList, Tab, TabPanels, TabPanel } from '@chakra-ui/react';
import LevelsTab from './components/LevelsTab';
import SubjectsTab from './components/SubjectsTab';
import TopicsTab from './components/TopicsTab';
import QuestionsTab from './components/QuestionsTab';

export default function AdminHome() {
  return (
    <Container maxW="container.xl" py={10}>
      <Heading as="h1" size="xl" textAlign="center" mb={8}>
        Panel Admin CBT
      </Heading>
      <Tabs variant="enclosed" colorScheme="blue" isLazy>
        <TabList>
          <Tab>Tingkat</Tab>
          <Tab>Mata Pelajaran</Tab>
          <Tab>Materi</Tab>
          <Tab>Soal</Tab>
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
        </TabPanels>
      </Tabs>
      <Box mt={8} textAlign="center">
        <Link href="/">
          <Button variant="outline">
            Kembali ke Beranda
          </Button>
        </Link>
      </Box>
    </Container>
  );
}