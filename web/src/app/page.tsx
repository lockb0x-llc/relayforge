import { PlayIcon, CogIcon, ChartBarIcon } from '@heroicons/react/24/outline';
import Link from 'next/link';

export default function Home() {
  return (
    <div className="space-y-12">
      {/* Hero Section */}
      <div className="text-center py-16">
        <h1 className="text-5xl font-bold text-gray-900 mb-6">
          Welcome to RelayForge
        </h1>
        <p className="text-xl text-gray-600 mb-8 max-w-3xl mx-auto">
          A decentralized infrastructure orchestration platform with cloud control plane, 
          federated runners, CLI, UI, and workflow templates.
        </p>
        <div className="flex justify-center space-x-4">
          <Link 
            href="/workflows" 
            className="bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700 transition-colors"
          >
            Get Started
          </Link>
          <Link 
            href="/docs" 
            className="border border-gray-300 text-gray-700 px-6 py-3 rounded-lg hover:bg-gray-50 transition-colors"
          >
            Documentation
          </Link>
        </div>
      </div>

      {/* Features */}
      <div className="grid md:grid-cols-3 gap-8">
        <div className="text-center p-6">
          <div className="w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <PlayIcon className="w-8 h-8 text-blue-600" />
          </div>
          <h3 className="text-xl font-semibold mb-3">Workflow Automation</h3>
          <p className="text-gray-600">
            Define and execute complex infrastructure workflows using YAML configurations
          </p>
        </div>
        
        <div className="text-center p-6">
          <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <CogIcon className="w-8 h-8 text-green-600" />
          </div>
          <h3 className="text-xl font-semibold mb-3">Federated Runners</h3>
          <p className="text-gray-600">
            Deploy runners across multiple environments for distributed execution
          </p>
        </div>
        
        <div className="text-center p-6">
          <div className="w-16 h-16 bg-purple-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <ChartBarIcon className="w-8 h-8 text-purple-600" />
          </div>
          <h3 className="text-xl font-semibold mb-3">Real-time Monitoring</h3>
          <p className="text-gray-600">
            Monitor workflow execution with real-time logs and status updates
          </p>
        </div>
      </div>

      {/* Quick Start */}
      <div className="bg-white rounded-lg shadow-md p-8">
        <h2 className="text-2xl font-bold mb-6">Quick Start</h2>
        <div className="space-y-4">
          <div className="flex items-start space-x-3">
            <div className="w-6 h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-sm font-semibold">
              1
            </div>
            <div>
              <h4 className="font-semibold">Create a Workflow</h4>
              <p className="text-gray-600">Define your infrastructure automation using YAML</p>
            </div>
          </div>
          <div className="flex items-start space-x-3">
            <div className="w-6 h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-sm font-semibold">
              2
            </div>
            <div>
              <h4 className="font-semibold">Deploy Runners</h4>
              <p className="text-gray-600">Set up federated runners in your target environments</p>
            </div>
          </div>
          <div className="flex items-start space-x-3">
            <div className="w-6 h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-sm font-semibold">
              3
            </div>
            <div>
              <h4 className="font-semibold">Execute & Monitor</h4>
              <p className="text-gray-600">Run workflows and monitor progress in real-time</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
