# Database
Honestly since the scope of this project is so small, I'm just going to use sqlite3. 

This means I won't have to do any configuration with PostgreSQL/MySQL/etc, nor will I have to set up any accounts. This will put more pressure on the API to sanitize inputs and all, but even if some form of SQL Injection goes through, there really is no sensitive data being stored in these tables. 

