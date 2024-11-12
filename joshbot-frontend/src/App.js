import './App.css';
import { JoshOTW } from './components/JoshOTW';
import { Leaderboard } from './components/Leaderboard';
import { StatHolder } from './components/StatHolder';

const API_URL = "https://joshbot.xyz";

function App() {
  return (
    <div className="App">
      <StatHolder API_URL={API_URL} ></StatHolder>

      <div className="leaderboardsWrapper">
        <Leaderboard API_URL={API_URL} title={'Leaderboard'} endpoint={'/api/v1/joshboard'} keyIdx={1} valueIdx={3}></Leaderboard>

        <div className="centerColumn">
          <JoshOTW API_URL={API_URL}></JoshOTW>
        </div>

        <Leaderboard API_URL={API_URL} title={'Wall of Shame'} endpoint={'/api/v1/joshofshame'} keyIdx={1} valueIdx={4}></Leaderboard>
      </div>
    </div>
  );
}

export default App;
