import sqlite3

conn = sqlite3.connect('library.db')
cursor = conn.cursor()

cursor.execute('''
    CREATE TABLE IF NOT EXISTS books (
        id INTEGER PRIMARY KEY,
        title TEXT NOT NULL,
        author TEXT NOT NULL,
        year INTEGER NOT NULL
    )
''')

def find_books(author):
    with sqlite3.connect('library.db') as conn:
        cursor = conn.cursor()
        cursor.execute("SELECT title, year FROM books WHERE author = ?", (author,))
        results = cursor.fetchall()

        if len(results) == 0:
            print(f"No books found for author: {author}")
        elif len(results) == 1:
            title, year = results[0]
            print(f"Found one book by {author}: '{title}' ({year})")
        else:
            print(f"Found {len(results)} books by {author}:")
            for title, year in results:
                print(f" - '{title}' ({year})")


