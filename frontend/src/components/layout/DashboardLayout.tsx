'use client';

import { Fragment, useState, useEffect } from 'react';
import { Dialog, Transition } from '@headlessui/react';
import { useSession, signOut } from 'next-auth/react';
import Link from 'next/link';
import Image from 'next/image';
import { usePathname } from 'next/navigation';
import {
    Bars3Icon,
    XMarkIcon,
    HomeIcon,
    FolderIcon,
    ShieldCheckIcon,
    ArrowRightOnRectangleIcon,
} from '@heroicons/react/24/outline';

const navigation = [
    { name: 'Dashboard', href: '/dashboard', icon: HomeIcon },
    { name: 'Repositories', href: '/repositories', icon: FolderIcon },
];

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
    const [sidebarOpen, setSidebarOpen] = useState(false);
    const { data: session } = useSession();
    const pathname = usePathname();
    const [scrolled, setScrolled] = useState(false);

    useEffect(() => {
        const handleScroll = () => {
            setScrolled(window.scrollY > 0);
        };

        window.addEventListener('scroll', handleScroll);
        return () => window.removeEventListener('scroll', handleScroll);
    }, []);

    return (
        <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
            <Transition.Root show={sidebarOpen} as={Fragment}>
                <Dialog as="div" className="relative z-50 lg:hidden" onClose={setSidebarOpen}>
                    <Transition.Child
                        as={Fragment}
                        enter="transition-opacity ease-linear duration-300"
                        enterFrom="opacity-0"
                        enterTo="opacity-100"
                        leave="transition-opacity ease-linear duration-300"
                        leaveFrom="opacity-100"
                        leaveTo="opacity-0"
                    >
                        <div className="fixed inset-0 bg-gray-900/80" />
                    </Transition.Child>

                    <div className="fixed inset-0 flex">
                        <Transition.Child
                            as={Fragment}
                            enter="transition ease-in-out duration-300 transform"
                            enterFrom="-translate-x-full"
                            enterTo="translate-x-0"
                            leave="transition ease-in-out duration-300 transform"
                            leaveFrom="translate-x-0"
                            leaveTo="-translate-x-full"
                        >
                            <Dialog.Panel className="relative mr-16 flex w-full max-w-xs flex-1">
                                <Transition.Child
                                    as={Fragment}
                                    enter="ease-in-out duration-300"
                                    enterFrom="opacity-0"
                                    enterTo="opacity-100"
                                    leave="ease-in-out duration-300"
                                    leaveFrom="opacity-100"
                                    leaveTo="opacity-0"
                                >
                                    <div className="absolute left-full top-0 flex w-16 justify-center pt-5">
                                        <button
                                            type="button"
                                            className="-m-2.5 p-2.5 text-white"
                                            onClick={() => setSidebarOpen(false)}
                                        >
                                            <span className="sr-only">Close sidebar</span>
                                            <XMarkIcon className="h-6 w-6" aria-hidden="true" />
                                        </button>
                                    </div>
                                </Transition.Child>

                                <div className="flex grow flex-col gap-y-5 overflow-y-auto bg-white dark:bg-gray-800 px-6 pb-4 ring-1 ring-white/10">
                                    <div className="flex h-16 shrink-0 items-center">
                                        <div className="flex items-center gap-x-3">
                                            <ShieldCheckIcon className="h-8 w-8 text-indigo-600" />
                                            <h1 className="text-xl font-semibold text-indigo-600">
                                                KeyGraph SAST
                                            </h1>
                                        </div>
                                    </div>
                                    <nav className="flex flex-1 flex-col">
                                        <ul role="list" className="flex flex-1 flex-col gap-y-7">
                                            <li>
                                                <ul role="list" className="-mx-2 space-y-1">
                                                    {navigation.map(item => (
                                                        <li key={item.name}>
                                                            <Link
                                                                href={item.href}
                                                                className={`group flex gap-x-3 rounded-md p-2 text-sm font-semibold leading-6 ${
                                                                    pathname === item.href
                                                                        ? 'bg-gray-100 dark:bg-gray-700 text-indigo-600'
                                                                        : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-indigo-600'
                                                                }`}
                                                            >
                                                                <item.icon
                                                                    className={`h-6 w-6 shrink-0 ${
                                                                        pathname === item.href
                                                                            ? 'text-indigo-600'
                                                                            : 'text-gray-500 dark:text-gray-400 group-hover:text-indigo-600'
                                                                    }`}
                                                                    aria-hidden="true"
                                                                />
                                                                {item.name}
                                                            </Link>
                                                        </li>
                                                    ))}
                                                </ul>
                                            </li>
                                        </ul>
                                    </nav>
                                </div>
                            </Dialog.Panel>
                        </Transition.Child>
                    </div>
                </Dialog>
            </Transition.Root>

            {/* Static sidebar for desktop */}
            <div className="hidden lg:fixed lg:inset-y-0 lg:z-50 lg:flex lg:w-72 lg:flex-col">
                <div className="flex grow flex-col gap-y-5 overflow-y-auto bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 px-6">
                    <div className="flex h-16 shrink-0 items-center">
                        <div className="flex items-center gap-x-3">
                            <ShieldCheckIcon className="h-8 w-8 text-indigo-600" />
                            <h1 className="text-xl font-semibold text-indigo-600">KeyGraph SAST</h1>
                        </div>
                    </div>
                    <nav className="flex flex-1 flex-col">
                        <ul role="list" className="flex flex-1 flex-col gap-y-7">
                            <li>
                                <ul role="list" className="-mx-2 space-y-1">
                                    {navigation.map(item => (
                                        <li key={item.name}>
                                            <Link
                                                href={item.href}
                                                className={`group flex gap-x-3 rounded-md p-2 text-sm font-semibold leading-6 ${
                                                    pathname === item.href
                                                        ? 'bg-gray-100 dark:bg-gray-700 text-indigo-600'
                                                        : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-indigo-600'
                                                }`}
                                            >
                                                <item.icon
                                                    className={`h-6 w-6 shrink-0 ${
                                                        pathname === item.href
                                                            ? 'text-indigo-600'
                                                            : 'text-gray-500 dark:text-gray-400 group-hover:text-indigo-600'
                                                    }`}
                                                    aria-hidden="true"
                                                />
                                                {item.name}
                                            </Link>
                                        </li>
                                    ))}
                                </ul>
                            </li>

                            <li className="mt-auto">
                                {session?.user && (
                                    <div className="flex items-center gap-x-4 border-t border-gray-200 dark:border-gray-700 py-3 mt-auto">
                                        {session.user.image && (
                                            <div className="relative h-10 w-10 rounded-full overflow-hidden bg-gray-100">
                                                <Image
                                                    src={session.user.image}
                                                    alt={session.user.name || 'User profile'}
                                                    fill
                                                    sizes="40px"
                                                    className="object-cover"
                                                    priority
                                                />
                                            </div>
                                        )}
                                        <div className="min-w-0 flex-auto">
                                            <p className="text-sm font-semibold text-gray-900 dark:text-gray-100">
                                                {session.user.name}
                                            </p>
                                            <p className="truncate text-xs text-gray-500 dark:text-gray-400">
                                                {session.user.email}
                                            </p>
                                        </div>
                                        <button
                                            onClick={() => signOut({ callbackUrl: '/' })}
                                            className="p-1.5 rounded-md text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700"
                                        >
                                            <ArrowRightOnRectangleIcon
                                                className="h-5 w-5"
                                                aria-hidden="true"
                                            />
                                        </button>
                                    </div>
                                )}
                            </li>
                        </ul>
                    </nav>
                </div>
            </div>

            <div className="lg:pl-72">
                <div
                    className={`sticky top-0 z-40 flex h-16 shrink-0 items-center gap-x-4 border-b border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-sm ${scrolled ? 'shadow' : ''} px-4 sm:gap-x-6 sm:px-6 lg:px-8`}
                >
                    <button
                        type="button"
                        className="-m-2.5 p-2.5 text-gray-700 dark:text-gray-300 lg:hidden"
                        onClick={() => setSidebarOpen(true)}
                    >
                        <span className="sr-only">Open sidebar</span>
                        <Bars3Icon className="h-6 w-6" aria-hidden="true" />
                    </button>

                    {/* Separator */}
                    <div
                        className="h-6 w-px bg-gray-200 dark:bg-gray-700 lg:hidden"
                        aria-hidden="true"
                    />

                    <div className="flex flex-1 gap-x-4 self-stretch lg:gap-x-6">
                        <div className="flex flex-1 items-center justify-between">
                            <div>
                                <h1 className="text-2xl font-semibold text-gray-900 dark:text-white">
                                    {pathname === '/dashboard'
                                        ? 'Dashboard'
                                        : pathname.startsWith('/repositories/')
                                          ? 'Repository Details'
                                          : pathname === '/repositories'
                                            ? 'Repositories'
                                            : ''}
                                </h1>
                            </div>
                        </div>
                    </div>
                </div>

                <main className="py-6">
                    <div className="px-4 sm:px-6 lg:px-8">{children}</div>
                </main>
            </div>
        </div>
    );
}
