import { createRoot } from 'react-dom/client';
import { AppShell } from './ui/AppShell';
import './styles.css';

const root = document.getElementById('root');

if (!root) {
  throw new Error('Root element not found');
}

createRoot(root).render(<AppShell />);
