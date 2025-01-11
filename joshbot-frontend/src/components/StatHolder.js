import { useState, useEffect } from 'react';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemText from '@mui/material/ListItemText';
import Divider from '@mui/material/Divider';

const AVG_ENDPOINT = "/api/v2/joshavg";
const TIME_ENDPOINT = "/api/v2/lastjosh";

function secondsToStr(seconds) {
    let days = 0; 
    let hours = 0; 
    let minutes = 0; 

    while (seconds > 59) {
        seconds -= 60; 
        minutes += 1; 

        if (minutes >= 60) {
            hours++; 
            minutes = 0; 
        }

        if (hours >= 24) {
            days++; 
            hours = 0; 
        }
    }

    let ret = ""; 
    if (days > 0) {
        ret += days.toString() + "d "; 
    }
    if (hours > 0) {
        ret += hours.toString() + "h ";
    }
    if (minutes > 0) {
        ret += minutes.toString() + "m ";
    }
    ret += seconds.toString() + "s";

    console.log(ret);

    return ret; 
}

function StatHolder(props) {
    const [avg, setAvg] = useState([]);
    const [seconds, setSeconds] = useState("Loading...");

    
    useEffect(() => {
        const getTime = async() => {
            fetch(props.API_URL+TIME_ENDPOINT)
                .then((response) => response.json())
                .then((data) => {
                    console.log(data);
                    setSeconds(secondsToStr(data));
                }); 
        }
        
        fetch(props.API_URL+AVG_ENDPOINT)
            .then((response) => response.json())
            .then((data) => {
                console.log(data);
                setAvg(parseFloat(data).toFixed(3));
            });

            const timer = setInterval(getTime, 2000);

            return () => clearInterval(timer);
    }, [props.API_URL]);

    let boxSX = {
        color: 'black',
        py: 0,
        borderRadius: 2,
        margin: '1%',
        minWidth: '10%',
        maxWidth: '30%',
        border: '1px solid',
        borderColor: 'white',
        backgroundColor: 'gray'
    }

    let listSX = {
        textAlign: 'center',

    }

    return (
        <div>
            <List className='statWrapper' sx={boxSX}>
                <ListItem sx={listSX}>
                    <ListItemText primary='Josh Avg (30d)'/>
                </ListItem>
                <Divider></Divider>
                <ListItem sx={listSX}>
                    <ListItemText primary={avg} />
                </ListItem>
            </List>

            <List className='statWrapper' sx={boxSX}>
                <ListItem sx={listSX}>
                    <ListItemText primary='Last Josh (🔄 2s)'/>
                </ListItem>
                <Divider></Divider>
                <ListItem sx={listSX}>
                    <ListItemText primary={seconds} />
                </ListItem>
            </List>
        </div>
    )

}

export {StatHolder}