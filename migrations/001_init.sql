-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Workflows table
CREATE TABLE workflows (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  name text NOT NULL,
  version integer NOT NULL DEFAULT 1,
  is_active boolean NOT NULL DEFAULT true,
  input_schema jsonb NOT NULL DEFAULT '{}'::jsonb,
  start_step_id uuid,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- Steps table
CREATE TABLE steps (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  workflow_id uuid NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
  name text NOT NULL,
  step_type text NOT NULL CHECK (step_type IN ('task','approval','notification')),
  "order" integer,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- Rules table
CREATE TABLE rules (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  step_id uuid NOT NULL REFERENCES steps(id) ON DELETE CASCADE,
  condition text NOT NULL,
  next_step_id uuid,
  priority integer NOT NULL DEFAULT 100,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- Executions table
CREATE TABLE executions (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  workflow_id uuid NOT NULL REFERENCES workflows(id),
  workflow_version integer NOT NULL,
  status text NOT NULL CHECK (status IN ('pending','in_progress','completed','failed','canceled')),
  data jsonb NOT NULL DEFAULT '{}'::jsonb,
  current_step_id uuid,
  retries integer NOT NULL DEFAULT 0,
  triggered_by uuid,
  started_at timestamptz,
  ended_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- Execution logs table
CREATE TABLE execution_logs (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  execution_id uuid NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
  step_id uuid,
  step_name text,
  step_type text,
  evaluated_rules jsonb,
  selected_next_step uuid,
  status text NOT NULL CHECK (status IN ('pending','completed','failed')),
  approver_id uuid,
  error_message text,
  started_at timestamptz,
  ended_at timestamptz
);