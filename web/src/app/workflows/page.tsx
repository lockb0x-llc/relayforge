'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import Link from 'next/link';
import { 
  PlusIcon, 
  PlayIcon, 
  PencilIcon,
  TrashIcon,
  DocumentTextIcon 
} from '@heroicons/react/24/outline';

interface Workflow {
  id: number;
  name: string;
  description: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export default function WorkflowsPage() {
  const { user, token } = useAuth();
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (user && token) {
      fetchWorkflows();
    }
  }, [user, token]);

  const fetchWorkflows = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/workflows', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });
      
      if (response.ok) {
        const data = await response.json();
        setWorkflows(data.workflows || []);
      }
    } catch (error) {
      console.error('Failed to fetch workflows:', error);
    } finally {
      setLoading(false);
    }
  };

  if (!user) {
    return (
      <div className="text-center py-16">
        <h2 className="text-2xl font-bold text-gray-900 mb-4">
          Please login to manage workflows
        </h2>
        <p className="text-gray-600">
          You need to be authenticated to create and manage workflows.
        </p>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="text-center py-16">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
        <p className="text-gray-600 mt-4">Loading workflows...</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Workflows</h1>
          <p className="text-gray-600 mt-2">
            Manage your infrastructure automation workflows
          </p>
        </div>
        <Link
          href="/workflows/new"
          className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors flex items-center space-x-2"
        >
          <PlusIcon className="w-5 h-5" />
          <span>New Workflow</span>
        </Link>
      </div>

      {/* Workflows List */}
      {workflows.length === 0 ? (
        <div className="text-center py-16 bg-white rounded-lg shadow-md">
          <DocumentTextIcon className="w-16 h-16 text-gray-400 mx-auto mb-4" />
          <h3 className="text-xl font-semibold text-gray-900 mb-2">
            No workflows yet
          </h3>
          <p className="text-gray-600 mb-6">
            Get started by creating your first workflow
          </p>
          <Link
            href="/workflows/new"
            className="bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700 transition-colors"
          >
            Create Workflow
          </Link>
        </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {workflows.map((workflow) => (
            <div key={workflow.id} className="bg-white rounded-lg shadow-md p-6">
              <div className="flex justify-between items-start mb-4">
                <div>
                  <h3 className="text-lg font-semibold text-gray-900">
                    {workflow.name}
                  </h3>
                  <p className="text-gray-600 text-sm mt-1">
                    {workflow.description || 'No description'}
                  </p>
                </div>
                <span
                  className={`px-2 py-1 text-xs rounded-full ${
                    workflow.is_active
                      ? 'bg-green-100 text-green-800'
                      : 'bg-gray-100 text-gray-800'
                  }`}
                >
                  {workflow.is_active ? 'Active' : 'Inactive'}
                </span>
              </div>

              <div className="text-sm text-gray-500 mb-4">
                Created: {new Date(workflow.created_at).toLocaleDateString()}
              </div>

              <div className="flex space-x-2">
                <Link
                  href={`/workflows/${workflow.id}/runs`}
                  className="flex-1 bg-blue-600 text-white text-center py-2 rounded hover:bg-blue-700 transition-colors flex items-center justify-center space-x-1"
                >
                  <PlayIcon className="w-4 h-4" />
                  <span>Run</span>
                </Link>
                <Link
                  href={`/workflows/${workflow.id}/edit`}
                  className="px-3 py-2 border border-gray-300 rounded hover:bg-gray-50 transition-colors"
                >
                  <PencilIcon className="w-4 h-4" />
                </Link>
                <button className="px-3 py-2 border border-red-300 text-red-600 rounded hover:bg-red-50 transition-colors">
                  <TrashIcon className="w-4 h-4" />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}