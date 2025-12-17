'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { fetchSoalList, deleteSoal } from '@/lib/api';
import { Soal, ListSoalResponse } from '@/lib/types';

export default function Home() {
  const [soalList, setSoalList] = useState<Soal[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadSoal();
  }, []);

  const loadSoal = async () => {
    try {
      const data: ListSoalResponse = await fetchSoalList();
      setSoalList(data.soal);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load soal');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this question?')) return;
    try {
      await deleteSoal(id);
      setSoalList(soalList.filter(s => s.id !== id));
    } catch (err) {
      alert('Failed to delete soal');
    }
  };

  if (loading) return <div className="p-8">Loading...</div>;
  if (error) return <div className="p-8 text-red-500">{error}</div>;

  return (
    <div className="p-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-2xl font-bold">Admin - Manage Questions</h1>
        <Link href="/questions/create" className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600">
          Add New Question
        </Link>
      </div>

      <div className="overflow-x-auto">
        <table className="min-w-full bg-white border border-gray-300">
          <thead>
            <tr className="bg-gray-100">
              <th className="px-4 py-2 border">ID</th>
              <th className="px-4 py-2 border">Question</th>
              <th className="px-4 py-2 border">Subject</th>
              <th className="px-4 py-2 border">Material</th>
              <th className="px-4 py-2 border">Level</th>
              <th className="px-4 py-2 border">Correct Answer</th>
              <th className="px-4 py-2 border">Actions</th>
            </tr>
          </thead>
          <tbody>
            {soalList.map((soal) => (
              <tr key={soal.id} className="hover:bg-gray-50">
                <td className="px-4 py-2 border">{soal.id}</td>
                <td className="px-4 py-2 border max-w-xs truncate">{soal.pertanyaan}</td>
                <td className="px-4 py-2 border">{soal.materi.mataPelajaran.nama}</td>
                <td className="px-4 py-2 border">{soal.materi.nama}</td>
                <td className="px-4 py-2 border">{soal.materi.tingkatan}</td>
                <td className="px-4 py-2 border">{['', 'A', 'B', 'C', 'D'][soal.jawabanBenar]}</td>
                <td className="px-4 py-2 border">
                  <Link href={`/questions/${soal.id}`} className="text-blue-500 hover:underline mr-2">View</Link>
                  <Link href={`/questions/${soal.id}/edit`} className="text-green-500 hover:underline mr-2">Edit</Link>
                  <button onClick={() => handleDelete(soal.id)} className="text-red-500 hover:underline">Delete</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {soalList.length === 0 && (
        <div className="text-center py-8 text-gray-500">No questions found. <Link href="/questions/create" className="text-blue-500">Create one</Link></div>
      )}
    </div>
  );
}
