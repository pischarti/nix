import sqlite3

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

# Fetch all employees and compute the result using Python only (no SQL self-join)
cursor = conn.cursor()
cursor.execute("SELECT Id, Name, Salary, ManagerId FROM Employee")
rows = cursor.fetchall()

# Build an in-memory index by Id
employees = {
    row[0]: {"id": row[0], "name": row[1], "salary": row[2], "manager_id": row[3]}
    for row in rows
}

# Find employees with salary higher than their manager's salary
higher_than_manager = []
for employee in employees.values():
    manager_id = employee["manager_id"]
    if manager_id is None:
        continue
    manager = employees.get(manager_id)
    if manager is not None and employee["salary"] > manager["salary"]:
        higher_than_manager.append(employee["name"])

conn.close()

print("EmployeeName")
for name in higher_than_manager:
    print(name)
