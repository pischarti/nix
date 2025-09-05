import sqlite3

# Connect to a database file (or create it if it doesn't exist)
conn = sqlite3.connect('employees.db')

# Create a cursor object
cursor = conn.cursor()

# Execute a SQL command to create a table
cursor.execute('''
    CREATE TABLE IF NOT EXISTS employees (
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        position TEXT NOT NULL
    )
''')

# Insert a new record into the table
cursor.execute('''
    INSERT INTO employees (
        name, 
        position
    ) VALUES (?, ?)
''', ('Jane Doe', 'Data Analyst'))

# Commit the changes to the database
conn.commit()

# Execute a SELECT query to fetch data
cursor.execute("SELECT * FROM employees")

# Fetch all the results
results = cursor.fetchall()

print("All employees:")

for row in results:
    print(row)

def safe_update(employee_id, new_position):
    conn = None # Initialize conn outside the try block
    try:
        conn = sqlite3.connect('employees.db')
        cursor = conn.cursor()
        cursor.execute("UPDATE employees SET position = ? WHERE id = ?", (new_position, employee_id))
        
        if cursor.rowcount == 0:
            print(f"Warning: No employee found with ID {employee_id}. No updates made.")
        else:
            conn.commit()
            print(f"Employee ID {employee_id} updated to '{new_position}'.")

    except sqlite3.Error as e:
        print(f"An error occurred: {e}")
        if conn:
            conn.rollback() # Roll back any changes on error

    finally:
        if conn:
            conn.close()

# Example usage
safe_update(1, 'Senior Data Analyst')
safe_update(99, 'Non-existent Job') # This will trigger the 'if' block and no-update message

with sqlite3.connect('employees.db') as conn:
    cursor = conn.cursor()
    cursor.execute("SELECT id, name, position FROM employees")

    print("Employee directory:")
    for employee_id, name, position in cursor:
        print(f"ID: {employee_id}, Name: {name}, Position: {position}")


# Close the connection
conn.close()
