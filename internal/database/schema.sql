CREATE TABLE IF NOT EXISTS counter (
	id INTEGER PRIMARY KEY,
    type TEXT NOT NULL UNIQUE,
	count INTEGER NOT NULL
) STRICT;

INSERT OR IGNORE INTO counter (id, type, count) VALUES (1, 'Random', 0);
INSERT OR IGNORE INTO counter (id, type, count) VALUES (2, 'Diceware', 0);
INSERT OR IGNORE INTO counter (id, type, count) VALUES (3, 'PIN', 0);
