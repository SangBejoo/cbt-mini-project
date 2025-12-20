import { useState, useCallback, useEffect } from 'react';
import { useToast } from '@chakra-ui/react';
import apiClient from '../services/api';
import { Question, Level, Subject, Topic } from '../types';

export interface UseQuestionsOptions {
  autoFetch?: boolean;
}

export function useQuestions(options: UseQuestionsOptions = { autoFetch: true }) {
  const [questions, setQuestions] = useState<Question[]>([]);
  const [levels, setLevels] = useState<Level[]>([]);
  const [subjects, setSubjects] = useState<Subject[]>([]);
  const [topics, setTopics] = useState<Topic[]>([]);
  const [loading, setLoading] = useState(true);
  const toast = useToast();

  const fetchLevels = useCallback(async () => {
    try {
      const response = await apiClient.get<any>('/levels');
      const data = Array.isArray(response.data)
        ? response.data
        : Array.isArray(response.data?.data)
        ? response.data.data
        : response.data?.tingkat || [];
      setLevels(data);
    } catch (error: any) {
      toast({
        title: 'Error',
        description: 'Gagal memuat tingkat',
        status: 'error',
      });
    }
  }, [toast]);

  const fetchSubjects = useCallback(async () => {
    try {
      const response = await apiClient.get<any>('/subjects');
      const data = Array.isArray(response.data)
        ? response.data
        : Array.isArray(response.data?.data)
        ? response.data.data
        : response.data?.mataPelajaran || [];
      setSubjects(data);
    } catch (error: any) {
      toast({
        title: 'Error',
        description: 'Gagal memuat mata pelajaran',
        status: 'error',
      });
    }
  }, [toast]);

  const fetchTopics = useCallback(async () => {
    try {
      const response = await apiClient.get<any>('/topics');
      const data = Array.isArray(response.data)
        ? response.data
        : Array.isArray(response.data?.data)
        ? response.data.data
        : response.data?.materi || [];
      setTopics(data);
    } catch (error: any) {
      toast({
        title: 'Error',
        description: 'Gagal memuat materi',
        status: 'error',
      });
    }
  }, [toast]);

  const fetchQuestions = useCallback(async () => {
    try {
      const response = await apiClient.get<any>('/questions');
      const data = Array.isArray(response.data)
        ? response.data
        : Array.isArray(response.data?.data)
        ? response.data.data
        : response.data?.soal || [];
      setQuestions(data);
    } catch (error: any) {
      toast({
        title: 'Error',
        description: 'Gagal memuat soal',
        status: 'error',
      });
    }
  }, [toast]);

  useEffect(() => {
    if (options.autoFetch !== false) {
      setLoading(true);
      Promise.all([
        fetchLevels(),
        fetchSubjects(),
        fetchTopics(),
        fetchQuestions(),
      ]).finally(() => setLoading(false));
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [options.autoFetch]); // Only depend on autoFetch flag

  const createQuestion = useCallback(
    async (data: Partial<Question>) => {
      try {
        const response = await apiClient.post<any>('/questions', data);
        const newQuestion = response.data?.data || response.data;
        setQuestions((prev) => [...prev, newQuestion]);
        toast({
          title: 'Berhasil',
          description: 'Soal berhasil dibuat',
          status: 'success',
        });
        return newQuestion;
      } catch (error: any) {
        const message =
          error.response?.data?.message || error.message || 'Gagal membuat soal';
        toast({
          title: 'Error',
          description: message,
          status: 'error',
        });
        throw error;
      }
    },
    [toast]
  );

  const updateQuestion = useCallback(
    async (id: number, data: Partial<Question>) => {
      try {
        const response = await apiClient.put<any>(`/questions/${id}`, data);
        const updatedQuestion = response.data?.data || response.data;
        setQuestions((prev) =>
          prev.map((q) => (q.id === id ? updatedQuestion : q))
        );
        toast({
          title: 'Berhasil',
          description: 'Soal berhasil diperbarui',
          status: 'success',
        });
        return updatedQuestion;
      } catch (error: any) {
        const message =
          error.response?.data?.message ||
          error.message ||
          'Gagal update soal';
        toast({
          title: 'Error',
          description: message,
          status: 'error',
        });
        throw error;
      }
    },
    [toast]
  );

  const deleteQuestion = useCallback(
    async (id: number) => {
      try {
        await apiClient.delete(`/questions/${id}`);
        setQuestions((prev) => prev.filter((q) => q.id !== id));
        toast({
          title: 'Berhasil',
          description: 'Soal berhasil dihapus',
          status: 'success',
        });
      } catch (error: any) {
        const message =
          error.response?.data?.message || error.message || 'Gagal hapus soal';
        toast({
          title: 'Error',
          description: message,
          status: 'error',
        });
        throw error;
      }
    },
    [toast]
  );

  const uploadImage = useCallback(
    async (questionId: number, files: FileList) => {
      try {
        const uploadPromises = Array.from(files).map(async (file, index) => {
          const base64 = await new Promise<string>((resolve, reject) => {
            const reader = new FileReader();
            reader.onload = () => resolve(reader.result as string);
            reader.onerror = reject;
            reader.readAsDataURL(file);
          });

          // Remove data:image/...;base64, prefix
          const imageBytes = base64.split(',')[1];

          return apiClient.post<any>(`/questions/${questionId}/images`, {
            imageBytes,
            namaFile: file.name,
            urutan: index,
            keterangan: '',
          });
        });

        await Promise.all(uploadPromises);

        toast({
          title: 'Berhasil',
          description: 'Gambar berhasil diupload',
          status: 'success',
        });
      } catch (error: any) {
        const message =
          error.response?.data?.message ||
          error.message ||
          'Gagal upload gambar';
        toast({
          title: 'Error',
          description: message,
          status: 'error',
        });
        throw error;
      }
    },
    [toast]
  );

  const deleteImage = useCallback(
    async (questionId: number, imageId: number) => {
      try {
        await apiClient.delete(`/questions/images/${imageId}`);
        setQuestions((prev) =>
          prev.map((q) =>
            q.id === questionId
              ? {
                  ...q,
                  gambar: q.gambar.filter((img) => img.id !== imageId),
                }
              : q
          )
        );
        toast({
          title: 'Berhasil',
          description: 'Gambar berhasil dihapus',
          status: 'success',
        });
      } catch (error: any) {
        const message =
          error.response?.data?.message ||
          error.message ||
          'Gagal hapus gambar';
        toast({
          title: 'Error',
          description: message,
          status: 'error',
        });
        throw error;
      }
    },
    [toast]
  );

  return {
    questions,
    levels,
    subjects,
    topics,
    loading,
    fetchQuestions,
    fetchLevels,
    fetchSubjects,
    fetchTopics,
    createQuestion,
    updateQuestion,
    deleteQuestion,
    uploadImage,
    deleteImage,
  };
}
