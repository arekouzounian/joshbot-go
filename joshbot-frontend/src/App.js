import './App.css';
import { JoshOTW } from './components/JoshOTW';
import { Leaderboard } from './components/Leaderboard';
import { StatHolder } from './components/StatHolder';

function App() {
  return (
    <div className="App">
      <StatHolder></StatHolder>

      <div className="leaderboardsWrapper">
        <Leaderboard title={'Leaderboard'} endpoint={'/api/v1/joshboard'} keyIdx={1} valueIdx={3}></Leaderboard>

        <div className="centerColumn">
          <JoshOTW></JoshOTW>
          <p>asdf</p>
          <p>asdf</p>
        </div>

        <Leaderboard title={'Wall of Shame'} endpoint={'/api/v1/joshofshame'} keyIdx={1} valueIdx={4}></Leaderboard>
      </div>
    </div>
  );
}

export default App;
