import sqlite3

# Connect to a database file (or create it if it doesn't exist)
conn = sqlite3.connect('inventory.db')

# Create a cursor object
cursor = conn.cursor()

# Execute a SQL command to create a table
cursor.execute('''
    CREATE TABLE IF NOT EXISTS products (
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        price REAL NOT NULL
    )
''')

new_products = [
    ('Laptop', 1200.00),
    ('Mouse', 25.50),
    ('Keyboard', 75.00),
]

with sqlite3.connect('inventory.db') as conn:
    cursor = conn.cursor()

    # Loop through the list and insert each product
    for name, price in new_products:
        cursor.execute("INSERT INTO products (name, price) VALUES (?, ?)", (name, price))

    conn.commit()
    print(f"{len(new_products)} products inserted.")
