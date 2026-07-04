DO $$
BEGIN
  ALTER TYPE payment_method ADD VALUE IF NOT EXISTS 'rub';
EXCEPTION
  WHEN undefined_object THEN
    NULL;
END $$;
