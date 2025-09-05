import sqlite3

# Connect to the SQLite database (or create it if it doesn't exist)
conn = sqlite3.connect('example.db')

# Create a cursor object to execute SQL commands
cursor = conn.cursor()

# Create a table (if it doesn't exist)
cursor.execute('''
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        age INTEGER
    )
''')

# Insert data into the table
cursor.execute("INSERT INTO users (name, age) VALUES (?, ?)", ('Alice', 30))
cursor.execute("INSERT INTO users (name, age) VALUES (?, ?)", ('Bob', 24))

# Commit the changes
conn.commit()

# Query data from the table
cursor.execute("SELECT * FROM users WHERE age > ?", (25,))
results = cursor.fetchall()

# Print the results
print("Users older than 25:")
for row in results:
    print(f"ID: {row[0]}, Name: {row[1]}, Age: {row[2]}")

# Close the connection
conn.close()
