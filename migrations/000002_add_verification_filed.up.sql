ALTER TABLE sms_tasks ADD COLUMN verified_at TIMESTAMPTZ;
CREATE INDEX sms_tasks_ip_created_idx ON sms_tasks (ip, created_at);