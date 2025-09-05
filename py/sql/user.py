import sqlite3

conn = sqlite3.connect('users.db')
cursor = conn.cursor()

cursor.execute('''
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        email TEXT NOT NULL
    )
''')

def add_user(name, email):
    with sqlite3.connect('users.db') as conn:
        cursor = conn.cursor()

        # Check if the user already exists
        cursor.execute("SELECT name FROM users WHERE email = ?", (email,))
        existing_user = cursor.fetchone()

        if existing_user:
            print(f"User with email {email} already exists: {existing_user[0]}")
        else:
            # If not, insert the new user
            cursor.execute("INSERT INTO users (name, email) VALUES (?, ?)", (name, email))
            conn.commit()
            print(f"User {name} added successfully.")

# Example usage
add_user('Alice', 'alice@example.com')
add_user('Alice', 'alice@example.com') # This will trigger the 'if' block
