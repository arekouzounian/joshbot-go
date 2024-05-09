import { useState, useEffect } from 'react';
import Box from '@mui/material/Box'; 

function Leaderboard(props) {
    const [users, setUsers] = useState([]); 

    const boxStyle = { 
        borderRadius: 1, 
        border: '1px solid gray', 
        padding: '1%', 
        fontSize: 'auto', 
        color: 'white', 
        minWidth: '15%',
        maxWidth: '100%', 
        height: 'auto',
        margin: 'auto',
        marginTop: '5%',
        backgroundColor: '#36323d', 
        display: 'inline-block', 
    };

    useEffect(() => {
        fetch(props.API_URL+props.endpoint)
            .then((response) => response.json())
            .then((data) => {
                console.log(data)
                setUsers(data);
            });
    }, [props.API_URL, props.endpoint]);

    return (
        <div className='leaderboardBoxWrapper'>
            <Box 
            sx={boxStyle} >
                <h1>{props.title}</h1>
                {users.length > 0 && 
                    users.map((user, index) => (
                        <Box sx ={{ borderRadius: 1, border: '1px solid gray', width: 'auto', height: '20%', margin: '5%', padding: '5%', maxWidth: '100%'}}>
                        {index+1}. {user[props.keyIdx]}: <b>{user[props.valueIdx]}</b>
                        </Box>
                    ))
                }
            </Box>
        </div>
    )
}

export { Leaderboard }