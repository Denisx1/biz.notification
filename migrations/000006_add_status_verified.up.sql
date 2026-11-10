ALTER TABLE sms_tasks DROP CONSTRAINT sms_tasks_status_check;

ALTER TABLE sms_tasks
ADD CONSTRAINT sms_tasks_status_check CHECK (
    status IN (
        'pending',
        'sent',
        'failed',
        'expired',
        'verified'
    )
);