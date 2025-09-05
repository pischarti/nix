import sqlite3

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

# Compute ranking metrics in Python (no SQL window functions, no pandas)
cursor.execute("SELECT name, gpa FROM student_grades ORDER BY rowid")
rows = cursor.fetchall()

# Preserve insertion order for ties using original index
records = [
    {"name": name, "gpa": gpa, "idx": i}
    for i, (name, gpa) in enumerate(rows)
]

# Sort by GPA descending, then by original index to stabilize ties
sorted_records = sorted(records, key=lambda r: (-r["gpa"], r["idx"]))

# Calculate row_number, rank (competition ranking), and dense_rank
previous_gpa = None
current_rank = 0
current_dense_rank = 0
for i, rec in enumerate(sorted_records):
    if previous_gpa is None or rec["gpa"] != previous_gpa:
        current_rank = i + 1            # rank jumps over ties
        current_dense_rank += 1          # dense rank increments by 1 per distinct GPA
    rec["row_number"] = i + 1
    rec["rank"] = current_rank
    rec["dense_rank"] = current_dense_rank
    previous_gpa = rec["gpa"]

print("\nStudent Grades with Ranks:")
print("name\tgpa\trank\tdense_rank\trow_number")
for rec in sorted_records:
    print(f"{rec['name']}\t{rec['gpa']:.1f}\t{rec['rank']}\t{rec['dense_rank']}\t{rec['row_number']}")

conn.close()
