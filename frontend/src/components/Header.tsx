import { FC } from "react"

interface HeaderProps {
    className?: string
    projectName?: string
}

const Header: FC<HeaderProps> = ({ className, projectName }) => {
    return (
        <header id='header' className={`bg-gray-800 text-white p-4 flex justify-between items-center ${className}`}>
            <div className="flex items-center space-x-4">
                <a href="/home" className="text-lg font-bold">Workshop</a>
                <button onClick={() => window.history.back()} className="px-4 hover:bg-gray-600 bg-gray-700 py-2 rounded-md">Back</button>
            </div>
            <div>
                <p className="text-xl font-mono">{projectName}</p>
            </div>
            <nav className="space-x-4">
                <a href="/projects" className="px-4 hover:bg-gray-600 bg-gray-700 py-2 rounded-md">Projects</a>
                <a href="/learn" className="px-4 hover:bg-gray-600 bg-gray-700 py-2 rounded-md">Learn</a>
                <a href="/data" className="px-4 hover:bg-gray-600 bg-gray-700 py-2 rounded-md">Data</a>
                <a href="/about" className="px-4 hover:bg-gray-600 bg-gray-700 py-2 rounded-md">About</a>
            </nav>
        </header>
    );
};

export default Header;
