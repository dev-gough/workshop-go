import { Link } from "react-router-dom";
import Header from "../../components/Header";

const Home = () => {
    return (
        <body className="bg-gray-400">
        <link rel="icon" href="/static/favicon.ico" type="image/x-icon" />
        <div className="mx-auto">
            <Header />
            <div className="min-h-screen flex flex-col">
                <div className="flex-1 bg-blue-200" style={{ height: "50vh" }}>
                    <div className="text-xl font-bold text-white p-5">Dashboard Area</div>
                    <div className="p-5 text-white">
                        Content that you can customize for tracking various metrics or statuses.
                    </div>
                </div>
                <div className="flex w-full" style={{ height: "50vh" }}>
                    <div className="w-1/2 bg-green-200 p-5">
                        <Link to="/projects" className="text-xl font-bold">Projects</Link>
                        <ul className="list-disc list-inside">
                            <li className="my-3">
                                <a className="p-1 bg-green-300 hover:bg-green-400 rounded-md" href="/projects/flashcard">Flashcards</a>
                            </li>
                            <li className="my-3">
                                <a className="p-1 bg-green-300 hover:bg-green-400 rounded-md" href="/projects/flashcard/random">Random Flashcard</a>
                            </li>
                            <li className="my-3">
                                <a className="p-1 bg-green-300 hover:bg-green-400 rounded-md" href="/projects/gol">Game of Life</a>
                            </li>
                            <li className="my-3">
                                <a className="p-1 bg-green-300 hover:bg-green-400 rounded-md" href="/projects/polarclock">Polar Clock</a>
                            </li>
                        </ul>
                    </div>
                    <div className="w-1/2 bg-red-200 p-5">
                        <div className="text-xl font-bold">Transactions</div>
                        <ul className="list-disc list-inside">
                            <li>Transaction 1</li>
                            <li>Transaction 2</li>
                            <li>Transaction 3</li>
                        </ul>
                    </div>
                </div>
            </div>
        </div>
    </body>
    )
}


export default Home;