import sqlite3
import matplotlib.pyplot as plt

# Connect to a SQLite database (or create a new one in memory)
conn = sqlite3.connect(':memory:')
cursor = conn.cursor()

# Create sample tables
cursor.execute('''
    CREATE TABLE employees (
        id INTEGER PRIMARY KEY,
        name TEXT,
        department_id INTEGER
    )
''')
cursor.execute('''
    CREATE TABLE departments (
        id INTEGER PRIMARY KEY,
        name TEXT
    )
''')
cursor.execute('''
    CREATE TABLE sales (
        employee_id INTEGER,
        amount REAL
    )
''')

# Insert sample data
cursor.execute("INSERT INTO employees VALUES (1, 'Alice', 1)")
cursor.execute("INSERT INTO employees VALUES (2, 'Bob', 2)")
cursor.execute("INSERT INTO employees VALUES (3, 'Charlie', 1)")
cursor.execute("INSERT INTO departments VALUES (1, 'Marketing')")
cursor.execute("INSERT INTO departments VALUES (2, 'Sales')")
cursor.execute("INSERT INTO sales VALUES (1, 150.00)")
cursor.execute("INSERT INTO sales VALUES (2, 200.00)")
cursor.execute("INSERT INTO sales VALUES (3, 100.00)")
conn.commit()

# Advanced SQL query: Join tables, group by, sum, and order
query = """
SELECT
    d.name AS Department,
    SUM(s.amount) AS TotalSales
FROM
    employees e
JOIN
    departments d ON e.department_id = d.id
JOIN
    sales s ON e.id = s.employee_id
GROUP BY
    d.name
ORDER BY
    TotalSales DESC;
"""

# Execute query and fetch results without pandas
cursor.execute(query)
rows = cursor.fetchall()

# Display results
print("Department Sales:")
for department, total_sales in rows:
    print(f"{department}\t{total_sales}")

# Visualize results (optional)
departments = [r[0] for r in rows]
totals = [r[1] for r in rows]

plt.figure(figsize=(8, 6))
plt.bar(departments, totals)
plt.title('Total Sales by Department')
plt.xlabel('Department')
plt.ylabel('Total Sales')
plt.tight_layout()
plt.show()

# Close connection
conn.close()
