curl http://localhost:5000/api/v1/joshjoin --data '{"userID":"'"$RANDOM"'", "totalUsers": []}' -H "Content-Type: application/json"