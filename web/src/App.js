import logo from './logo.svg';
import './App.css';
import AgentsTable from './components/AgentsTable';
import AgentsTableTwo from './components/AgentsTableTwo';
import 'antd/dist/antd.css';

import { SocketProvider } from './hooks/useWebSocket';
import Main from './components/Main';
function App() {


  return (
    <div>
      <SocketProvider >
        <Main />
      </SocketProvider>
    </div>
  );
}

export default App;
