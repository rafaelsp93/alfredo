CREATE TABLE IF NOT EXISTS health_profiles (
    id         INTEGER PRIMARY KEY CHECK (id = 1),
    height_cm  REAL    NOT NULL CHECK (height_cm > 0),
    birth_date TEXT    NOT NULL,
    sex        TEXT    NOT NULL CHECK (sex IN ('male', 'female', 'other')),
    created_at TEXT    NOT NULL,
    updated_at TEXT    NOT NULL
);
