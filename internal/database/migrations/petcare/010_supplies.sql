CREATE TABLE IF NOT EXISTS supplies (
    id TEXT PRIMARY KEY,
    pet_id TEXT NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    last_purchased_at TEXT NOT NULL,
    estimated_days_supply INTEGER NOT NULL CHECK(estimated_days_supply > 0),
    notes TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_supplies_pet_id ON supplies(pet_id);
CREATE INDEX IF NOT EXISTS idx_supplies_next_reorder ON supplies(pet_id, last_purchased_at, estimated_days_supply);
