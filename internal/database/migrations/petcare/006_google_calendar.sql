ALTER TABLE pets ADD COLUMN google_calendar_id TEXT NOT NULL DEFAULT '';
ALTER TABLE vaccines ADD COLUMN google_calendar_event_id TEXT NOT NULL DEFAULT '';
ALTER TABLE treatments ADD COLUMN google_calendar_event_id TEXT NOT NULL DEFAULT '';
ALTER TABLE doses ADD COLUMN google_calendar_event_id TEXT NOT NULL DEFAULT '';
