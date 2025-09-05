import sqlite3
import pandas as pd

conn = sqlite3.connect(':memory:')

# Create a sample employee table
conn.execute("""
CREATE TABLE IF NOT EXISTS Employee (
    Id INTEGER PRIMARY KEY,
    Name TEXT,
    Salary INTEGER,
    ManagerId INTEGER
);
""")

# Insert sample data
conn.execute("INSERT OR IGNORE INTO Employee VALUES (1, 'Joe', 70000, 3);")
conn.execute("INSERT OR IGNORE INTO Employee VALUES (2, 'Henry', 80000, 4);")
conn.execute("INSERT OR IGNORE INTO Employee VALUES (3, 'Sam', 60000, NULL);")
conn.execute("INSERT OR IGNORE INTO Employee VALUES (4, 'Max', 90000, NULL);")
conn.commit()

# SQL query using a self-join to find employees with higher salaries than their managers
query = """
SELECT 
    a.Name as EmployeeName
FROM 
    Employee AS a
JOIN 
    Employee AS b ON a.ManagerId = b.Id
WHERE 
    a.Salary > b.Salary;
"""

# Load the query results into a DataFrame and print
df = pd.read_sql_query(query, conn)
conn.close()

print(df)
