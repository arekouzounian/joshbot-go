import { useState, useEffect } from "react";


function JoshOTW(props) {
    const [avatar, setAvatar] = useState("");
    const [username, setUsername] = useState("");


    useEffect(() => {
        // fetch(API_URL+AVG_ENDPOINT)
        //     .then((response) => response.json())
        //     .then((data) => {
        //         // console.log(data);
        //         setAvg(parseFloat(data).toFixed(3));
        //     });

        //     const timer = setInterval(getTime, 2000);

        //     return () => clearInterval(timer);
    }, []);
}

export { JoshOTW }