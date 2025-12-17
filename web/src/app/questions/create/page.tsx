'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { createSoal, fetchMateriList } from '@/lib/api';
import { CreateSoalRequest, JawabanOption, Materi } from '@/lib/types';

export default function CreateSoalPage() {
  const router = useRouter();
  const [materiList, setMateriList] = useState<Materi[]>([]);
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState<CreateSoalRequest>({
    id_materi: 0,
    pertanyaan: '',
    opsiA: '',
    opsiB: '',
    opsiC: '',
    opsiD: '',
    jawaban_benar: JawabanOption.A,
  });

  useEffect(() => {
    loadMateri();
  }, []);

  const loadMateri = async () => {
    try {
      const data = await fetchMateriList();
      setMateriList(data.materi);
    } catch (err) {
      alert('Failed to load materi');
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.id_materi || !formData.pertanyaan.trim()) {
      alert('Please fill in all required fields');
      return;
    }
    setLoading(true);
    try {
      await createSoal(formData);
      router.push('/');
    } catch (err) {
      alert('Failed to create soal');
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: name === 'id_materi' || name === 'jawaban_benar' ? parseInt(value) : value,
    }));
  };

  return (
    <div className="p-8 max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-8">Create New Question</h1>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div>
          <label className="block text-sm font-medium mb-2">Material</label>
          <select
            name="id_materi"
            value={formData.id_materi}
            onChange={handleChange}
            className="w-full p-2 border border-gray-300 rounded"
            required
          >
            <option value={0}>Select Material</option>
            {materiList.map(materi => (
              <option key={materi.id} value={materi.id}>
                {materi.mataPelajaran.nama} - {materi.nama} (Level {materi.tingkatan})
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium mb-2">Question</label>
          <textarea
            name="pertanyaan"
            value={formData.pertanyaan}
            onChange={handleChange}
            className="w-full p-2 border border-gray-300 rounded h-24"
            placeholder="Enter the question"
            required
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium mb-2">Option A</label>
            <input
              type="text"
              name="opsiA"
              value={formData.opsiA}
              onChange={handleChange}
              className="w-full p-2 border border-gray-300 rounded"
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-2">Option B</label>
            <input
              type="text"
              name="opsiB"
              value={formData.opsiB}
              onChange={handleChange}
              className="w-full p-2 border border-gray-300 rounded"
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-2">Option C</label>
            <input
              type="text"
              name="opsiC"
              value={formData.opsiC}
              onChange={handleChange}
              className="w-full p-2 border border-gray-300 rounded"
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-2">Option D</label>
            <input
              type="text"
              name="opsiD"
              value={formData.opsiD}
              onChange={handleChange}
              className="w-full p-2 border border-gray-300 rounded"
              required
            />
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium mb-2">Correct Answer</label>
          <select
            name="jawaban_benar"
            value={formData.jawaban_benar}
            onChange={handleChange}
            className="w-full p-2 border border-gray-300 rounded"
            required
          >
            <option value={JawabanOption.A}>A</option>
            <option value={JawabanOption.B}>B</option>
            <option value={JawabanOption.C}>C</option>
            <option value={JawabanOption.D}>D</option>
          </select>
        </div>

        <div className="flex gap-4">
          <button
            type="submit"
            disabled={loading}
            className="bg-blue-500 text-white px-6 py-2 rounded hover:bg-blue-600 disabled:opacity-50"
          >
            {loading ? 'Creating...' : 'Create Question'}
          </button>
          <button
            type="button"
            onClick={() => router.back()}
            className="bg-gray-500 text-white px-6 py-2 rounded hover:bg-gray-600"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}