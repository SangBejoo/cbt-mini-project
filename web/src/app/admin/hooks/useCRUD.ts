import { useState, useCallback, useEffect, useMemo } from 'react';
import { useToast } from '@chakra-ui/react';
import apiClient from '../services/api';
import { BaseEntity, ApiResponse } from '../types';
import { AxiosError } from 'axios';

export interface UseCRUDOptions {
  onSuccess?: (data: any) => void;
  onError?: (error: AxiosError) => void;
  autoFetch?: boolean;
}

export function useCRUD<T extends BaseEntity>(
  endpoint: string,
  options: UseCRUDOptions = { autoFetch: true }
) {
  const [data, setData] = useState<T[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const toast = useToast();

  const handleError = useCallback((err: any, defaultMessage: string) => {
    const message = err.response?.data?.message || err.message || defaultMessage;
    setError(message);
    toast({
      title: 'Error',
      description: message,
      status: 'error',
      duration: 4000,
      isClosable: true,
    });
    options.onError?.(err);
  }, [toast, options]);

  const handleSuccess = useCallback((message: string, newData?: any) => {
    toast({
      title: 'Berhasil',
      description: message,
      status: 'success',
      duration: 3000,
      isClosable: true,
    });
    setError(null);
    options.onSuccess?.(newData);
  }, [toast, options]);

  const fetch = useCallback(async () => {
    setLoading(true);
    try {
      const res = await apiClient.get<any>(endpoint);

      // Handle different response formats based on endpoint
      let items: T[] = [];
      if (endpoint === 'levels') {
        items = res.data?.tingkat || [];
      } else if (endpoint === 'subjects') {
        items = res.data?.mataPelajaran || [];
      } else if (endpoint === 'topics') {
        items = res.data?.materi || [];
      } else if (endpoint === 'auth/users') {
        items = res.data?.users || [];
      } else {
        // Default parsing for other endpoints
        items = Array.isArray(res.data)
          ? res.data
          : Array.isArray(res.data?.data)
          ? res.data.data
          : Array.isArray(res.data?.items)
          ? res.data.items
          : [];
      }

      console.log('Parsed items:', items);
      setData(items);
      setError(null);
    } catch (err: any) {
      console.error('API Error:', err);
      handleError(err, 'Gagal mengambil data');
    } finally {
      setLoading(false);
    }
  }, [endpoint, handleError]);

  const create = useCallback(
    async (payload: Omit<T, 'id'>) => {
      try {
        const res = await apiClient.post<any>(endpoint, payload);

        // Handle different response formats based on endpoint
        let newItem: T;
        if (endpoint === 'levels') {
          newItem = res.data?.tingkat as T;
        } else if (endpoint === 'subjects') {
          newItem = res.data?.mata_pelajaran as T;
        } else if (endpoint === 'topics') {
          newItem = res.data?.materi as T;
        } else if (endpoint === 'auth/users') {
          newItem = res.data?.user as T;
        } else {
          newItem = (res.data?.data || res.data) as T;
        }

        setData((prev) => [...prev, newItem]);
        handleSuccess('Data berhasil dibuat', newItem);
        return newItem;
      } catch (err: any) {
        handleError(err, 'Gagal membuat data');
        throw err;
      }
    },
    [endpoint, handleError, handleSuccess]
  );

  const update = useCallback(
    async (id: number, payload: Partial<Omit<T, 'id'>>) => {
      try {
        const res = await apiClient.put<any>(
          `${endpoint}/${id}`,
          payload
        );

        // Handle different response formats based on endpoint
        let updatedItem: T;
        if (endpoint === 'levels') {
          updatedItem = res.data?.tingkat as T;
        } else if (endpoint === 'subjects') {
          updatedItem = res.data?.mata_pelajaran as T;
        } else if (endpoint === 'topics') {
          updatedItem = res.data?.materi as T;
        } else if (endpoint === 'auth/users') {
          updatedItem = res.data?.user as T;
        } else {
          updatedItem = (res.data?.data || res.data) as T;
        }

        setData((prev) =>
          prev.map((item) => (item.id === id ? updatedItem : item))
        );
        handleSuccess('Data berhasil diperbarui', updatedItem);
        return updatedItem;
      } catch (err: any) {
        handleError(err, 'Gagal memperbarui data');
        throw err;
      }
    },
    [endpoint, handleError, handleSuccess]
  );

  const remove = useCallback(
    async (id: number) => {
      try {
        await apiClient.delete(`${endpoint}/${id}`);
        setData((prev) => prev.filter((item) => item.id !== id));
        handleSuccess('Data berhasil dihapus');
      } catch (err: any) {
        handleError(err, 'Gagal menghapus data');
        throw err;
      }
    },
    [endpoint, handleError, handleSuccess]
  );

  useEffect(() => {
    if (options.autoFetch !== false) {
      fetch();
    }
  }, [endpoint, options.autoFetch]);

  // Memoize the data array to prevent unnecessary re-renders
  const memoizedData = useMemo(() => data, [data]);

  return { data: memoizedData, loading, error, fetch, create, update, remove };
}
