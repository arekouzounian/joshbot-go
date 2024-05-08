import { Avatar } from "@mui/material";
import { useState, useEffect } from "react";

const JOSHOTW_ENDPOINT = "/api/v1/joshotw"; 

const USERNAME_IDX = 1; 
const AVATAR_IDX = 2; 

function JoshOTW(props) {
    const [avatar, setAvatar] = useState("");
    const [username, setUsername] = useState("");


    useEffect(() => {
        fetch(props.API_URL+JOSHOTW_ENDPOINT)
            .then((response) => response.json())
            .then((data) => {
                setAvatar(data[0][AVATAR_IDX]);
                setUsername(data[0][USERNAME_IDX]);
                console.log(data);
            });
    }, [props.API_URL]);

    return (
        <div class='joshOtwWrapper'>
            <p>Josh Of The Week</p>
            <Avatar src={avatar} alt="Avatar for this week's josh" sx={{width: '100%', height: '100%' }}></Avatar>
            <p>{username}</p>
        </div>
    )
}

export { JoshOTW }