'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { fetchSoal } from '@/lib/api';
import { Soal } from '@/lib/types';

export default function ViewSoalPage() {
  const params = useParams();
  const router = useRouter();
  const id = parseInt(params.id as string);
  const [soal, setSoal] = useState<Soal | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (id) loadSoal();
  }, [id]);

  const loadSoal = async () => {
    try {
      const data = await fetchSoal(id);
      setSoal(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load soal');
    } finally {
      setLoading(false);
    }
  };

  if (loading) return <div className="p-8">Loading...</div>;
  if (error) return <div className="p-8 text-red-500">{error}</div>;
  if (!soal) return <div className="p-8">Question not found</div>;

  return (
    <div className="p-8 max-w-2xl mx-auto">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-2xl font-bold">View Question</h1>
        <div className="space-x-4">
          <Link href={`/questions/${id}/edit`} className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600">
            Edit
          </Link>
          <button onClick={() => router.back()} className="bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600">
            Back
          </button>
        </div>
      </div>

      <div className="bg-white border border-gray-300 rounded p-6 space-y-4">
        <div>
          <strong>Subject:</strong> {soal.materi.mataPelajaran.nama}
        </div>
        <div>
          <strong>Material:</strong> {soal.materi.nama} (Level {soal.materi.tingkatan})
        </div>
        <div>
          <strong>Question:</strong>
          <p className="mt-2 p-3 bg-gray-50 rounded">{soal.pertanyaan}</p>
        </div>
        <div>
          <strong>Options:</strong>
          <ul className="mt-2 space-y-1">
            <li className={soal.jawabanBenar === 1 ? 'font-bold text-green-600' : ''}>
              A. {soal.opsiA}
            </li>
            <li className={soal.jawabanBenar === 2 ? 'font-bold text-green-600' : ''}>
              B. {soal.opsiB}
            </li>
            <li className={soal.jawabanBenar === 3 ? 'font-bold text-green-600' : ''}>
              C. {soal.opsiC}
            </li>
            <li className={soal.jawabanBenar === 4 ? 'font-bold text-green-600' : ''}>
              D. {soal.opsiD}
            </li>
          </ul>
        </div>
        <div>
          <strong>Correct Answer:</strong> {['', 'A', 'B', 'C', 'D'][soal.jawabanBenar]}
        </div>
      </div>
    </div>
  );
}