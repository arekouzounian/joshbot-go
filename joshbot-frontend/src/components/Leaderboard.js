import { useState, useEffect } from 'react';
import Box from '@mui/material/Box'; 

const API_URL = "http://localhost:5000";

function Leaderboard(props) {
    const [users, setUsers] = useState([]); 

    const boxStyle = { 
        borderRadius: 1, 
        border: '1px solid gray', 
        padding: '1%', 
        fontSize: '80%', 
        color: 'white', 
        width: '15%', 
        height: 'auto',
        margin: 'auto',
        marginTop: '5%',
        backgroundColor: '#36323d', 
        display: 'inline-block', 
    };

    useEffect(() => {
        fetch(API_URL+props.endpoint)
            .then((response) => response.json())
            .then((data) => {
                console.log(data)
                setUsers(data);
            });
    }, [props.endpoint]);

    return (
        <Box 
        sx={boxStyle}>
            <h1>Leaderboard</h1>
            {users.length > 0 && 
                users.map((user, index) => (
                    <Box sx ={{ borderRadius: 1, border: '1px solid gray', width: 'auto', height: '20%', margin: '5%', padding: '5%'}}>
                        {user[props.keyIdx]}:{user[props.valueIdx]}
                    </Box>
                ))
            }
        </Box>
    )
}

export {Leaderboard }