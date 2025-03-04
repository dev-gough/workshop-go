import { FC } from 'react';

const Navbar: FC = () => {
    return (
        <nav className="space-x-4">
            <a href="/projects" className="px-4 hover:bg-gray-600 bg-gray-700 py-2 rounded-md">Projects</a>
            <a href="/learn" className="px-4 hover:bg-gray-600 bg-gray-700 py-2 rounded-md">Learn</a>
            <a href="/data" className="px-4 hover:bg-gray-600 bg-gray-700 py-2 rounded-md">Data</a>
            <a href="/about" className="px-4 hover:bg-gray-600 bg-gray-700 py-2 rounded-md">About</a>
        </nav>
    );
};

export default Navbar;
