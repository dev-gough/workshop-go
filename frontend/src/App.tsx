import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Home from './pages/home'
import Study from './pages/projects/flashcard';
import Projects from './pages/projects';
import PolarClock from './pages/projects/polarclock';

function App() {
  return (
    <BrowserRouter>
        <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/home" element={<Home />} />
            <Route path="/projects" element={<Projects />} />
            <Route path="/projects/polarclock" element={<PolarClock />} />
        </Routes>
    </BrowserRouter>
  );
}

export default App;
