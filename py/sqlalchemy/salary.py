from sqlalchemy import create_engine, text, select
from sqlalchemy.orm import sessionmaker

# Use a memory-based SQLite database for demonstration
engine = create_engine('sqlite:///:memory:')

# In a real application, you would define your ORM models here
# For this example, we'll use a raw query with CTEs for clarity

# Prepare a session
Session = sessionmaker(bind=engine)
session = Session()


# Create necessary tables and seed data
from sqlalchemy import text as _text  # alias to avoid shadowing later imports

with engine.begin() as conn:
    # DDL
    conn.execute(_text(
        """
        CREATE TABLE IF NOT EXISTS people (
            id INTEGER PRIMARY KEY,
            name TEXT NOT NULL,
            salary REAL NOT NULL
        );
        """
    ))
    conn.execute(_text(
        """
        CREATE TABLE IF NOT EXISTS salaries (
            id INTEGER PRIMARY KEY,
            salary REAL NOT NULL,
            gender TEXT NOT NULL
        );
        """
    ))

    # Seed data
    people_rows = [
        {"name": "Alice", "salary": 52000.0},
        {"name": "Bob", "salary": 70000.0},
        {"name": "Carol", "salary": 61000.0},
        {"name": "Dave", "salary": 45000.0},
        {"name": "Eve", "salary": 58000.0},
        {"name": "Frank", "salary": 80000.0},
    ]
    salary_rows = [
        # Female salaries (average ~57k)
        {"salary": 52000.0, "gender": "Female"},
        {"salary": 58000.0, "gender": "Female"},
        {"salary": 61000.0, "gender": "Female"},
        # Male salaries
        {"salary": 45000.0, "gender": "Male"},
        {"salary": 70000.0, "gender": "Male"},
        {"salary": 80000.0, "gender": "Male"},
    ]

    conn.execute(
        _text("INSERT INTO people (name, salary) VALUES (:name, :salary)"),
        people_rows,
    )
    conn.execute(
        _text("INSERT INTO salaries (salary, gender) VALUES (:salary, :gender)"),
        salary_rows,
    )

# Define the SQL query with a CTE
# This CTE example finds the average salary of female employees first
query_with_cte = """
WITH avg_female_salary AS (
    SELECT AVG(salary) AS avgSalary
    FROM salaries
    WHERE gender = 'Female'
)
SELECT name, salary
FROM people
WHERE salary >= (SELECT avgSalary FROM avg_female_salary);
"""

# Execute the query and fetch results
try:
    with engine.connect() as conn:
        result = conn.execute(text(query_with_cte)).fetchall()
        for row in result:
            print(f"Name: {row.name}, Salary: {row.salary}")
finally:
    session.close()

