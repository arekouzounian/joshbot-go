import './App.css';
import { Leaderboard } from './components/Leaderboard';

function App() {
  return (
    <div className="App">
      <div className="leaderboardsWrapper">
        <Leaderboard endpoint={'/api/v1/joshboard'} keyIdx={1} valueIdx={3}></Leaderboard>
        <Leaderboard endpoint={'/api/v1/joshofshame'} keyIdx={1} valueIdx={4}></Leaderboard>
      </div>
    </div>
  );
}

export default App;
