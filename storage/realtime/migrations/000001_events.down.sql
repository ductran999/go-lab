DROP TRIGGER IF EXISTS events_notify ON app.events;
DROP FUNCTION IF EXISTS app.notify_event();
DROP TABLE IF EXISTS app.events;
DROP SCHEMA IF EXISTS app;
