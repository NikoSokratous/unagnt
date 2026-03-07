-- Migration 015: Cost Workflow Analytics (v4)
-- Add workflow_id and workflow_name to cost_entries for usage analytics by workflow

ALTER TABLE cost_entries ADD COLUMN workflow_id TEXT;
ALTER TABLE cost_entries ADD COLUMN workflow_name TEXT;

CREATE INDEX IF NOT EXISTS idx_cost_entries_workflow ON cost_entries(workflow_id);
