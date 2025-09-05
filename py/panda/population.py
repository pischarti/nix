import sqlite3
import pandas as pd
import matplotlib.pyplot as plt

# Connect to a database file (e.g., 'world_population.db')
conn = sqlite3.connect(':memory:')
cursor = conn.cursor()

# Create necessary tables if they do not exist
cursor.executescript('''
    CREATE TABLE IF NOT EXISTS population (
        CountryCode TEXT NOT NULL,
        Year INTEGER NOT NULL,
        PopChange REAL NOT NULL
    );

    CREATE TABLE IF NOT EXISTS country_mapping (
        CountryCode TEXT PRIMARY KEY,
        region TEXT NOT NULL,
        subregion TEXT NOT NULL
    );
''')

# Populate tables with example records
country_mapping_rows = [
    ("USA", "Americas", "Northern America"),
    ("CAN", "Americas", "Northern America"),
    ("DEU", "Europe", "Western Europe"),
    ("FRA", "Europe", "Western Europe"),
    ("JPN", "Asia", "Eastern Asia"),
    ("KOR", "Asia", "Eastern Asia"),
    ("IND", "Asia", "Southern Asia"),
    ("PAK", "Asia", "Southern Asia"),
    ("BRA", "Americas", "South America"),
    ("ARG", "Americas", "South America"),
]

population_rows = [
    ("USA", 2010, 2300000.0), ("USA", 2015, 2700000.0), ("USA", 2020, 1500000.0),
    ("CAN", 2010, 300000.0), ("CAN", 2015, 350000.0), ("CAN", 2020, 400000.0),
    ("DEU", 2010, -200000.0), ("DEU", 2015, 500000.0), ("DEU", 2020, -100000.0),
    ("FRA", 2010, 300000.0), ("FRA", 2015, 250000.0), ("FRA", 2020, 200000.0),
    ("JPN", 2010, -150000.0), ("JPN", 2015, -300000.0), ("JPN", 2020, -400000.0),
    ("KOR", 2010, 50000.0), ("KOR", 2015, 30000.0), ("KOR", 2020, -50000.0),
    ("IND", 2010, 12000000.0), ("IND", 2015, 13000000.0), ("IND", 2020, 14000000.0),
    ("PAK", 2010, 4000000.0), ("PAK", 2015, 4500000.0), ("PAK", 2020, 5000000.0),
    ("BRA", 2010, 1800000.0), ("BRA", 2015, 1600000.0), ("BRA", 2020, 1300000.0),
    ("ARG", 2010, 500000.0), ("ARG", 2015, 450000.0), ("ARG", 2020, 400000.0),
]

cursor.executemany(
    "INSERT OR REPLACE INTO country_mapping (CountryCode, region, subregion) VALUES (?, ?, ?)",
    country_mapping_rows,
)
cursor.executemany(
    "INSERT INTO population (CountryCode, Year, PopChange) VALUES (?, ?, ?)",
    population_rows,
)
conn.commit()

# Use a complex query with JOIN, GROUP BY, and a window function
query = """
SELECT 
    region, 
    subregion, 
    SUM(PopChange) AS TotalPopChange
FROM 
    population p 
JOIN 
    country_mapping c ON p.CountryCode = c.CountryCode
WHERE 
    Year BETWEEN 2010 AND 2020
GROUP BY 
    region, subregion
ORDER BY 
    TotalPopChange DESC
LIMIT 10;
"""

# Execute the query and load the results into a pandas DataFrame
df = pd.read_sql_query(query, conn)

# Close the database connection
conn.close()

# Use matplotlib to create a visualization from the DataFrame
plt.barh(df['subregion'], df['TotalPopChange'])
plt.title('Top 10 Subregions by Population Change (2010-2020)')
plt.xlabel('Population Change')
plt.ylabel('Subregion')
plt.tight_layout()
plt.show()
