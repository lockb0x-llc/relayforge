'use client';

import { Fragment } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import Link from 'next/link';
import { Menu, Transition } from '@headlessui/react';
import { 
  Bars3Icon, 
  HomeIcon, 
  CogIcon, 
  UserIcon,
  ArrowRightOnRectangleIcon,
  ArrowLeftOnRectangleIcon
} from '@heroicons/react/24/outline';

export default function Navigation() {
  const { user, logout } = useAuth();

  const handleLogin = () => {
    window.location.href = 'http://localhost:8080/api/auth/github';
  };

  return (
    <nav className="bg-white shadow-lg">
      <div className="container mx-auto px-4">
        <div className="flex justify-between items-center py-4">
          {/* Logo */}
          <Link href="/" className="flex items-center space-x-2">
            <CogIcon className="w-8 h-8 text-blue-600" />
            <span className="text-xl font-bold text-gray-900">RelayForge</span>
          </Link>

          {/* Navigation Links */}
          <div className="hidden md:flex items-center space-x-6">
            <Link href="/" className="flex items-center space-x-1 text-gray-600 hover:text-blue-600">
              <HomeIcon className="w-5 h-5" />
              <span>Home</span>
            </Link>
            
            {user && (
              <>
                <Link href="/workflows" className="text-gray-600 hover:text-blue-600">
                  Workflows
                </Link>
                <Link href="/runners" className="text-gray-600 hover:text-blue-600">
                  Runners
                </Link>
              </>
            )}
          </div>

          {/* User Menu */}
          <div className="flex items-center space-x-4">
            {user ? (
              <Menu as="div" className="relative">
                <Menu.Button className="flex items-center space-x-2 text-gray-600 hover:text-blue-600">
                  <img 
                    src={user.avatar_url} 
                    alt={user.username}
                    className="w-8 h-8 rounded-full"
                  />
                  <span>{user.username}</span>
                </Menu.Button>
                
                <Transition
                  as={Fragment}
                  enter="transition ease-out duration-100"
                  enterFrom="transform opacity-0 scale-95"
                  enterTo="transform opacity-100 scale-100"
                  leave="transition ease-in duration-75"
                  leaveFrom="transform opacity-100 scale-100"
                  leaveTo="transform opacity-0 scale-95"
                >
                  <Menu.Items className="absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none">
                    <div className="py-1">
                      <Menu.Item>
                        {({ active }) => (
                          <Link
                            href="/profile"
                            className={`${
                              active ? 'bg-gray-100' : ''
                            } flex items-center px-4 py-2 text-sm text-gray-700`}
                          >
                            <UserIcon className="w-4 h-4 mr-2" />
                            Profile
                          </Link>
                        )}
                      </Menu.Item>
                      <Menu.Item>
                        {({ active }) => (
                          <button
                            onClick={logout}
                            className={`${
                              active ? 'bg-gray-100' : ''
                            } flex items-center w-full px-4 py-2 text-sm text-gray-700`}
                          >
                            <ArrowRightOnRectangleIcon className="w-4 h-4 mr-2" />
                            Logout
                          </button>
                        )}
                      </Menu.Item>
                    </div>
                  </Menu.Items>
                </Transition>
              </Menu>
            ) : (
              <button
                onClick={handleLogin}
                className="flex items-center space-x-2 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
              >
                <ArrowLeftOnRectangleIcon className="w-5 h-5" />
                <span>Login with GitHub</span>
              </button>
            )}
          </div>
        </div>
      </div>
    </nav>
  );
}