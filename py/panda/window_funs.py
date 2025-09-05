import sqlite3
import pandas as pd

conn = sqlite3.connect(':memory:')
cursor = conn.cursor()

cursor.execute('''
    CREATE TABLE student_grades (
        name TEXT,
        gpa REAL
    )
''')
cursor.execute("INSERT INTO student_grades VALUES ('Alice', 3.8)")
cursor.execute("INSERT INTO student_grades VALUES ('Bob', 3.5)")
cursor.execute("INSERT INTO student_grades VALUES ('Charlie', 3.8)")
cursor.execute("INSERT INTO student_grades VALUES ('David', 3.2)")
conn.commit()

query_rank = """
SELECT
    name,
    gpa,
    RANK() OVER (ORDER BY gpa DESC) AS Rank,
    DENSE_RANK() OVER (ORDER BY gpa DESC) AS DenseRank,
    ROW_NUMBER() OVER (ORDER BY gpa DESC) AS RowNumber
FROM
    student_grades
ORDER BY
    gpa DESC;
"""

df_rank = pd.read_sql_query(query_rank, conn)
print("\nStudent Grades with Ranks:")
print(df_rank)

conn.close()
